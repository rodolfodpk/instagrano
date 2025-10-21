import http from 'k6/http';
import { check, sleep } from 'k6';
import { textSummary } from 'https://jslib.k6.io/k6-summary/0.0.1/index.js';
import { config, generateUserData, authHeader } from '../helpers/config.js';

export const options = {
  vus: 3,
  duration: '30s',
  thresholds: config.thresholds,
};

const TEST_IMAGE_URLS = [
  'https://via.placeholder.com/150.jpg',
  'https://httpbin.org/image/png',
  'https://example.com/test.jpg',
];

export default function () {
  const userData = generateUserData();
  
  // Register and login
  http.post(`${config.apiUrl}/auth/register`, JSON.stringify(userData), 
    { headers: { 'Content-Type': 'application/json' } });
  
  const loginRes = http.post(`${config.apiUrl}/auth/login`, 
    JSON.stringify({ email: userData.email, password: userData.password }), 
    { headers: { 'Content-Type': 'application/json' } });
  
  const token = loginRes.json('token');
  if (!token) return;
  
  sleep(1);
  
  // Create post from URL
  const randomURL = TEST_IMAGE_URLS[Math.floor(Math.random() * TEST_IMAGE_URLS.length)];
  
  const postRes = http.post(
    `${config.apiUrl}/posts`,
    {
      title: `URL Post ${Date.now()}`,
      caption: 'Created from external URL',
      media_url: randomURL,
    },
    { headers: authHeader(token) }
  );
  
  check(postRes, {
    'post from URL status is 201': (r) => r.status === 201,
    'post has ID': (r) => r.json('id') !== undefined,
    'post has media_url': (r) => r.json('media_url') !== undefined,
  });
  
  sleep(2);
}

export function handleSummary(data) {
  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
    'tests/k6/results/posts-url-summary.json': JSON.stringify(data),
  };
}
