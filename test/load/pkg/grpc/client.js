import { check } from 'k6';
import grpc from 'k6/net/grpc';
import { config } from '../config.js';
import { Rate, Counter } from 'k6/metrics';

const client = new grpc.Client();
client.load(['pkg/grpc'], 'protos/example.proto');

const SERVICE_PREFIX = '/example.v1.ExampleService/';

const methodMetrics = {
    CreateExample: {
        requestCounter: new Counter('CreateExample_Requests'),
        requestRate: new Rate('CreateExample_Success_Rate'),
    },
    GetExample: {
        requestCounter: new Counter('GetExample_Requests'),
        requestRate: new Rate('GetExample_Success_Rate'),
    },
    ListExamples: {
        requestCounter: new Counter('ListExamples_Requests'),
        requestRate: new Rate('ListExamples_Success_Rate'),
    },
    DeleteExample: {
        requestCounter: new Counter('DeleteExample_Requests'),
        requestRate: new Rate('DeleteExample_Success_Rate'),
    },
};

const invokeMethod = function (methodName, data, metadata) {
    client.connect(config.grpcHost, {
        timeout: '5s',
        plaintext: config.insecurePlaintext,
    });

    const params = metadata ? { metadata } : {};
    const response = client.invoke(SERVICE_PREFIX + methodName, data, params);

    const isOk = check(response, {
        [`${methodName} status is OK`]: (r) => r.status === grpc.StatusOK,
    });

    if (!isOk) {
        console.log(`${methodName} failed:`, JSON.stringify(response));
    }

    client.close();

    const metrics = methodMetrics[methodName];
    if (metrics) {
        metrics.requestRate.add(isOk);
        metrics.requestCounter.add(1);
    }

    return { ok: isOk, response };
};

export const ExampleService = function (method, data, metadata) {
    return invokeMethod(method, data, metadata);
};
