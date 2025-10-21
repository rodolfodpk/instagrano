// K6 Post Creation Load Test
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';
import { textSummary } from 'https://jslib.k6.io/k6-summary/0.0.1/index.js';
import { config, generateUserData, authHeader } from '../helpers/config.js';

export const options = {
  vus: 3,
  duration: '30s',
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
  
  // Create post with URL upload (using our new feature)
  const postRes = http.post(
    `${config.apiUrl}/api/posts`,
    {
      title: `Load Test Post ${Date.now()}`,
      caption: `This is a load test post created at ${new Date().toISOString()}`,
      media_url: 'https://via.placeholder.com/150.jpg', // Will be mapped to static image
    },
    { 
      headers: {
        'Authorization': `Bearer ${token}`,
      }
    }
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

