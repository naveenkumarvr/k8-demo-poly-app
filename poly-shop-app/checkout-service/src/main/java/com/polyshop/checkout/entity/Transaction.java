package com.polyshop.checkout.entity;

import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import java.time.LocalDateTime;

/**
 * Transaction Entity - JPA persistence for checkout history
 * 
 * This entity stores transaction records in the H2 in-memory database.
 * 
 * Database Schema (auto-created by Hibernate):
 * CREATE TABLE transactions (
 * id BIGINT AUTO_INCREMENT PRIMARY KEY,
 * transaction_id VARCHAR(255) UNIQUE NOT NULL,
 * user_id VARCHAR(255) NOT NULL,
 * status VARCHAR(50) NOT NULL,
 * total_items INT,
 * trace_id VARCHAR(255),
 * created_at TIMESTAMP
 * );
 * 
 * OpenTelemetry Instrumentation:
 * - The OTel Java Agent automatically instruments JPA/Hibernate
 * - Each database query (INSERT, SELECT) creates a span
 * - Span attributes include: db.statement, db.system (H2), query duration
 * 
 * Example trace:
 * POST /checkout
 * ├─ GET http://cart-service:8080/v1/cart/user123 [cart-service span]
 * ├─ INSERT INTO transactions [...] [db span]
 * └─ http.status_code: 200
 */
@Entity
@Table(name = "transactions")
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class Transaction {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    /**
     * Unique transaction identifier (UUID)
     * Indexed for fast lookups
     */
    @Column(unique = true, nullable = false)
    private String transactionId;

    /**
     * User who initiated the checkout
     */
    @Column(nullable = false)
    private String userId;

    /**
     * Transaction status: SUCCESS, FAILED, CART_EMPTY, CART_NOT_FOUND
     */
    @Column(nullable = false)
    private String status;

    /**
     * Number of items in the cart at checkout time
     */
    private int totalItems;

    /**
     * OpenTelemetry trace ID for correlation
     * Allows linking this database record to distributed traces
     */
    private String traceId;

    /**
     * Timestamp when transaction was created
     */
    @Column(nullable = false)
    private LocalDateTime createdAt;

    /**
     * Auto-populate timestamp before persisting
     */
    @PrePersist
    protected void onCreate() {
        createdAt = LocalDateTime.now();
    }
}
