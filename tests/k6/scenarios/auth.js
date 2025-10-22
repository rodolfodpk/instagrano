// K6 Authentication Load Test
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';
import { textSummary } from 'https://jslib.k6.io/k6-summary/0.0.1/index.js';
import { config, generateUserData, authHeader } from '../helpers/config.js';

export const options = {
  vus: 5,
  duration: '30s',
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
      `${config.apiUrl}/api/auth/me`,
      { headers: authHeader(token) }
    );
    
    check(meRes, {
      'me endpoint status is 200': (r) => r.status === 200,
      'me endpoint returns user_id': (r) => r.json('user.id') !== undefined,
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

