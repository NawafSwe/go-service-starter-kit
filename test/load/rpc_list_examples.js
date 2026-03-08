import { config, getLoadConfig, log as logConfig } from './pkg/config.js';
import { ExampleService } from './pkg/grpc/client.js';
import { sleep } from 'k6';

export const options = {
    discardResponseBodies: true,
    stages: getLoadConfig(config.loadType),
};

export function setup() {
    logConfig();
}

export default function () {
    ExampleService('ListExamples', {});

    if (config.sleepDuration > 0) {
        sleep(config.sleepDuration);
    }
}
