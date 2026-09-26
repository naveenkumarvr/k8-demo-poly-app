package com.polyshop.checkout.controller;

import com.polyshop.checkout.model.CheckoutRequest;
import com.polyshop.checkout.model.CheckoutResponse;
import com.polyshop.checkout.service.CheckoutService;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.WebMvcTest;
import org.springframework.boot.test.mock.mockito.MockBean;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;

import java.util.List;

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.when;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.*;

/**
 * Integration tests for CheckoutController
 * 
 * Uses Spring Boot Test with MockMvc to test HTTP endpoints
 * 
 * Tests cover:
 * 1. POST /checkout with valid request
 * 2. POST /checkout with invalid request
 * 3. Response status codes
 * 4. JSON serialization/deserialization
 * 
 * @see CheckoutController
 */
@WebMvcTest(CheckoutController.class)
@DisplayName("CheckoutController Integration Tests")
class CheckoutControllerTest {

    @Autowired
    private MockMvc mockMvc;

    @MockBean
    private CheckoutService checkoutService;

    @Test
    @DisplayName("Should return 200 OK for successful checkout")
    void testSuccessfulCheckout() throws Exception {
        // Arrange
        CheckoutResponse mockResponse = CheckoutResponse.builder()
                .transactionId("txn-123")
                .status("SUCCESS")
                .message("Payment processed successfully")
                .userId("user-123")
                .totalItems(2)
                .traceId("trace-abc")
                .build();

        when(checkoutService.processCheckout(any(CheckoutRequest.class)))
                .thenReturn(mockResponse);

        String requestJson = """
                {
                    "userId": "user-123"
                }
                """;

        // Act & Assert
        mockMvc.perform(post("/checkout")
                .contentType(MediaType.APPLICATION_JSON)
                .content(requestJson))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.transaction_id").value("txn-123"))
                .andExpect(jsonPath("$.status").value("SUCCESS"))
                .andExpect(jsonPath("$.user_id").value("user-123"))
                .andExpect(jsonPath("$.total_items").value(2));
    }

    @Test
    @DisplayName("Should return 400 Bad Request when cart not found")
    void testCheckoutWithCartNotFound() throws Exception {
        // Arrange
        CheckoutResponse mockResponse = CheckoutResponse.builder()
                .transactionId("txn-456")
                .status("CART_NOT_FOUND")
                .message("Cart not found in cart-service")
                .userId("user-999")
                .totalItems(0)
                .build();

        when(checkoutService.processCheckout(any(CheckoutRequest.class)))
                .thenReturn(mockResponse);

        String requestJson = """
                {
                    "userId": "user-999"
                }
                """;

        // Act & Assert
        mockMvc.perform(post("/checkout")
                .contentType(MediaType.APPLICATION_JSON)
                .content(requestJson))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.status").value("CART_NOT_FOUND"));
    }

    @Test
    @DisplayName("Should return 400 Bad Request for invalid JSON")
    void testCheckoutWithInvalidJson() throws Exception {
        // Arrange
        String invalidJson = "{ invalid json }";

        // Act & Assert
        mockMvc.perform(post("/checkout")
                .contentType(MediaType.APPLICATION_JSON)
                .content(invalidJson))
                .andExpect(status().isBadRequest());
    }

    @Test
    @DisplayName("Should return 400 Bad Request when user_id is missing")
    void testCheckoutWithMissingUserId() throws Exception {
        // Arrange
        String requestJson = """
                {
                }
                """;

        // Act & Assert
        mockMvc.perform(post("/checkout")
                .contentType(MediaType.APPLICATION_JSON)
                .content(requestJson))
                .andExpect(status().isBadRequest());
    }

    @Test
    @DisplayName("Should return 200 OK for root endpoint")
    void testRootEndpoint() throws Exception {
        // Act & Assert
        mockMvc.perform(get("/"))
                .andExpect(status().isOk())
                .andExpect(content().string(org.hamcrest.Matchers.containsString("Checkout Service is running")));
    }
}
