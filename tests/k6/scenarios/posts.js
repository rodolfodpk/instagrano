// K6 Post Creation Load Test
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';
import { config, generateUserData, authHeader } from '../helpers/config.js';

export const options = {
  stages: [
    { duration: '30s', target: 5 },   // Ramp up to 5 users
    { duration: '1m', target: 10 },   // Ramp up to 10 users
    { duration: '1m', target: 10 },  // Stay at 10 users
    { duration: '30s', target: 0 },  // Ramp down
  ],
  thresholds: config.thresholds,
};

const errorRate = new Rate('errors');

export default function () {
  const userData = generateUserData();
  
  // Register and login
  const registerRes = http.post(
    `${config.apiUrl}/api/auth/register`,
    JSON.stringify({
      username: userData.username,
      email: userData.email,
      password: userData.password,
    }),
    { headers: { 'Content-Type': 'application/json' } }
  );
  
  const loginRes = http.post(
    `${config.apiUrl}/api/auth/login`,
    JSON.stringify({
      email: userData.email,
      password: userData.password,
    }),
    { headers: { 'Content-Type': 'application/json' } }
  );
  
  const token = loginRes.json('token');
  
  if (!token) {
    errorRate.add(1);
    return;
  }
  
  sleep(1);
  
  // Create post with file upload
  const postRes = http.post(
    `${config.apiUrl}/api/posts`,
    {
      title: `Load Test Post ${Date.now()}`,
      caption: `This is a load test post created at ${new Date().toISOString()}`,
      media_type: 'image',
      media: http.file('tests/k6/fixtures/test-image.png', 'test-image.png', 'image/png'),
    },
    { headers: authHeader(token) }
  );
  
  const postSuccess = check(postRes, {
    'post creation status is 201': (r) => r.status === 201,
    'post has ID': (r) => r.json('id') !== undefined,
    'post has media_url': (r) => r.json('media_url') !== undefined,
  });
  
  if (!postSuccess) {
    errorRate.add(1);
  }
  
  sleep(2);
}

export function handleSummary(data) {
  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
    'tests/k6/results/posts-summary.json': JSON.stringify(data),
  };
}

