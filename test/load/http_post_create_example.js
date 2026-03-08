import { config, getLoadConfig, log as logConfig } from './pkg/config.js';
import { randomInt } from './pkg/util.js';
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    discardResponseBodies: false,
    stages: getLoadConfig(config.loadType),
};

export function setup() {
    logConfig();
}

export default function () {
    const params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': config.AUTHORIZATION,
        },
    };

    const payload = JSON.stringify({
        name: `load-test-example-${randomInt(1, 100000)}`,
    });

    const res = http.post(`${config.baseURL}/app/api/v1/examples`, payload, params);

    check(res, {
        'status is 201': (r) => r.status === 201,
        'response body is not empty': (r) => r.body !== '',
        'response has id': (r) => JSON.parse(r.body).id !== '',
    });

    if (config.sleepDuration > 0) {
        sleep(config.sleepDuration);
    }
}
