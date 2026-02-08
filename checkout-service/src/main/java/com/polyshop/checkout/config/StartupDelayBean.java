package com.polyshop.checkout.config;

import jakarta.annotation.PostConstruct;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

/**
 * Startup Delay Configuration Bean
 * 
 * This component simulates a slow-starting application to test Kubernetes startupProbe configuration.
 * 
 * Purpose:
 * - In production, applications may have long initialization times (loading caches, warming up JVM, etc.)
 * - Kubernetes needs to distinguish between "still starting" vs "crashed and needs restart"
 * - The startupProbe gives the app extra time before liveness/readiness checks begin
 * 
 * K8s Probe Strategy:
 * 1. startupProbe: Runs FIRST, allows up to N failures before giving up
 *    - Example: initialDelaySeconds=5, periodSeconds=5, failureThreshold=6
 *    - Allows 30 seconds for startup (5 + 5*6 = 30s)
 *    - If still failing after 30s → CrashLoopBackOff
 * 
 * 2. livenessProbe: Runs AFTER startup succeeds
 *    - Checks if app is still alive (not deadlocked/frozen)
 *    - Restarts container if failing
 * 
 * 3. readinessProbe: Runs AFTER startup succeeds
 *    - Checks if app is ready to serve traffic
 *    - Removes from service load balancer if failing
 * 
 * Common Pitfall:
 * - If livenessProbe has short timeout + app has 15s startup → CrashLoopBackOff!
 * - Solution: Use startupProbe with longer timeout OR increase initialDelaySeconds
 * 
 * @see <a href="https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/">K8s Probe Docs</a>
 */
@Component
public class StartupDelayBean {
    
    private static final Logger logger = LoggerFactory.getLogger(StartupDelayBean.class);
    
    @Value("${checkout.startup-delay-seconds:15}")
    private int startupDelaySeconds;
    
    /**
     * Simulates slow application initialization
     * 
     * @PostConstruct runs AFTER dependency injection but BEFORE the app is marked as "started"
     * This means:
     * - Health endpoints are NOT yet available during this sleep
     * - K8s startupProbe will fail until this completes
     * - livenessProbe and readinessProbe won't run until startupProbe succeeds
     * 
     * Real-world scenarios this simulates:
     * - Loading large ML models into memory
     * - Populating Redis cache from database
     * - Establishing connection pools to external services
     * - JVM warm-up (JIT compilation, class loading)
     */
    @PostConstruct
    public void simulateSlowStartup() {
        logger.warn("⏳ Simulating slow startup - sleeping for {} seconds...", startupDelaySeconds);
        logger.warn("🔍 This tests Kubernetes startupProbe configuration");
        logger.warn("📖 In production, this could be cache warming, DB migration, etc.");
        
        try {
            // Sleep to simulate initialization work
            Thread.sleep(startupDelaySeconds * 1000L);
            logger.info("✅ Startup delay complete - application is now initializing");
        } catch (InterruptedException e) {
            logger.error("❌ Startup delay interrupted!", e);
            Thread.currentThread().interrupt();
        }
    }
}
