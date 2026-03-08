// Package worker provides graceful-shutdown workers for long-running processes
// (HTTP, gRPC, message consumer). Each worker listens for SIGINT/SIGTERM and
// shuts down cleanly within a configurable timeout.
package worker
