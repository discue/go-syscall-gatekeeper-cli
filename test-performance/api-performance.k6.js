import { tagWithCurrentStageProfile } from 'https://jslib.k6.io/k6-utils/1.3.0/index.js'
import { check, group } from 'k6'
import http from 'k6/http'

const target = 125

export const options = {
  scenarios: {
    warmup: {
      executor: 'ramping-vus',
      startVUs: 75,
      stages: [
        { duration: '5m', target }
      ],
      gracefulRampDown: '5s',
      startTime: '0s'
    },
  },
  thresholds: { http_req_duration: ['avg<300', 'p(95)<700', 'p(99)<1000'] },
  maxRedirects: 0,
  userAgent: 'dsq-k6s/1.0',
  summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(50)', 'p(75)', 'p(95)', 'p(99)', 'p(99.99)', 'count'],
  tlsVersion: {
    min: 'tls1.2',
    max: 'tls1.3',
  }
}

export default function () {
  tagWithCurrentStageProfile()

  const options = {
    headers: {
    }
  }

  group('queues', () => {
    const res = http.get('http://host.docker.internal:8080', options)

    check(res, {
      'status is 200': (r) => r.status === 200,
      'result is 8': (r) => r.body.includes('Hello World')
    })
  })
}