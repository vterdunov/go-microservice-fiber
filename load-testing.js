import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:3000';

export const options = {
  vus: 500,
  duration: '60s',
};

// Setup - create test users
export function setup() {
  const headers = { 'Content-Type': 'application/json' };

  for (let i = 1; i <= 5; i++) {
    const payload = JSON.stringify({
      name: `TestUser${i}`,
      email: `testuser${i}@example.com`,
    });
    http.post(`${BASE_URL}/api/users`, payload, { headers });
  }

  return { baseUrl: BASE_URL };
}

// Main test
export default function (data) {
  const res = http.get(`${data.baseUrl}/api/users`);

  check(res, {
    'status 200': (r) => r.status === 200,
  });
}
