import http from 'k6/http';
import { check } from 'k6';
import { Trend } from 'k6/metrics';
import { textSummary } from 'https://jslib.k6.io/k6-summary/0.0.1/index.js';
import { config, generateUserData, authHeader } from '../helpers/config.js';

export const options = {
  vus: 5,
  duration: '30s',
  thresholds: {
    ...config.thresholds,
    'cache_hit_duration': ['p(95)<50'],
    'cache_miss_duration': ['p(95)<200'],
  },
};

const cacheHitDuration = new Trend('cache_hit_duration');
const cacheMissDuration = new Trend('cache_miss_duration');

export default function () {
  const userData = generateUserData();
  
  // Register, login, create post
  http.post(`${config.apiUrl}/api/auth/register`, JSON.stringify(userData), 
    { headers: { 'Content-Type': 'application/json' } });
  
  const loginRes = http.post(`${config.apiUrl}/api/auth/login`, 
    JSON.stringify({ email: userData.email, password: userData.password }), 
    { headers: { 'Content-Type': 'application/json' } });
  
  const token = loginRes.json('token');
  if (!token) return;
  
  // Create a post with manual multipart construction
  const boundary = '----formdata-k6-' + Math.random().toString(36);
  const formData = [
    `--${boundary}`,
    `Content-Disposition: form-data; name="title"`,
    '',
    `Test Post ${Date.now()}`,
    `--${boundary}`,
    `Content-Disposition: form-data; name="caption"`,
    '',
    'Cache test',
    `--${boundary}`,
    `Content-Disposition: form-data; name="media_url"`,
    '',
    'https://via.placeholder.com/150.jpg',
    `--${boundary}--`,
  ].join('\r\n');
  
  const postRes = http.post(`${config.apiUrl}/api/posts`, formData, {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': `multipart/form-data; boundary=${boundary}`,
    }
  });
  
  const postID = postRes.json('id');
  if (!postID) return;
  
  // First request - cache miss
  const coldStart = Date.now();
  const coldRes = http.get(`${config.apiUrl}/api/posts/${postID}`, { headers: authHeader(token) });
  cacheMissDuration.add(Date.now() - coldStart);
  
  check(coldRes, { 'cold post status 200': (r) => r.status === 200 });
  
  // Second request - cache hit
  const warmStart = Date.now();
  const warmRes = http.get(`${config.apiUrl}/api/posts/${postID}`, { headers: authHeader(token) });
  cacheHitDuration.add(Date.now() - warmStart);
  
  check(warmRes, { 'warm post status 200': (r) => r.status === 200 });
}

export function handleSummary(data) {
  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
    'tests/k6/results/post-retrieval-summary.json': JSON.stringify(data),
  };
}
