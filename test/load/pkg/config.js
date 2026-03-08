const config = {
    baseURL: __ENV.K6_BASE_URL || 'http://localhost:8080',
    grpcHost: __ENV.K6_GRPC_HOST || 'localhost:50051',
    insecurePlaintext: (__ENV.K6_INSECURE_PLAINTEXT || 'true') === 'true',
    duration: __ENV.K6_DURATION || '30m',
    rampUpDuration: __ENV.K6_RAMP_UP_DURATION || '5m',
    vus: parseInt(__ENV.K6_VUS || '50'),
    loadType: __ENV.K6_LOAD_TYPE || 'linear',
    sleepDuration: parseFloat(__ENV.K6_SLEEP_DURATION || '1'),
    AUTHORIZATION: __ENV.K6_AUTHORIZATION || 'Bearer ',
};

const configMap = {
    linear: [
        {
            duration: config.duration,
            target: config.vus,
        },
    ],
    load: [
        {
            duration: config.rampUpDuration,
            target: config.vus,
        },
        {
            duration: config.duration,
            target: config.vus,
        },
    ],
    spike: [
        {
            duration: '10s',
            target: config.vus,
        },
        {
            duration: '1m',
            target: config.vus * 5,
        },
        {
            duration: '10s',
            target: config.vus,
        },
    ],
    stress: [
        {
            duration: config.rampUpDuration,
            target: config.vus,
        },
        {
            duration: config.duration,
            target: config.vus * 3,
        },
        {
            duration: '5m',
            target: 0,
        },
    ],
};

function getLoadConfig(loadType) {
    return configMap[loadType] || configMap['linear'];
}

function log() {
    console.log('=== Load Test Configuration ===');
    console.log(`Base URL:       ${config.baseURL}`);
    console.log(`gRPC Host:      ${config.grpcHost}`);
    console.log(`Load Type:      ${config.loadType}`);
    console.log(`VUs:            ${config.vus}`);
    console.log(`Duration:       ${config.duration}`);
    console.log(`Sleep:          ${config.sleepDuration}s`);
    console.log('===============================');
}

export { config, getLoadConfig, log };
