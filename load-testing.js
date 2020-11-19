import http from 'k6/http';
import { sleep, check } from 'k6';
import { Counter, Rate } from 'k6/metrics';

// A simple counter for http requests
export let errorCounter = new Counter("Error HTTP");
export let errorRate = new Rate("Error HTTP Rate")

export default function () {
    // our HTTP request, note that we are saving the response to res, which can be accessed later
    const res = http.get('http://localhost:1323/dora-the-explorer?is_rainy_day=true');

    const checkRes = check(res, {
        'status is 200': (r) => r.status === 200,
    });

    errorCounter.add(!checkRes);
    errorRate.add(!checkRes);
}
