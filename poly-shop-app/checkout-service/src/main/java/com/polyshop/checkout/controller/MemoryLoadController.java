package com.polyshop.checkout.controller;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import java.util.HashMap;
import java.util.Map;

/**
 * Memory Load Testing Controller
 * 
 * This controller provides endpoints to simulate high memory usage for testing:
 * 1. Kubernetes resource limits (memory.limits)
 * 2. JVM heap exhaustion
 * 3. OOMKill scenarios
 * 
 * ⚠️ WARNING: These endpoints should be DISABLED in production!
 * 
 * Use Cases:
 * - Testing K8s resource limits enforcement
 * - Verifying OOMKill behavior
 * - Testing horizontal pod autoscaling (HPA) based on memory metrics
 * - Validating JVM heap settings (-Xms/-Xmx)
 * 
 * Example:
 * POST /admin/memory-load?sizeMB=700
 * 
 * If JVM max heap is 768MB, this will cause high memory pressure.
 * If heap is exceeded → OutOfMemoryError → Container restart
 */
@RestController
@RequestMapping("/admin")
public class MemoryLoadController {

    private static final Logger logger = LoggerFactory.getLogger(MemoryLoadController.class);

    // Store allocated arrays to prevent garbage collection
    private static byte[][] allocatedMemory = new byte[10][];
    private static int allocationIndex = 0;

    /**
     * Allocate memory to simulate high memory usage
     * 
     * Endpoint: POST /admin/memory-load
     * 
     * Query Parameters:
     * - sizeMB: Size in megabytes to allocate (required)
     * 
     * This endpoint allocates a byte array of the specified size.
     * The array is stored in a static variable to prevent GC from freeing it.
     * 
     * Testing Scenarios:
     * 
     * 1. Normal Operation (within limits):
     * POST /admin/memory-load?sizeMB=100
     * → Memory usage increases but stays within heap
     * 
     * 2. Near Limit (test pressure):
     * POST /admin/memory-load?sizeMB=600
     * → If max heap = 768MB, this triggers frequent GC
     * 
     * 3. Exceed Heap (trigger OOMKill):
     * POST /admin/memory-load?sizeMB=800
     * → Exceeds max heap → OutOfMemoryError → K8s restarts container
     * 
     * Monitor with:
     * - docker stats (Docker)
     * - kubectl top pod (Kubernetes)
     * - JConsole / VisualVM (JVM monitoring)
     * 
     * @param sizeMB Size in MB to allocate
     * @return Response with memory allocation details
     */
    @PostMapping("/memory-load")
    public ResponseEntity<Map<String, Object>> allocateMemory(@RequestParam int sizeMB) {
        logger.warn("⚠️ Memory allocation requested: {} MB", sizeMB);

        if (sizeMB <= 0 || sizeMB > 2048) {
            Map<String, Object> error = new HashMap<>();
            error.put("error", "Invalid size. Must be between 1 and 2048 MB");
            return ResponseEntity.badRequest().body(error);
        }

        try {
            // Get JVM memory info before allocation
            Runtime runtime = Runtime.getRuntime();
            long beforeUsed = runtime.totalMemory() - runtime.freeMemory();

            // Allocate byte array (1 MB = 1024 * 1024 bytes)
            int bytes = sizeMB * 1024 * 1024;
            byte[] allocation = new byte[bytes];

            // Fill with random data to ensure actual allocation (prevent JVM optimization)
            for (int i = 0; i < allocation.length; i += 1024) {
                allocation[i] = (byte) (i % 256);
            }

            // Store in static array to prevent GC
            allocatedMemory[allocationIndex % allocatedMemory.length] = allocation;
            allocationIndex++;

            // Get memory info after allocation
            long afterUsed = runtime.totalMemory() - runtime.freeMemory();
            long maxMemory = runtime.maxMemory();

            logger.warn("✅ Allocated {} MB. Before: {} MB, After: {} MB, Max: {} MB",
                    sizeMB,
                    beforeUsed / (1024 * 1024),
                    afterUsed / (1024 * 1024),
                    maxMemory / (1024 * 1024));

            Map<String, Object> response = new HashMap<>();
            response.put("allocated_mb", sizeMB);
            response.put("used_before_mb", beforeUsed / (1024 * 1024));
            response.put("used_after_mb", afterUsed / (1024 * 1024));
            response.put("max_heap_mb", maxMemory / (1024 * 1024));
            response.put("usage_percent", (afterUsed * 100.0) / maxMemory);
            response.put("warning", "Memory allocation successful. Monitor container for OOMKill!");

            return ResponseEntity.ok(response);

        } catch (OutOfMemoryError e) {
            logger.error("❌ OutOfMemoryError triggered!", e);

            Map<String, Object> error = new HashMap<>();
            error.put("error", "OutOfMemoryError: JVM heap exhausted");
            error.put("requested_mb", sizeMB);
            error.put("message", "Heap size exceeded. Container may be killed by K8s.");

            return ResponseEntity.status(500).body(error);
        }
    }

    /**
     * Clear allocated memory
     * 
     * Endpoint: POST /admin/memory-clear
     * 
     * Clears the allocated memory and triggers garbage collection
     */
    @PostMapping("/memory-clear")
    public ResponseEntity<Map<String, Object>> clearMemory() {
        logger.info("Clearing allocated memory");

        // Clear static array
        for (int i = 0; i < allocatedMemory.length; i++) {
            allocatedMemory[i] = null;
        }
        allocationIndex = 0;

        // Suggest garbage collection (not guaranteed)
        System.gc();

        Runtime runtime = Runtime.getRuntime();
        long usedMemory = runtime.totalMemory() - runtime.freeMemory();

        Map<String, Object> response = new HashMap<>();
        response.put("message", "Memory cleared and GC suggested");
        response.put("used_memory_mb", usedMemory / (1024 * 1024));

        return ResponseEntity.ok(response);
    }

    /**
     * Get current JVM memory stats
     * 
     * Endpoint: GET /admin/memory-stats
     */
    @GetMapping("/memory-stats")
    public ResponseEntity<Map<String, Object>> getMemoryStats() {
        Runtime runtime = Runtime.getRuntime();

        long maxMemory = runtime.maxMemory();
        long totalMemory = runtime.totalMemory();
        long freeMemory = runtime.freeMemory();
        long usedMemory = totalMemory - freeMemory;

        Map<String, Object> stats = new HashMap<>();
        stats.put("max_heap_mb", maxMemory / (1024 * 1024));
        stats.put("total_heap_mb", totalMemory / (1024 * 1024));
        stats.put("used_heap_mb", usedMemory / (1024 * 1024));
        stats.put("free_heap_mb", freeMemory / (1024 * 1024));
        stats.put("usage_percent", (usedMemory * 100.0) / maxMemory);

        return ResponseEntity.ok(stats);
    }
}
