// K6 Authentication Load Test
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';
import { config, generateUserData, authHeader } from '../helpers/config.js';

export const options = {
  stages: [
    { duration: '30s', target: 10 },  // Ramp up to 10 users
    { duration: '1m', target: 10 },    // Stay at 10 users
    { duration: '30s', target: 20 },   // Ramp up to 20 users
    { duration: '1m', target: 20 },   // Stay at 20 users
    { duration: '30s', target: 0 },    // Ramp down to 0 users
  ],
  thresholds: config.thresholds,
};

const errorRate = new Rate('errors');

export default function () {
  const userData = generateUserData();
  
  // Test 1: User Registration
  const registerRes = http.post(
    `${config.apiUrl}/api/auth/register`,
    JSON.stringify({
      username: userData.username,
      email: userData.email,
      password: userData.password,
    }),
    { headers: { 'Content-Type': 'application/json' } }
  );
  
  const registerSuccess = check(registerRes, {
    'registration status is 200': (r) => r.status === 200,
    'registration has user data': (r) => r.json('user') !== undefined,
  });
  
  if (!registerSuccess) {
    errorRate.add(1);
  }
  
  sleep(1);
  
  // Test 2: User Login
  const loginRes = http.post(
    `${config.apiUrl}/api/auth/login`,
    JSON.stringify({
      email: userData.email,
      password: userData.password,
    }),
    { headers: { 'Content-Type': 'application/json' } }
  );
  
  const loginSuccess = check(loginRes, {
    'login status is 200': (r) => r.status === 200,
    'login returns token': (r) => r.json('token') !== undefined,
    'login returns user data': (r) => r.json('user') !== undefined,
  });
  
  if (!loginSuccess) {
    errorRate.add(1);
  }
  
  const token = loginRes.json('token');
  
  sleep(1);
  
  // Test 3: Authenticated Request (/me)
  if (token) {
    const meRes = http.get(
      `${config.apiUrl}/api/me`,
      { headers: authHeader(token) }
    );
    
    check(meRes, {
      'me endpoint status is 200': (r) => r.status === 200,
      'me endpoint returns user_id': (r) => r.json('user_id') !== undefined,
    });
  }
  
  sleep(1);
}

export function handleSummary(data) {
  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
    'tests/k6/results/auth-summary.json': JSON.stringify(data),
  };
}

