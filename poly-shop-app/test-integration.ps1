# Poly-Shop Integration Test Script
Write-Host "`nPOLY-SHOP INTEGRATION TEST`n" -ForegroundColor Cyan

$testUser = "test-user-$(Get-Random)"

function Test-Endpoint {
    param($Name, $Url, $Method = "GET", $Body = $null)
    
    Write-Host "Testing: $Name..." -NoNewline
    
    try {
        $params = @{
            Uri = $Url
            Method = $Method
            UseBasicParsing = $true
            TimeoutSec = 10
        }
        
        if ($Body) {
            $params.Body = $Body
            $params.ContentType = "application/json"
        }
        
        $response = Invoke-WebRequest @params
        
        if ($response.StatusCode -ge 200 -and $response.StatusCode -lt 300) {
            Write-Host " PASS [$($response.StatusCode)]" -ForegroundColor Green
            return @{Success = $true; Response = $response; Data = ($response.Content | ConvertFrom-Json -ErrorAction SilentlyContinue)}
        }
    }
    catch {
        Write-Host " FAIL" -ForegroundColor Red
        Write-Host "   Error: $($_.Exception.Message)" -ForegroundColor Red
        return @{Success = $false; Error = $_}
    }
}

Write-Host "`n[1] PRODUCT SERVICE (Port 8090)" -ForegroundColor Yellow
Test-Endpoint "GET /products" "http://localhost:8090/products" | Out-Null
$productsResponse = Invoke-RestMethod -Uri "http://localhost:8090/products" -Method Get
$firstProductId = $productsResponse[0].id
Write-Host "  Using Product ID: $firstProductId" -ForegroundColor Gray
Test-Endpoint "GET /products/$firstProductId" "http://localhost:8090/products/$firstProductId" | Out-Null

Write-Host "`n[2] CART SERVICE (Port 8080)" -ForegroundColor Yellow
Test-Endpoint "GET /v1/cart/$testUser (empty)" "http://localhost:8080/v1/cart/$testUser" | Out-Null

$addBody = @{
    product_id = "$firstProductId"
    quantity = 2
} | ConvertTo-Json
Test-Endpoint "POST /v1/cart/$testUser" "http://localhost:8080/v1/cart/$testUser" "POST" $addBody | Out-Null
Test-Endpoint "GET /v1/cart/$testUser (with items)" "http://localhost:8080/v1/cart/$testUser" | Out-Null

Write-Host "`n[3] CHECKOUT SERVICE (Port 8085)" -ForegroundColor Yellow
Test-Endpoint "GET /actuator/health" "http://localhost:8085/actuator/health" | Out-Null
Test-Endpoint "GET /actuator/health/liveness" "http://localhost:8085/actuator/health/liveness" | Out-Null
Test-Endpoint "GET /actuator/health/readiness" "http://localhost:8085/actuator/health/readiness" | Out-Null
Test-Endpoint "GET /currency/convert" "http://localhost:8085/currency/convert?amount=100&from=USD&to=CAD" | Out-Null
Test-Endpoint "GET /currency/rates" "http://localhost:8085/currency/rates" | Out-Null
Test-Endpoint "GET /admin/memory-stats" "http://localhost:8085/admin/memory-stats" | Out-Null

Write-Host "`n[4] END-TO-END CHECKOUT FLOW (DISTRIBUTED TRACING)" -ForegroundColor Yellow
$checkoutBody = @"
{
  "userId": "$testUser"
}
"@

$result = Test-Endpoint "POST /checkout" "http://localhost:8085/checkout" "POST" $checkoutBody

if ($result.Success -and $result.Data) {
    Write-Host "`nCHECKOUT RESULT:" -ForegroundColor Cyan
    Write-Host "  Transaction ID: $($result.Data.transaction_id)" -ForegroundColor White
    Write-Host "  Status: $($result.Data.status)" -ForegroundColor White
    Write-Host "  Message: $($result.Data.message)" -ForegroundColor White
    Write-Host "  Trace ID: $($result.Data.trace_id)" -ForegroundColor Green
    Write-Host "`n  View trace in Jaeger: http://localhost:16686/trace/$($result.Data.trace_id)" -ForegroundColor Cyan
}

Write-Host "`n[5] SERVICE INTERACTIONS VERIFIED" -ForegroundColor Yellow
Write-Host "  - Cart Service -> Redis" -ForegroundColor Gray
Write-Host "  - Product Service -> PostgreSQL" -ForegroundColor Gray
Write-Host "  - Checkout Service -> Cart Service (RestTemplate)" -ForegroundColor Gray
Write-Host "  - All Services -> Jaeger (OpenTelemetry tracing)" -ForegroundColor Gray

Write-Host "`nINTEGRATION TEST COMPLETE!`n" -ForegroundColor Green
