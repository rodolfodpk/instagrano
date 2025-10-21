// K6 Feed Cache Performance Test
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend } from 'k6/metrics';
import { textSummary } from 'https://jslib.k6.io/k6-summary/0.0.1/index.js';
import { config, generateUserData, authHeader, logCacheComparison } from '../helpers/config.js';

export const options = {
  vus: 3,
  duration: '45s',
  thresholds: {
    ...config.thresholds,
    'cache_hit_duration': ['p(95)<100'],
    'cache_miss_duration': ['p(95)<500'],
  },
};

const cacheHitDuration = new Trend('cache_hit_duration');
const cacheMissDuration = new Trend('cache_miss_duration');

export default function () {
  const userData = generateUserData();
  
  // Step 1: Register and login
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
    return;
  }
  
  sleep(1);
  
  // Step 2: Create a post to ensure feed has data
  const postRes = http.post(
    `${config.apiUrl}/api/posts`,
    {
      title: `Test Post ${Date.now()}`,
      caption: 'K6 load test post',
      media_type: 'image',
      media: http.file('tests/k6/fixtures/test-image.png', 'test-image.png', 'image/png'),
    },
    { headers: authHeader(token) }
  );
  
  sleep(2);
  
  // Step 3: First feed request (COLD CACHE - database query)
  const coldStart = Date.now();
  const coldFeedRes = http.get(
    `${config.apiUrl}/api/feed?limit=10`,
    { headers: authHeader(token) }
  );
  const coldDuration = Date.now() - coldStart;
  
  check(coldFeedRes, {
    'cold feed status is 200': (r) => r.status === 200,
    'cold feed has posts': (r) => r.json('posts') !== undefined,
  });
  
  cacheMissDuration.add(coldDuration);
  
  sleep(1);
  
  // Step 4: Second feed request (WARM CACHE - Redis hit)
  const warmStart = Date.now();
  const warmFeedRes = http.get(
    `${config.apiUrl}/api/feed?limit=10`,
    { headers: authHeader(token) }
  );
  const warmDuration = Date.now() - warmStart;
  
  check(warmFeedRes, {
    'warm feed status is 200': (r) => r.status === 200,
    'warm feed has posts': (r) => r.json('posts') !== undefined,
  });
  
  cacheHitDuration.add(warmDuration);
  
  // Log comparison
  logCacheComparison(coldDuration, warmDuration);
  
  sleep(1);
}

export function handleSummary(data) {
  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
    'tests/k6/results/cache-summary.json': JSON.stringify(data),
  };
}

