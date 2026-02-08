package com.polyshop.checkout.controller;

import com.polyshop.checkout.service.CurrencyService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import java.math.BigDecimal;
import java.util.HashMap;
import java.util.Map;

/**
 * Currency Conversion Controller
 * 
 * Utility endpoint for converting product prices between currencies.
 * 
 * Example Usage:
 * GET /currency/convert?amount=100&from=USD&to=CAD
 * 
 * Response:
 * {
 * "original_amount": 100.00,
 * "original_currency": "USD",
 * "converted_amount": 135.00,
 * "target_currency": "CAD",
 * "exchange_rate": 1.35
 * }
 */
@RestController
@RequestMapping("/currency")
public class CurrencyController {

    private final CurrencyService currencyService;

    public CurrencyController(CurrencyService currencyService) {
        this.currencyService = currencyService;
    }

    /**
     * Convert currency
     * 
     * Endpoint: GET /currency/convert
     * 
     * Query Parameters:
     * - amount: Amount to convert (required)
     * - from: Source currency code (required) - USD, CAD, EUR, GBP
     * - to: Target currency code (required) - USD, CAD, EUR, GBP
     * 
     * @param amount       Amount to convert
     * @param fromCurrency Source currency
     * @param toCurrency   Target currency
     * @return Conversion result with exchange rate
     */
    @GetMapping("/convert")
    public ResponseEntity<Map<String, Object>> convert(
            @RequestParam BigDecimal amount,
            @RequestParam String from,
            @RequestParam String to) {
        try {
            BigDecimal convertedAmount = currencyService.convert(amount, from, to);

            // Calculate exchange rate
            BigDecimal rate = convertedAmount.divide(amount, 4, BigDecimal.ROUND_HALF_UP);

            Map<String, Object> response = new HashMap<>();
            response.put("original_amount", amount);
            response.put("original_currency", from.toUpperCase());
            response.put("converted_amount", convertedAmount);
            response.put("target_currency", to.toUpperCase());
            response.put("exchange_rate", rate);

            return ResponseEntity.ok(response);

        } catch (IllegalArgumentException e) {
            Map<String, Object> error = new HashMap<>();
            error.put("error", e.getMessage());
            error.put("supported_currencies", new String[] { "USD", "CAD", "EUR", "GBP" });
            return ResponseEntity.badRequest().body(error);
        }
    }

    /**
     * Get all exchange rates
     * 
     * Endpoint: GET /currency/rates
     * 
     * Returns all supported exchange rates
     */
    @GetMapping("/rates")
    public ResponseEntity<Map<String, Double>> getRates() {
        return ResponseEntity.ok(currencyService.getAllRates());
    }
}
