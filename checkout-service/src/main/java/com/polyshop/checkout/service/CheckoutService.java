package com.polyshop.checkout.service;

import com.polyshop.checkout.entity.Transaction;
import com.polyshop.checkout.model.CheckoutRequest;
import com.polyshop.checkout.model.CheckoutResponse;
import com.polyshop.checkout.repository.TransactionRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestTemplate;
import java.util.Map;
import java.util.Random;
import java.util.UUID;

/**
 * Checkout Service - Core business logic for payment processing
 * 
 * This service demonstrates:
 * 1. Inter-service communication (calling cart-service via RestTemplate)
 * 2. OpenTelemetry trace propagation (automatic via Java Agent)
 * 3. Transaction persistence (JPA with H2)
 * 4. Simulated payment processing
 * 
 * Distributed Tracing Flow:
 * POST /checkout → CheckoutService.processCheckout()
 * ├─ restTemplate.getForObject() → cart-service [HTTP span created by OTel
 * agent]
 * │ └─ Trace context headers injected: traceparent, tracestate
 * ├─ simulatePaymentProcessing() [method span if annotated]
 * ├─ transactionRepository.save() → H2 INSERT [DB span created by OTel agent]
 * └─ Return CheckoutResponse with trace_id
 */
@Service
public class CheckoutService {

    private static final Logger logger = LoggerFactory.getLogger(CheckoutService.class);
    private static final Random random = new Random();

    private final RestTemplate restTemplate;
    private final TransactionRepository transactionRepository;

    @Value("${checkout.cart-service-url}")
    private String cartServiceUrl;

    @Value("${checkout.payment.success-rate:0.8}")
    private double paymentSuccessRate;

    @Value("${checkout.payment.processing-time-ms:200}")
    private int processingTimeMs;

    public CheckoutService(RestTemplate restTemplate, TransactionRepository transactionRepository) {
        this.restTemplate = restTemplate;
        this.transactionRepository = transactionRepository;
    }

    /**
     * Process checkout for a user's cart
     * 
     * Workflow:
     * 1. Verify cart exists by calling cart-service
     * 2. Simulate payment processing (80% success rate)
     * 3. Save transaction to database
     * 4. Return response with transaction ID
     * 
     * @param request Checkout request with user_id
     * @return CheckoutResponse with transaction status
     */
    public CheckoutResponse processCheckout(CheckoutRequest request) {
        logger.info("Processing checkout for user: {}", request.getUserId());

        String transactionId = UUID.randomUUID().toString();
        String traceId = getCurrentTraceId();

        try {
            // Step 1: Verify cart with cart-service
            // The OpenTelemetry agent will:
            // - Create a new span for this HTTP request
            // - Inject trace context headers (traceparent, tracestate)
            // - Link this span to the current checkout request span
            Map<String, Object> cartData = verifyCart(request.getUserId());

            if (cartData == null || cartData.isEmpty()) {
                return saveTransaction(transactionId, request.getUserId(), 0, "CART_NOT_FOUND",
                        "Cart not found in cart-service", traceId);
            }

            // Step 2: Check if cart has items
            // Refactored: Fetch items from cart-service response instead of request input
            int totalItems = 0;
            if (cartData.containsKey("items") && cartData.get("items") instanceof java.util.List) {
                java.util.List<?> items = (java.util.List<?>) cartData.get("items");
                totalItems = items.size();
            }

            if (totalItems == 0) {
                return saveTransaction(transactionId, request.getUserId(), 0, "CART_EMPTY",
                        "Cart is empty, cannot process payment", traceId);
            }

            // Step 3: Simulate payment processing
            boolean paymentSuccess = simulatePaymentProcessing();

            if (paymentSuccess) {
                logger.info("✅ Payment successful for user: {} (txn: {})", request.getUserId(), transactionId);
                return saveTransaction(transactionId, request.getUserId(), totalItems, "SUCCESS",
                        "Payment processed successfully", traceId);
            } else {
                logger.warn("❌ Payment failed for user: {} (txn: {})", request.getUserId(), transactionId);
                return saveTransaction(transactionId, request.getUserId(), totalItems, "FAILED",
                        "Payment processing failed, please try again", traceId);
            }

        } catch (Exception e) {
            logger.error("Error processing checkout for user: {}", request.getUserId(), e);
            return saveTransaction(transactionId, request.getUserId(), 0, "ERROR",
                    "Internal error: " + e.getMessage(), traceId);
        }
    }

    /**
     * Verify cart exists by calling cart-service
     * 
     * This method makes an HTTP GET request to:
     * http://cart-service:8080/v1/cart/{userId}
     * 
     * OpenTelemetry Automatic Instrumentation:
     * - The Java Agent intercepts RestTemplate.getForObject()
     * - Creates a span: "GET /v1/cart/{userId}"
     * - Injects W3C Trace Context headers:
     * traceparent: 00-{trace-id}-{span-id}-01
     * tracestate: ...
     * - cart-service receives these headers and continues the trace
     * 
     * Result: Both services share the same trace ID in Jaeger!
     * 
     * @param userId User identifier
     * @return Cart data from cart-service
     */
    @SuppressWarnings("unchecked")
    private Map<String, Object> verifyCart(String userId) {
        try {
            String cartUrl = cartServiceUrl + "/v1/cart/" + userId;
            logger.debug("Calling cart-service at: {}", cartUrl);

            // RestTemplate is auto-instrumented by OpenTelemetry Java Agent
            // No manual trace context propagation needed!
            return restTemplate.getForObject(cartUrl, Map.class);

        } catch (Exception e) {
            logger.error("Failed to verify cart for user: {}", userId, e);
            return null;
        }
    }

    /**
     * Simulate payment processing
     * 
     * In a real application, this would:
     * - Call payment gateway (Stripe, PayPal, etc.)
     * - Validate credit card
     * - Process transaction
     * 
     * For demo purposes, we simulate:
     * - Processing delay (200ms default)
     * - Random success/failure (80% success rate)
     * 
     * @return true if payment succeeds, false otherwise
     */
    private boolean simulatePaymentProcessing() {
        try {
            // Simulate processing time
            Thread.sleep(processingTimeMs);

            // Random success based on configured rate (default 80%)
            return random.nextDouble() < paymentSuccessRate;

        } catch (InterruptedException e) {
            logger.error("Payment processing interrupted", e);
            Thread.currentThread().interrupt();
            return false;
        }
    }

    /**
     * Save transaction to database and build response
     */
    private CheckoutResponse saveTransaction(String transactionId, String userId,
            int totalItems, String status, String message, String traceId) {
        // Create transaction entity
        Transaction transaction = Transaction.builder()
                .transactionId(transactionId)
                .userId(userId)
                .status(status)
                .totalItems(totalItems)
                .traceId(traceId)
                .build();

        // Save to H2 database (auto-instrumented by OTel agent)
        transactionRepository.save(transaction);

        // Build response
        return CheckoutResponse.builder()
                .transactionId(transactionId)
                .status(status)
                .message(message)
                .userId(userId)
                .totalItems(totalItems)
                .traceId(traceId)
                .build();
    }

    /**
     * Get current trace ID from OpenTelemetry context
     * 
     * In a real implementation, you would extract this from:
     * io.opentelemetry.api.trace.Span.current().getSpanContext().getTraceId()
     * 
     * For simplicity, we return a placeholder
     */
    private String getCurrentTraceId() {
        // TODO: Extract actual trace ID from OTel context
        // This requires adding opentelemetry-api dependency
        return "trace-" + UUID.randomUUID().toString().substring(0, 8);
    }
}
