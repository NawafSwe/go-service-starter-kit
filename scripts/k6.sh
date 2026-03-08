#!/bin/bash
# k6 load test runner.
#
# Usage:
#   ./scripts/k6.sh                                  # list available tests
#   ./scripts/k6.sh http_post_create_example          # run a test
#   K6_VUS=100 K6_DURATION=5m ./scripts/k6.sh http_get_list_examples
#
# Environment variables (K6_ prefix):
#   K6_BASE_URL             HTTP base URL     (default: http://localhost:8080)
#   K6_GRPC_HOST            gRPC host:port    (default: localhost:50051)
#   K6_INSECURE_PLAINTEXT   Use plaintext gRPC (default: true)
#   K6_VUS                  Virtual users     (default: 50)
#   K6_DURATION             Test duration     (default: 30m)
#   K6_RAMP_UP_DURATION     Ramp-up time      (default: 5m)
#   K6_LOAD_TYPE            Profile: linear | load | spike | stress (default: linear)
#   K6_SLEEP_DURATION       Pause between requests in seconds       (default: 1)
#   K6_AUTHORIZATION        Bearer token header value

set -euo pipefail

directory="./test/load"

if [ $# -eq 0 ]; then
    echo "Available load tests:"
    echo ""
    find "$directory" -maxdepth 1 -type f -name "*.js" -exec basename {} \; | sed 's/\.js$//' | sort
    echo ""
    echo "Usage: $0 <test-name>"
    echo "Example: K6_VUS=100 $0 http_post_create_example"
    exit 0
fi

test=$1
js_file="$test.js"

if [ ! -e "$directory/$js_file" ]; then
    echo "Error: unknown test '$test'."
    echo "Run '$0' without arguments to list available tests."
    exit 1
fi

echo "Running ${test}"
k6 run "$directory/$js_file" \
    -e K6_BASE_URL="${K6_BASE_URL:-}" \
    -e K6_GRPC_HOST="${K6_GRPC_HOST:-}" \
    -e K6_INSECURE_PLAINTEXT="${K6_INSECURE_PLAINTEXT:-}" \
    -e K6_DURATION="${K6_DURATION:-}" \
    -e K6_VUS="${K6_VUS:-}" \
    -e K6_RAMP_UP_DURATION="${K6_RAMP_UP_DURATION:-}" \
    -e K6_LOAD_TYPE="${K6_LOAD_TYPE:-}" \
    -e K6_SLEEP_DURATION="${K6_SLEEP_DURATION:-}" \
    -e K6_AUTHORIZATION="${K6_AUTHORIZATION:-}"
