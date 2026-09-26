import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

/**
 * k6 Load Test for Checkout Service
 * 
 * This script tests the checkout flow end-to-end, simulating real user behavior.
 * 
 * Test Scenarios:
 * 1. POST /checkout - Process payment
 * 2. GET /currency/convert - Currency conversion
 * 3. GET /admin/memory-stats - Memory monitoring
 * 
 * Usage:
 * k6 run k6/checkout-flow.js
 * 
 * Metrics:
 * - Response time (p95 < 500ms)
 * - Success rate (> 70% for checkout)
 * - Throughput (requests per second)
 */

// Custom metrics
const checkoutSuccessRate = new Rate('checkout_success_rate');
const paymentFailureRate = new Rate('payment_failure_rate');

// Test configuration
export const options = {
    stages: [
        { duration: '10s', target: 5 },   // Ramp up to 5 VUs
        { duration: '30s', target: 10 },  // Stay at 10 VUs for 30s
        { duration: '10s', target: 0 },   // Ramp down
    ],
    thresholds: {
        'http_req_duration': ['p(95)<500'], // 95% of requests under 500ms
        'checkout_success_rate': ['rate>0.7'], // At least 70% success
        'http_req_failed': ['rate<0.1'],   // Less than 10% HTTP failures
    },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8085';

export default function () {
    const userId = `user-load-${__VU}-${__ITER}`;

    // Test 1: Process checkout
    const checkoutPayload = JSON.stringify({
        userId: userId,
        cartItems: [
            { productId: 'prod-1', quantity: 2 },
            { productId: 'prod-2', quantity: 1 },
        ],
    });

    const checkoutRes = http.post(`${BASE_URL}/checkout`, checkoutPayload, {
        headers: { 'Content-Type': 'application/json' },
    });

    const checkoutSuccess = check(checkoutRes, {
        'checkout status is 200': (r) => r.status === 200,
        'checkout has transaction_id': (r) => JSON.parse(r.body).transactionId !== undefined,
        'checkout response time < 500ms': (r) => r.timings.duration < 500,
    });

    // Track checkout results
    if (checkoutRes.status === 200) {
        const responseBody = JSON.parse(checkoutRes.body);
        if (responseBody.status === 'SUCCESS') {
            checkoutSuccessRate.add(1);
        } else if (responseBody.status === 'FAILED') {
            paymentFailureRate.add(1);
        }
    }

    sleep(0.5);

    // Test 2: Currency conversion
    const currencyRes = http.get(`${BASE_URL}/currency/convert?amount=100&from=USD&to=CAD`);

    check(currencyRes, {
        'currency status is 200': (r) => r.status === 200,
        'currency has converted_amount': (r) => JSON.parse(r.body).converted_amount !== undefined,
        'currency response time < 200ms': (r) => r.timings.duration < 200,
    });

    sleep(0.5);

    // Test 3: Memory stats (every 10th iteration)
    if (__ITER % 10 === 0) {
        const memoryRes = http.get(`${BASE_URL}/admin/memory-stats`);

        check(memoryRes, {
            'memory stats status is 200': (r) => r.status === 200,
            'memory stats has usage_percent': (r) => JSON.parse(r.body).usage_percent !== undefined,
        });
    }

    sleep(1);
}

export function handleSummary(data) {
    return {
        'stdout': textSummary(data, { indent: ' ', enableColors: true }),
    };
}

function textSummary(data, options) {
    const successRate = data.metrics.checkout_success_rate
        ? (data.metrics.checkout_success_rate.values.rate * 100).toFixed(2)
        : 'N/A';

    const p95Duration = data.metrics.http_req_duration
        ? data.metrics.http_req_duration.values['p(95)'].toFixed(2)
        : 'N/A';

    return `
=================================================================
  K6 Load Test Summary - Checkout Service
=================================================================
  
  Total Requests:        ${data.metrics.http_reqs ? data.metrics.http_reqs.values.count : 'N/A'}
  Request Duration (p95): ${p95Duration}ms
  Checkout Success Rate: ${successRate}%
  Failed Requests:       ${data.metrics.http_req_failed ? (data.metrics.http_req_failed.values.rate * 100).toFixed(2) : 'N/A'}%
  
  Virtual Users:         ${data.metrics.vus ? data.metrics.vus.values.max : 'N/A'}
  Test Duration:         ${data.state.testRunDurationMs / 1000}s
  
=================================================================
`;
}
