import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 500 },
    { duration: '1m', target: 2000 },
    { duration: '30s', target: 500 },
    { duration: '10s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(99)<50'],
    http_req_failed: ['rate<0.01'],
  },
};

const shortCodes = ['test1', 'test2', 'test3', 'test4', 'test5'];

export default function () {
  const code = shortCodes[__VU % shortCodes.length];
  const res = http.get(`http://localhost:8080/${code}`, { redirects: 0 });
  check(res, {
    'status is 301': (r) => r.status === 301,
    'has location header': (r) => r.headers.Location !== undefined,
  });
  sleep(0.1);
}
