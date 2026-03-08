import { config, getLoadConfig, log as logConfig } from './pkg/config.js';
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

    const res = http.get(`${config.baseURL}/app/api/v1/examples`, params);

    check(res, {
        'status is 200': (r) => r.status === 200,
        'response body is not empty': (r) => r.body !== '',
        'response is array': (r) => Array.isArray(JSON.parse(r.body)),
    });

    if (config.sleepDuration > 0) {
        sleep(config.sleepDuration);
    }
}
