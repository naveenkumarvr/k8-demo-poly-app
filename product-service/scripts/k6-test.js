import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');

// Test configuration
export const options = {
    stages: [
        // Smoke test: Verify basic functionality
        { duration: '30s', target: 1 },

        // Ramp-up: Gradually increase load
        { duration: '1m', target: 10 },

        // Load test: Sustained load
        { duration: '2m', target: 50 },

        // Stress test: Push to higher load
        { duration: '1m', target: 100 },

        // Spike test: Sudden traffic increase
        { duration: '30s', target: 150 },

        // Ramp-down: Graceful decrease
        { duration: '1m', target: 0 },
    ],

    thresholds: {
        // HTTP request duration thresholds
        'http_req_duration': ['p(95)<500', 'p(99)<1000'],

        // Error rate should be less than 1%
        'errors': ['rate<0.01'],

        // 95% of requests should succeed
        'http_req_failed': ['rate<0.05'],
    },
};

// Base URL - update this to match your deployment
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
    // Scenario 1: Test /products endpoint (70% of traffic)
    if (Math.random() < 0.7) {
        testProductsEndpoint();
    }

    // Scenario 2: Test /stress endpoint (20% of traffic)
    else if (Math.random() < 0.9) {
        testStressEndpoint();
    }

    // Scenario 3: Test health endpoints (10% of traffic)
    else {
        testHealthEndpoints();
    }

    // Wait between requests (think time)
    sleep(1);
}

function testProductsEndpoint() {
    const response = http.get(`${BASE_URL}/products`);

    const result = check(response, {
        'products: status is 200': (r) => r.status === 200,
        'products: response time < 500ms': (r) => r.timings.duration < 500,
        'products: is JSON': (r) => r.headers['Content-Type'].includes('application/json'),
        'products: has array': (r) => {
            try {
                const body = JSON.parse(r.body);
                return Array.isArray(body);
            } catch (e) {
                return false;
            }
        },
        'products: has at least 10 products': (r) => {
            try {
                const body = JSON.parse(r.body);
                return body.length >= 10;
            } catch (e) {
                return false;
            }
        },
        'products: each product has required fields': (r) => {
            try {
                const products = JSON.parse(r.body);
                return products.every(p =>
                    p.id && p.name && p.description && p.price && p.image_url
                );
            } catch (e) {
                return false;
            }
        },
    });

    errorRate.add(!result);
}

function testStressEndpoint() {
    // Use different Fibonacci values for variety
    const fibonacciValues = [35, 38, 40, 42];
    const n = fibonacciValues[Math.floor(Math.random() * fibonacciValues.length)];

    const response = http.get(`${BASE_URL}/stress?n=${n}`);

    const result = check(response, {
        'stress: status is 200': (r) => r.status === 200,
        'stress: response time < 10s': (r) => r.timings.duration < 10000,
        'stress: is JSON': (r) => r.headers['Content-Type'].includes('application/json'),
        'stress: has computation result': (r) => {
            try {
                const body = JSON.parse(r.body);
                return body.result !== undefined && body.computation_time !== undefined;
            } catch (e) {
                return false;
            }
        },
        'stress: input matches request': (r) => {
            try {
                const body = JSON.parse(r.body);
                return body.input === n;
            } catch (e) {
                return false;
            }
        },
    });

    errorRate.add(!result);
}

function testHealthEndpoints() {
    const endpoints = ['/healthz', '/ready', '/live'];
    const endpoint = endpoints[Math.floor(Math.random() * endpoints.length)];

    const response = http.get(`${BASE_URL}${endpoint}`);

    const result = check(response, {
        'health: status is 200': (r) => r.status === 200,
        'health: response time < 100ms': (r) => r.timings.duration < 100,
        'health: is JSON': (r) => r.headers['Content-Type'].includes('application/json'),
        'health: has status field': (r) => {
            try {
                const body = JSON.parse(r.body);
                return body.status !== undefined;
            } catch (e) {
                return false;
            }
        },
    });

    errorRate.add(!result);
}

// Setup function - runs once before the test
export function setup() {
    console.log('Starting k6 load test for Product Service');
    console.log(`Base URL: ${BASE_URL}`);

    // Verify service is accessible
    const response = http.get(`${BASE_URL}/healthz`);
    if (response.status !== 200) {
        throw new Error('Service is not accessible. Please ensure it is running.');
    }

    console.log('Service is accessible. Starting load test...');
}

// Teardown function - runs once after the test
export function teardown(data) {
    console.log('Load test completed');
}
