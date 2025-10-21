// K6 Authentication Load Test (Quick Version)
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';
import { config, generateUserData, authHeader } from '../helpers/config.js';

export const options = {
  vus: 5,
  duration: '15s',
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
  
  sleep(0.5);
  
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
  
  sleep(0.5);
  
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
  
  sleep(0.5);
}

