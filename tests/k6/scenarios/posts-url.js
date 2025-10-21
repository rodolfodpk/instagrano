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
  const registerRes = http.post(`${config.apiUrl}/api/auth/register`, JSON.stringify(userData), 
    { headers: { 'Content-Type': 'application/json' } });
  
  if (!check(registerRes, { 'registration successful': (r) => r.status === 200 })) {
    return;
  }
  
  const loginRes = http.post(`${config.apiUrl}/api/auth/login`, 
    JSON.stringify({ email: userData.email, password: userData.password }), 
    { headers: { 'Content-Type': 'application/json' } });
  
  if (!check(loginRes, { 'login successful': (r) => r.status === 200 })) {
    return;
  }
  
  const token = loginRes.json('token');
  if (!token) return;
  
  sleep(1);
  
  // Create post from URL
  const randomURL = TEST_IMAGE_URLS[Math.floor(Math.random() * TEST_IMAGE_URLS.length)];
  
  // Create post from URL using manual multipart construction
  const boundary = '----formdata-k6-' + Math.random().toString(36);
  const formData = [
    `--${boundary}`,
    `Content-Disposition: form-data; name="title"`,
    '',
    `URL Post ${Date.now()}`,
    `--${boundary}`,
    `Content-Disposition: form-data; name="caption"`,
    '',
    'Created from external URL',
    `--${boundary}`,
    `Content-Disposition: form-data; name="media_url"`,
    '',
    randomURL,
    `--${boundary}--`,
  ].join('\r\n');
  
  const postRes = http.post(
    `${config.apiUrl}/api/posts`,
    formData,
    { 
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': `multipart/form-data; boundary=${boundary}`,
      }
    }
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
