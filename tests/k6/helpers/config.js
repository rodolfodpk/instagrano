// Shared configuration for K6 tests

export const config = {
  // API base URL
  apiUrl: __ENV.API_URL || 'http://localhost:8080',
  
  // Test thresholds (basic - specific tests can override)
  thresholds: {
    // Response time thresholds
    'http_req_duration': [
      'p(95)<500',  // 95% of requests should be below 500ms
      'p(99)<1000', // 99% of requests should be below 1s
    ],
    
    // Error rate threshold
    'http_req_failed': ['rate<0.01'], // Less than 1% errors
  },
  
  // Summary statistics
  summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(90)', 'p(95)', 'p(99)', 'count'],
};

// Helper function to generate unique user data
export function generateUserData() {
  const timestamp = Date.now();
  const random = Math.floor(Math.random() * 10000);
  
  return {
    username: `user_${timestamp}_${random}`,
    email: `user_${timestamp}_${random}@test.com`,
    password: 'testpassword123',
  };
}

// Helper function to create JWT auth header
export function authHeader(token) {
  return {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json',
  };
}

// Helper function to log cache performance comparison (disabled for cleaner output)
export function logCacheComparison(coldDuration, warmDuration) {
  // Logging disabled to reduce K6 test verbosity
  // const improvement = ((coldDuration - warmDuration) / coldDuration * 100).toFixed(1);
  // console.log(`Cache Performance: Cold=${coldDuration}ms, Warm=${warmDuration}ms, Improvement=${improvement}%`);
}

