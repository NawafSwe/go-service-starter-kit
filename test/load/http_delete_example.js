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

// Each VU iteration creates an example and then deletes it.
export default function () {
    const params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': config.AUTHORIZATION,
        },
    };

    // Create
    const payload = JSON.stringify({
        name: `load-test-delete-${randomInt(1, 100000)}`,
    });
    const createRes = http.post(`${config.baseURL}/app/api/v1/examples`, payload, params);

    check(createRes, {
        'create: status is 201': (r) => r.status === 201,
    });

    if (createRes.status !== 201) {
        return;
    }

    const id = JSON.parse(createRes.body).id;

    // Delete
    const deleteRes = http.del(`${config.baseURL}/app/api/v1/examples/${id}`, null, params);

    check(deleteRes, {
        'delete: status is 204': (r) => r.status === 204,
    });

    if (config.sleepDuration > 0) {
        sleep(config.sleepDuration);
    }
}
