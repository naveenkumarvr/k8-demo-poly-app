package com.polyshop.checkout.model;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

/**
 * Checkout Response DTO
 * 
 * Represents the response after processing a checkout request.
 * 
 * Example JSON response:
 * {
 * "transaction_id": "txn-7f8e3a1b-9c4d-4e2a-b3f1-5d8c9e7a2b4f",
 * "status": "SUCCESS",
 * "message": "Payment processed successfully",
 * "user_id": "user-123",
 * "total_items": 3,
 * "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736"
 * }
 * 
 * The trace_id is automatically populated by OpenTelemetry for distributed
 * tracing.
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class CheckoutResponse {

    /**
     * Unique transaction identifier (generated UUID)
     */
    private String transactionId;

    /**
     * Transaction status: SUCCESS, FAILED, CART_EMPTY, CART_NOT_FOUND
     */
    private String status;

    /**
     * Human-readable message
     */
    private String message;

    /**
     * User ID from request
     */
    private String userId;

    /**
     * Total number of items in cart
     */
    private int totalItems;

    /**
     * OpenTelemetry trace ID for distributed tracing
     * This allows correlating logs and traces across services
     */
    private String traceId;
}
