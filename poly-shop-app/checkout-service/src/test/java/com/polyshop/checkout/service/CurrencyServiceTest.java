package com.polyshop.checkout.service;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.CsvSource;

import java.math.BigDecimal;
import java.util.Map;

import static org.junit.jupiter.api.Assertions.*;

/**
 * Unit tests for CurrencyService
 * 
 * Tests cover:
 * 1. Currency conversion calculations
 * 2. Invalid input handling
 * 3. Same currency conversion
 * 4. Exchange rate retrieval
 * 
 * @see CurrencyService
 */
@DisplayName("CurrencyService Unit Tests")
class CurrencyServiceTest {

    private final CurrencyService currencyService = new CurrencyService(
            1.35, // USD_TO_CAD
            0.92, // USD_TO_EUR
            0.79, // USD_TO_GBP
            0.74, // CAD_TO_USD
            1.09, // EUR_TO_USD
            1.27 // GBP_TO_USD
    );

    @ParameterizedTest
    @CsvSource({
            "100, USD, CAD, 135.00",
            "100, USD, EUR, 92.00",
            "100, USD, GBP, 79.00",
            "100, CAD, USD, 74.00",
            "100, EUR, USD, 109.00",
            "100, GBP, USD, 127.00"
    })
    @DisplayName("Should convert currency correctly")
    void testCurrencyConversion(String amount, String from, String to, String expected) {
        // Arrange
        BigDecimal amountBd = new BigDecimal(amount);
        BigDecimal expectedBd = new BigDecimal(expected);

        // Act
        BigDecimal result = currencyService.convert(amountBd, from, to);

        // Assert
        assertEquals(expectedBd, result);
    }

    @Test
    @DisplayName("Should return same amount when converting to same currency")
    void testSameCurrencyConversion() {
        // Arrange
        BigDecimal amount = new BigDecimal("123.45");

        // Act
        BigDecimal result = currencyService.convert(amount, "USD", "USD");

        // Assert
        assertEquals(new BigDecimal("123.45"), result);
    }

    @Test
    @DisplayName("Should handle case-insensitive currency codes")
    void testCaseInsensitiveCurrencies() {
        // Arrange
        BigDecimal amount = new BigDecimal("100");

        // Act
        BigDecimal result = currencyService.convert(amount, "usd", "cad");

        // Assert
        assertEquals(new BigDecimal("135.00"), result);
    }

    @Test
    @DisplayName("Should throw exception for negative amount")
    void testNegativeAmount() {
        // Arrange
        BigDecimal amount = new BigDecimal("-100");

        // Act & Assert
        assertThrows(IllegalArgumentException.class, () -> currencyService.convert(amount, "USD", "CAD"));
    }

    @Test
    @DisplayName("Should throw exception for null amount")
    void testNullAmount() {
        // Act & Assert
        assertThrows(IllegalArgumentException.class, () -> currencyService.convert(null, "USD", "CAD"));
    }

    @Test
    @DisplayName("Should throw exception for unsupported currency")
    void testUnsupportedCurrency() {
        // Arrange
        BigDecimal amount = new BigDecimal("100");

        // Act & Assert
        assertThrows(IllegalArgumentException.class, () -> currencyService.convert(amount, "USD", "JPY"));
    }

    @Test
    @DisplayName("Should return all exchange rates")
    void testGetAllRates() {
        // Act
        Map<String, Double> rates = currencyService.getAllRates();

        // Assert
        assertNotNull(rates);
        assertTrue(rates.size() > 6); // At least 6 base rates + derived rates
        assertTrue(rates.containsKey("USD_TO_CAD"));
        assertTrue(rates.containsKey("CAD_TO_USD"));
        assertEquals(1.35, rates.get("USD_TO_CAD"));
    }

    @Test
    @DisplayName("Should round result to 2 decimal places")
    void testRounding() {
        // Arrange
        BigDecimal amount = new BigDecimal("100.123");

        // Act
        BigDecimal result = currencyService.convert(amount, "USD", "CAD");

        // Assert
        assertEquals(2, result.scale()); // 2 decimal places
        assertEquals(new BigDecimal("135.17"), result); // 100.123 * 1.35 = 135.16605 → 135.17
    }
}
