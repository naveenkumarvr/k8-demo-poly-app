import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Counter } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const cartOperations = new Counter('cart_operations');

// Test configuration
export const options = {
    stages: [
        // Warmup: Verify basic functionality
        { duration: '30s', target: 5 },

        // Ramp-up: Gradually increase load
        { duration: '1m', target: 50 },

        // Sustained load: Maintain 50 concurrent users
        { duration: '2m', target: 50 },

        // Ramp-down: Graceful decrease
        { duration: '30s', target: 0 },
    ],

    thresholds: {
        // Cart operations should complete within 200ms for 95% of requests
        'http_req_duration{endpoint:add_item}': ['p(95)<200'],
        'http_req_duration{endpoint:get_cart}': ['p(95)<200'],
        'http_req_duration{endpoint:delete_cart}': ['p(95)<200'],

        // Stress endpoint can take longer (CPU-intensive)
        'http_req_duration{endpoint:stress}': ['p(95)<5000'],

        // Health checks should be fast
        'http_req_duration{endpoint:healthz}': ['p(95)<50'],

        // Error rate should be less than 1%
        'errors': ['rate<0.01'],
        'http_req_failed': ['rate<0.01'],
    },
};

// Base URL - can be overridden via environment variable
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Product catalog for realistic test data
const products = [
    'prod-001-laptop', 'prod-002-phone', 'prod-003-tablet',
    'prod-004-headphones', 'prod-005-keyboard', 'prod-006-mouse',
    'prod-007-monitor', 'prod-008-webcam', 'prod-009-speaker',
    'prod-010-charger'
];

// Generate random user ID
function randomUserId() {
    return `user-${Math.floor(Math.random() * 1000)}`;
}

// Generate random product ID
function randomProductId() {
    return products[Math.floor(Math.random() * products.length)];
}

export default function () {
    // Simulated user journey: Shopping cart workflow
    const userId = randomUserId();

    // Step 1: Add first item to cart (POST /v1/cart/:user_id)
    const product1 = randomProductId();
    const quantity1 = Math.floor(Math.random() * 5) + 1; // 1-5
    const addItemPayload1 = JSON.stringify({
        product_id: product1,
        quantity: quantity1,
    });

    const addItemResponse1 = http.post(
        `${BASE_URL}/v1/cart/${userId}`,
        addItemPayload1,
        {
            headers: { 'Content-Type': 'application/json' },
            tags: { endpoint: 'add_item' },
        }
    );

    const addItemResult1 = check(addItemResponse1, {
        'add_item: status is 200': (r) => r.status === 200,
        'add_item: response time < 200ms': (r) => r.timings.duration < 200,
        'add_item: is JSON': (r) => r.headers['Content-Type'].includes('application/json'),
        'add_item: has user_id': (r) => {
            try {
                const body = JSON.parse(r.body);
                return body.user_id === userId;
            } catch (e) {
                return false;
            }
        },
        'add_item: has items array': (r) => {
            try {
                const body = JSON.parse(r.body);
                return Array.isArray(body.items) && body.items.length > 0;
            } catch (e) {
                return false;
            }
        },
    });

    errorRate.add(!addItemResult1);
    cartOperations.add(1);

    sleep(0.5);

    // Step 2: Add another item to cart
    const product2 = randomProductId();
    const quantity2 = Math.floor(Math.random() * 3) + 1; // 1-3
    const addItemPayload2 = JSON.stringify({
        product_id: product2,
        quantity: quantity2,
    });

    const addItemResponse2 = http.post(
        `${BASE_URL}/v1/cart/${userId}`,
        addItemPayload2,
        {
            headers: { 'Content-Type': 'application/json' },
            tags: { endpoint: 'add_item' },
        }
    );

    const addItemResult2 = check(addItemResponse2, {
        'add_item2: status is 200': (r) => r.status === 200,
    });

    errorRate.add(!addItemResult2);
    cartOperations.add(1);

    sleep(0.3);

    // Step 3: Get cart (GET /v1/cart/:user_id)
    const getCartResponse = http.get(
        `${BASE_URL}/v1/cart/${userId}`,
        {
            tags: { endpoint: 'get_cart' },
        }
    );

    const getCartResult = check(getCartResponse, {
        'get_cart: status is 200': (r) => r.status === 200,
        'get_cart: response time < 200ms': (r) => r.timings.duration < 200,
        'get_cart: has items': (r) => {
            try {
                const body = JSON.parse(r.body);
                return body.total_items > 0;
            } catch (e) {
                return false;
            }
        },
    });

    errorRate.add(!getCartResult);
    cartOperations.add(1);

    sleep(0.5);

    // Step 4: Occasional stress test (20% of users)
    // This simulates users checking out or performing heavy operations
    if (Math.random() < 0.2) {
        const cpuIterations = 500; // Reduced for load testing
        const memoryMB = 50;

        const stressResponse = http.post(
            `${BASE_URL}/stress?cpu_iterations=${cpuIterations}&memory_mb=${memoryMB}`,
            null,
            {
                tags: { endpoint: 'stress' },
            }
        );

        const stressResult = check(stressResponse, {
            'stress: status is 200': (r) => r.status === 200,
            'stress: response time < 5s': (r) => r.timings.duration < 5000,
            'stress: has computation metrics': (r) => {
                try {
                    const body = JSON.parse(r.body);
                    return body.primes_calculated !== undefined && body.computation_time !== undefined;
                } catch (e) {
                    return false;
                }
            },
        });

        errorRate.add(!stressResult);
    }

    sleep(0.5);

    // Step 5: Clear cart (DELETE /v1/cart/:user_id)
    const deleteCartResponse = http.del(
        `${BASE_URL}/v1/cart/${userId}`,
        null,
        {
            tags: { endpoint: 'delete_cart' },
        }
    );

    const deleteCartResult = check(deleteCartResponse, {
        'delete_cart: status is 200': (r) => r.status === 200,
        'delete_cart: response time < 200ms': (r) => r.timings.duration < 200,
    });

    errorRate.add(!deleteCartResult);
    cartOperations.add(1);

    // Occasional health check (10% of iterations)
    if (Math.random() < 0.1) {
        const healthResponse = http.get(`${BASE_URL}/healthz`, {
            tags: { endpoint: 'healthz' },
        });

        const healthResult = check(healthResponse, {
            'healthz: status is 200': (r) => r.status === 200,
            'healthz: response time < 50ms': (r) => r.timings.duration < 50,
            'healthz: redis is healthy': (r) => {
                try {
                    const body = JSON.parse(r.body);
                    return body.status === 'healthy' && body.redis === 'healthy';
                } catch (e) {
                    return false;
                }
            },
        });

        errorRate.add(!healthResult);
    }

    // Think time between user journeys
    sleep(Math.random() * 2 + 1); // 1-3 seconds
}

// Setup function - runs once before the test
export function setup() {
    console.log('========================================');
    console.log('Starting k6 load test for Cart Service');
    console.log(`Base URL: ${BASE_URL}`);
    console.log('========================================');

    // Verify service is accessible
    const response = http.get(`${BASE_URL}/healthz`);
    if (response.status !== 200) {
        throw new Error(`Service is not accessible. Status: ${response.status}. Please ensure cart-service is running.`);
    }

    console.log('✓ Service is accessible');
    console.log('✓ Starting load test...\n');
}

// Teardown function - runs once after the test
export function teardown(data) {
    console.log('\n========================================');
    console.log('Load test completed');
    console.log('========================================');
}
