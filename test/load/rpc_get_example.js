import { config, getLoadConfig, log as logConfig } from './pkg/config.js';
import { ExampleService } from './pkg/grpc/client.js';
import { sleep } from 'k6';

export const options = {
    discardResponseBodies: true,
    stages: getLoadConfig(config.loadType),
};

// Setup creates a seed example via gRPC and returns its ID for the load test.
export function setup() {
    logConfig();

    const { ok, response } = ExampleService('CreateExample', { name: 'load-test-seed' });
    if (!ok) {
        console.error('setup: failed to create seed example');
        return { id: '' };
    }

    const id = response.message.example.id;
    console.log(`setup: seeded example ${id}`);
    return { id };
}

export default function (data) {
    if (!data.id) {
        console.error('no seed example ID — skipping');
        return;
    }

    ExampleService('GetExample', { id: data.id });

    if (config.sleepDuration > 0) {
        sleep(config.sleepDuration);
    }
}
