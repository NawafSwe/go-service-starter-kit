import { config, getLoadConfig, log as logConfig } from './pkg/config.js';
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    discardResponseBodies: false,
    stages: getLoadConfig(config.loadType),
};

// Setup creates a test example and returns its ID for use in the default function.
export function setup() {
    logConfig();

    const params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': config.AUTHORIZATION,
        },
    };

    const payload = JSON.stringify({ name: 'load-test-seed' });
    const res = http.post(`${config.baseURL}/app/api/v1/examples`, payload, params);

    if (res.status !== 201) {
        console.error(`setup: failed to create seed example: ${res.status} ${res.body}`);
        return { id: '' };
    }

    const body = JSON.parse(res.body);
    console.log(`setup: seeded example ${body.id}`);
    return { id: body.id };
}

export default function (data) {
    if (!data.id) {
        console.error('no seed example ID — skipping');
        return;
    }

    const params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': config.AUTHORIZATION,
        },
    };

    const res = http.get(`${config.baseURL}/app/api/v1/examples/${data.id}`, params);

    check(res, {
        'status is 200': (r) => r.status === 200,
        'response body is not empty': (r) => r.body !== '',
        'response has correct id': (r) => JSON.parse(r.body).id === data.id,
    });

    if (config.sleepDuration > 0) {
        sleep(config.sleepDuration);
    }
}
