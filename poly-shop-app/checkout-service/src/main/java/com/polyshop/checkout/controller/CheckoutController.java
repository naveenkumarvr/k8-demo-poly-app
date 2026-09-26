package com.polyshop.checkout.controller;

import com.polyshop.checkout.model.CheckoutRequest;
import com.polyshop.checkout.model.CheckoutResponse;
import com.polyshop.checkout.service.CheckoutService;
import jakarta.validation.Valid;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

/**
 * Checkout REST Controller
 * 
 * Exposes the checkout endpoint for payment processing.
 * 
 * OpenTelemetry Auto-Instrumentation:
 * - The Java Agent automatically instruments @RestController classes
 * - Each HTTP request creates a root span
 * - Span name: "POST /checkout"
 * - Span attributes include:
 * * http.method = POST
 * * http.url = /checkout
 * * http.status_code = 200/400/500
 * * user_id (custom attribute)
 * 
 * @see CheckoutService for business logic
 */
@RestController
@RequestMapping
public class CheckoutController {

    private static final Logger logger = LoggerFactory.getLogger(CheckoutController.class);

    private final CheckoutService checkoutService;

    public CheckoutController(CheckoutService checkoutService) {
        this.checkoutService = checkoutService;
    }

    /**
     * Process checkout and payment
     * 
     * Endpoint: POST /checkout
     * 
     * Request Body:
     * {
     * "user_id": "user-123",
     * "cart_items": [
     * {"product_id": "1", "quantity": 2}
     * ]
     * }
     * 
     * Response (Success):
     * {
     * "transaction_id": "txn-7f8e3a1b-9c4d-4e2a-b3f1-5d8c9e7a2b4f",
     * "status": "SUCCESS",
     * "message": "Payment processed successfully",
     * "user_id": "user-123",
     * "total_items": 1,
     * "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736"
     * }
     * 
     * Distributed Tracing Flow:
     * 1. OTel agent creates root span for incoming HTTP request
     * 2. CheckoutService.processCheckout() executes
     * 3. RestTemplate call to cart-service creates child span
     * 4. JPA save() creates database span
     * 5. All spans linked by same trace_id
     * 
     * @param request Checkout request with user_id and cart_items
     * @return CheckoutResponse with transaction details
     */
    @PostMapping("/checkout")
    public ResponseEntity<CheckoutResponse> processCheckout(@Valid @RequestBody CheckoutRequest request) {
        logger.info("📦 Received checkout request for user: {}", request.getUserId());

        try {
            CheckoutResponse response = checkoutService.processCheckout(request);

            // Return appropriate status code
            if ("SUCCESS".equals(response.getStatus())) {
                return ResponseEntity.ok(response);
            } else if ("CART_NOT_FOUND".equals(response.getStatus()) || "CART_EMPTY".equals(response.getStatus())) {
                return ResponseEntity.status(HttpStatus.BAD_REQUEST).body(response);
            } else {
                return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body(response);
            }

        } catch (Exception e) {
            logger.error("Error processing checkout", e);

            CheckoutResponse errorResponse = CheckoutResponse.builder()
                    .transactionId("error")
                    .status("ERROR")
                    .message("Internal server error: " + e.getMessage())
                    .userId(request.getUserId())
                    .totalItems(0)
                    .build();

            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body(errorResponse);
        }
    }

    /**
     * Health check endpoint
     * 
     * Simple endpoint to verify service is running.
     * For production, use Spring Actuator health endpoints instead.
     */
    @GetMapping("/")
    public ResponseEntity<String> root() {
        return ResponseEntity.ok("Checkout Service is running! POST /checkout to process payments.");
    }
}
