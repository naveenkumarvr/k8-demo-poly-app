package com.polyshop.checkout.repository;

import com.polyshop.checkout.entity.Transaction;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;
import java.util.Optional;

/**
 * Transaction Repository
 * 
 * Spring Data JPA repository for Transaction entity.
 * 
 * Spring Data JPA automatically implements:
 * - save() - Insert/update transaction
 * - findById() - Query by primary key
 * - findAll() - Get all transactions
 * - Custom query methods based on method naming convention
 * 
 * OpenTelemetry Instrumentation:
 * - All repository methods are automatically instrumented
 * - Each database operation creates a span with SQL details
 * - Span name format: "SELECT transactions" or "INSERT transactions"
 * - Attributes include: db.statement, db.system=H2, rows affected
 * 
 * No implementation code needed - Spring generates it at runtime!
 */
@Repository
public interface TransactionRepository extends JpaRepository<Transaction, Long> {

    /**
     * Find transaction by unique transaction ID
     * 
     * Spring Data JPA auto-generates implementation:
     * SELECT * FROM transactions WHERE transaction_id = ?
     * 
     * @param transactionId UUID transaction identifier
     * @return Optional containing transaction if found
     */
    Optional<Transaction> findByTransactionId(String transactionId);

    /**
     * Find all transactions for a specific user
     * 
     * Auto-generated query:
     * SELECT * FROM transactions WHERE user_id = ? ORDER BY created_at DESC
     * 
     * @param userId User identifier
     * @return List of transactions for user
     */
    java.util.List<Transaction> findByUserIdOrderByCreatedAtDesc(String userId);
}
