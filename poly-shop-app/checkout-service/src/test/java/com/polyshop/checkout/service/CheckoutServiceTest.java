package com.polyshop.checkout.service;

import com.polyshop.checkout.entity.Transaction;
import com.polyshop.checkout.model.CheckoutRequest;
import com.polyshop.checkout.model.CheckoutResponse;
import com.polyshop.checkout.repository.TransactionRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.test.util.ReflectionTestUtils;
import org.springframework.web.client.RestTemplate;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

/**
 * Unit tests for CheckoutService
 * 
 * Tests cover:
 * 1. Successful checkout flow
 * 2. Cart not found scenario
 * 3. Empty cart scenario
 * 4. Payment processing simulation
 * 5. RestTemplate integration (mocked)
 * 
 * @see CheckoutService
 */
@ExtendWith(MockitoExtension.class)
@DisplayName("CheckoutService Unit Tests")
class CheckoutServiceTest {

    @Mock
    private RestTemplate restTemplate;

    @Mock
    private TransactionRepository transactionRepository;

    @InjectMocks
    private CheckoutService checkoutService;

    private CheckoutRequest validRequest;

    @BeforeEach
    void setUp() {
        // Set configuration values using ReflectionTestUtils
        ReflectionTestUtils.setField(checkoutService, "cartServiceUrl", "http://cart-service:8080");
        ReflectionTestUtils.setField(checkoutService, "paymentSuccessRate", 1.0); // 100% success for tests
        ReflectionTestUtils.setField(checkoutService, "processingTimeMs", 0); // No delay in tests

        // Create valid checkout request
        validRequest = new CheckoutRequest();
        validRequest.setUserId("user-test-123");
    }

    @Test
    @DisplayName("Should process checkout successfully when cart exists and payment succeeds")
    void testSuccessfulCheckout() {
        // Arrange
        Map<String, Object> mockCartData = new HashMap<>();
        mockCartData.put("user_id", "user-test-123");
        mockCartData.put("items", List.of(Map.of("product_id", "prod-1", "quantity", 2)));

        when(restTemplate.getForObject(anyString(), eq(Map.class)))
                .thenReturn(mockCartData);

        when(transactionRepository.save(any(Transaction.class)))
                .thenAnswer(invocation -> invocation.getArgument(0));

        // Act
        CheckoutResponse response = checkoutService.processCheckout(validRequest);

        // Assert
        assertNotNull(response);
        assertEquals("SUCCESS", response.getStatus());
        assertEquals("user-test-123", response.getUserId());
        assertEquals(1, response.getTotalItems());
        assertNotNull(response.getTransactionId());
        assertTrue(response.getMessage().contains("successfully"));

        // Verify interactions
        verify(restTemplate, times(1)).getForObject(anyString(), eq(Map.class));
        verify(transactionRepository, times(1)).save(any(Transaction.class));
    }

    @Test
    @DisplayName("Should return CART_NOT_FOUND when cart service returns null")
    void testCheckoutWithCartNotFound() {
        // Arrange
        when(restTemplate.getForObject(anyString(), eq(Map.class)))
            .thenReturn(null);
        
        when(transactionRepository.save(any(Transaction.class)))
            .thenAnswer(invocation -> invocation.getArgument(0));
        
        // Act
        CheckoutResponse response = checkoutService.processCheckout(validRequest);
        
        // Assert
        assertNotNull(response);
        assertEquals("CART_NOT_FOUND", response.getStatus());
        assertEquals("user-test-123", response.getUserId());
        assertEquals(0, response.getTotalItems());
        assertTrue(response.getMessage().contains("not found"));
        
        verify(restTemplate, times(1)).getForObject(anyString(), eq(Map.class));
        verify(transactionRepository, times(1)).save(any(Transaction.class));
    }

    @Test
    @DisplayName("Should return CART_EMPTY when cart has no items")
    void testCheckoutWithEmptyCart() {
        // Arrange
        Map<String, Object> mockCartData = new HashMap<>();
        mockCartData.put("user_id", "user-test-123");
        mockCartData.put("items", List.of());

        when(restTemplate.getForObject(anyString(), eq(Map.class)))
                .thenReturn(mockCartData);

        when(transactionRepository.save(any(Transaction.class)))
                .thenAnswer(invocation -> invocation.getArgument(0));

        // Act
        CheckoutResponse response = checkoutService.processCheckout(validRequest);

        // Assert
        assertNotNull(response);
        assertEquals("CART_EMPTY", response.getStatus());
        assertEquals(0, response.getTotalItems());
        assertTrue(response.getMessage().contains("empty"));

        verify(transactionRepository, times(1)).save(any(Transaction.class));
    }

    @Test
    @DisplayName("Should handle cart service exception gracefully")
    void testCheckoutWithCartServiceException() {
        // Arrange
        when(restTemplate.getForObject(anyString(), eq(Map.class)))
            .thenThrow(new RuntimeException("Connection refused"));
        
        when(transactionRepository.save(any(Transaction.class)))
            .thenAnswer(invocation -> invocation.getArgument(0));
        
        // Act
        CheckoutResponse response = checkoutService.processCheckout(validRequest);
        
        // Assert
        assertNotNull(response);
        assertEquals("CART_NOT_FOUND", response.getStatus());
        
        verify(restTemplate, times(1)).getForObject(anyString(), eq(Map.class));
        verify(transactionRepository, times(1)).save(any(Transaction.class));
    }

    @Test
    @DisplayName("Should save transaction to repository")
    void testTransactionPersistence() {
        // Arrange
        Map<String, Object> mockCartData = new HashMap<>();
        mockCartData.put("user_id", "user-test-123");
        mockCartData.put("items", List.of(Map.of("product_id", "prod-1", "quantity", 1)));

        when(restTemplate.getForObject(anyString(), eq(Map.class)))
                .thenReturn(mockCartData);

        when(transactionRepository.save(any(Transaction.class)))
                .thenAnswer(invocation -> {
                    Transaction saved = invocation.getArgument(0);
                    assertNotNull(saved.getTransactionId());
                    assertEquals("user-test-123", saved.getUserId());
                    assertEquals("SUCCESS", saved.getStatus());
                    return saved;
                });

        // Act
        checkoutService.processCheckout(validRequest);

        // Assert
        verify(transactionRepository, times(1))
                .save(argThat(transaction -> transaction.getUserId().equals("user-test-123") &&
                        transaction.getStatus().equals("SUCCESS")));
    }
}
