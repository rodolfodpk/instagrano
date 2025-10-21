// K6 Full User Journey Test
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';
import { config, generateUserData, authHeader } from '../helpers/config.js';

export const options = {
  stages: [
    { duration: '1m', target: 10 },  // Ramp up to 10 concurrent users
    { duration: '2m', target: 10 },  // Stay at 10 users
    { duration: '30s', target: 0 }, // Ramp down
  ],
  thresholds: config.thresholds,
};

const errorRate = new Rate('errors');

export default function () {
  const userData = generateUserData();
  
  // Step 1: Register
  const registerRes = http.post(
    `${config.apiUrl}/api/auth/register`,
    JSON.stringify({
      username: userData.username,
      email: userData.email,
      password: userData.password,
    }),
    { headers: { 'Content-Type': 'application/json' } }
  );
  
  if (!check(registerRes, { 'register status is 200': (r) => r.status === 200 })) {
    errorRate.add(1);
    return;
  }
  
  sleep(1);
  
  // Step 2: Login
  const loginRes = http.post(
    `${config.apiUrl}/api/auth/login`,
    JSON.stringify({
      email: userData.email,
      password: userData.password,
    }),
    { headers: { 'Content-Type': 'application/json' } }
  );
  
  const token = loginRes.json('token');
  
  if (!check(loginRes, { 'login status is 200': (r) => r.status === 200 }) || !token) {
    errorRate.add(1);
    return;
  }
  
  sleep(1);
  
  // Step 3: Create Post
  const postRes = http.post(
    `${config.apiUrl}/api/posts`,
    {
      title: `User Journey Post ${Date.now()}`,
      caption: 'Complete user journey test',
      media_type: 'image',
      media: http.file('tests/k6/fixtures/test-image.png', 'test-image.png', 'image/png'),
    },
    { headers: authHeader(token) }
  );
  
  const postId = postRes.json('id');
  
  if (!check(postRes, { 'post creation status is 201': (r) => r.status === 201 })) {
    errorRate.add(1);
    return;
  }
  
  sleep(1);
  
  // Step 4: View Feed
  const feedRes = http.get(
    `${config.apiUrl}/api/feed?limit=10`,
    { headers: authHeader(token) }
  );
  
  check(feedRes, {
    'feed status is 200': (r) => r.status === 200,
    'feed has posts': (r) => r.json('posts') !== undefined,
  });
  
  sleep(1);
  
  // Step 5: Like Post (if we have a post ID)
  if (postId) {
    const likeRes = http.post(
      `${config.apiUrl}/api/posts/${postId}/like`,
      null,
      { headers: authHeader(token) }
    );
    
    check(likeRes, {
      'like status is 200': (r) => r.status === 200,
    });
    
    sleep(1);
    
    // Step 6: Comment on Post
    const commentRes = http.post(
      `${config.apiUrl}/api/posts/${postId}/comment`,
      JSON.stringify({ text: 'Great post from K6 load test!' }),
      { headers: authHeader(token) }
    );
    
    check(commentRes, {
      'comment status is 200': (r) => r.status === 200,
    });
  }
  
  sleep(1);
}

export function handleSummary(data) {
  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
    'tests/k6/results/journey-summary.json': JSON.stringify(data),
  };
}

