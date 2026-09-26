package com.polyshop.checkout.model;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;
import jakarta.validation.constraints.NotBlank;

/**
 * Checkout Request DTO
 * 
 * Represents the incoming request to process a checkout/payment.
 * 
 * Expected JSON format:
 * {
 * "user_id": "user-123"
 * }
 * 
 * Workflow:
 * 1. Client sends this request to POST /checkout
 * 2. CheckoutService fetches cart from cart-service
 * 3. Payment is simulated (80% success rate)
 * 4. Transaction is saved to H2 database
 * 5. Response includes transaction ID and status
 */
@Data
@NoArgsConstructor
@AllArgsConstructor
public class CheckoutRequest {

    /**
     * User identifier (must match cart in cart-service)
     */
    @NotBlank(message = "user_id is required")
    private String userId;

}
