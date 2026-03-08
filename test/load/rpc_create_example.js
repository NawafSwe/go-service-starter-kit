import { config, getLoadConfig, log as logConfig } from './pkg/config.js';
import { ExampleService } from './pkg/grpc/client.js';
import { randomInt } from './pkg/util.js';
import { sleep } from 'k6';

export const options = {
    discardResponseBodies: true,
    stages: getLoadConfig(config.loadType),
};

const data = JSON.parse(open('./data/create_example_request.json'));

export function setup() {
    logConfig();
}

export default function () {
    const req = Object.assign({}, data, {
        name: `load-test-example-${randomInt(1, 100000)}`,
    });

    ExampleService('CreateExample', req);

    if (config.sleepDuration > 0) {
        sleep(config.sleepDuration);
    }
}
