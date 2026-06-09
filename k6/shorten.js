import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 50 },
    { duration: '1m', target: 200 },
    { duration: '30s', target: 50 },
    { duration: '10s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],
    http_req_failed: ['rate<0.01'],
  },
};

export default function () {
  const payload = JSON.stringify({
    long_url: 'https://example.com/some/very/long/path/for/testing/' + __VU + '-' + __ITER,
  });
  const res = http.post('http://localhost:8080/api/shorten', payload, {
    headers: { 'Content-Type': 'application/json' },
  });
  check(res, {
    'status is 201': (r) => r.status === 201,
    'has short_code': (r) => JSON.parse(r.body).short_code !== undefined,
  });
  sleep(1);
}
