import { config, getLoadConfig, log as logConfig } from './pkg/config.js';
import { ExampleService } from './pkg/grpc/client.js';
import { randomInt } from './pkg/util.js';
import { sleep } from 'k6';

export const options = {
    discardResponseBodies: true,
    stages: getLoadConfig(config.loadType),
};

export function setup() {
    logConfig();
}

// Each iteration creates an example via gRPC then deletes it.
export default function () {
    const { ok, response } = ExampleService('CreateExample', {
        name: `load-test-delete-${randomInt(1, 100000)}`,
    });

    if (!ok) {
        return;
    }

    const id = response.message.example.id;
    ExampleService('DeleteExample', { id });

    if (config.sleepDuration > 0) {
        sleep(config.sleepDuration);
    }
}
