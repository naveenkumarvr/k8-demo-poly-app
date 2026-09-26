package com.polyshop.checkout.service;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;
import java.math.BigDecimal;
import java.math.RoundingMode;
import java.util.HashMap;
import java.util.Map;

/**
 * Currency Conversion Service
 * 
 * Provides currency conversion functionality using mock exchange rates.
 * In production, this would call a real currency API (e.g.,
 * exchangerate-api.com)
 * 
 * Supported Currencies: USD, CAD, EUR, GBP
 */
@Service
public class CurrencyService {

    // Exchange rates loaded from application.yml
    private final Map<String, Double> exchangeRates = new HashMap<>();

    public CurrencyService(
            @Value("${currency.rates.USD_TO_CAD:1.35}") double usdToCad,
            @Value("${currency.rates.USD_TO_EUR:0.92}") double usdToEur,
            @Value("${currency.rates.USD_TO_GBP:0.79}") double usdToGbp,
            @Value("${currency.rates.CAD_TO_USD:0.74}") double cadToUsd,
            @Value("${currency.rates.EUR_TO_USD:1.09}") double eurToUsd,
            @Value("${currency.rates.GBP_TO_USD:1.27}") double gbpToUsd) {
        // Build exchange rate map
        exchangeRates.put("USD_TO_CAD", usdToCad);
        exchangeRates.put("USD_TO_EUR", usdToEur);
        exchangeRates.put("USD_TO_GBP", usdToGbp);
        exchangeRates.put("CAD_TO_USD", cadToUsd);
        exchangeRates.put("EUR_TO_USD", eurToUsd);
        exchangeRates.put("GBP_TO_USD", gbpToUsd);

        // Derived rates (calculated from USD rates)
        exchangeRates.put("CAD_TO_EUR", usdToEur / usdToCad);
        exchangeRates.put("CAD_TO_GBP", usdToGbp / usdToCad);
        exchangeRates.put("EUR_TO_CAD", usdToCad / usdToEur);
        exchangeRates.put("EUR_TO_GBP", usdToGbp / usdToEur);
        exchangeRates.put("GBP_TO_CAD", usdToCad / usdToGbp);
        exchangeRates.put("GBP_TO_EUR", usdToEur / usdToGbp);
    }

    /**
     * Convert amount from one currency to another
     * 
     * @param amount       Amount to convert
     * @param fromCurrency Source currency (USD, CAD, EUR, GBP)
     * @param toCurrency   Target currency (USD, CAD, EUR, GBP)
     * @return Converted amount rounded to 2 decimal places
     * @throws IllegalArgumentException if currencies are invalid or same
     */
    public BigDecimal convert(BigDecimal amount, String fromCurrency, String toCurrency) {
        if (amount == null || amount.compareTo(BigDecimal.ZERO) < 0) {
            throw new IllegalArgumentException("Amount must be positive");
        }

        fromCurrency = fromCurrency.toUpperCase();
        toCurrency = toCurrency.toUpperCase();

        // Same currency, no conversion needed
        if (fromCurrency.equals(toCurrency)) {
            return amount.setScale(2, RoundingMode.HALF_UP);
        }

        // Get exchange rate
        String rateKey = fromCurrency + "_TO_" + toCurrency;
        Double rate = exchangeRates.get(rateKey);

        if (rate == null) {
            throw new IllegalArgumentException(
                    "Unsupported currency conversion: " + fromCurrency + " to " + toCurrency);
        }

        // Convert and round to 2 decimal places
        return amount.multiply(BigDecimal.valueOf(rate))
                .setScale(2, RoundingMode.HALF_UP);
    }

    /**
     * Get all supported exchange rates
     */
    public Map<String, Double> getAllRates() {
        return new HashMap<>(exchangeRates);
    }
}
