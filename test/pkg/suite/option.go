package suite

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	httpserver "github.com/nawafswe/go-service-starter-kit/internal/app/transport/http"
	httpbootstrap "github.com/nawafswe/go-service-starter-kit/internal/app/transport/http/bootstrap"
	"github.com/nawafswe/go-service-starter-kit/internal/auth"
	corehttp "github.com/nawafswe/mockchaos/core/http"
	"github.com/nawafswe/mockchaos/grpctest"
	"github.com/nawafswe/mockchaos/httptest"
	"google.golang.org/grpc"
)

// Option configures the test suite before the test runs.
type Option func(t *testing.T, s *Suite)

// WithMockHTTPServer starts a mock HTTP server using mockchaos and sets its URL
// on the suite config. The configFn receives the server URL so the caller can
// wire it into the appropriate config field.
func WithMockHTTPServer(handlers []corehttp.Handler, configFn func(s *Suite, serverURL string)) Option {
	return func(t *testing.T, s *Suite) {
		t.Helper()

		srv := httptest.NewServer(t, handlers...)
		t.Cleanup(func() { srv.Close() })

		configFn(s, srv.URL())
	}
}

// WithMockGRPCServer starts a mock gRPC server using mockchaos and sets its
// address on the suite config.
func WithMockGRPCServer(registerFn func(srv *grpc.Server), configFn func(s *Suite, addr string)) Option {
	return func(t *testing.T, s *Suite) {
		t.Helper()

		srv := grpctest.NewServer(t, registerFn)
		t.Cleanup(func() { srv.Close() })

		configFn(s, srv.Addr())
	}
}

// WithAuthMock enables the authorization mock so tests don't need real JWT tokens.
func WithAuthMock(userID, deviceID string) Option {
	return func(_ *testing.T, s *Suite) {
		s.Cfg.AuthorizationMock.Enabled = true
		s.Cfg.AuthorizationMock.UserID = userID
		s.Cfg.AuthorizationMock.DeviceID = deviceID
	}
}

// RunHTTPService boots the real HTTP server with the test suite's database and
// config. Returns the base URL. The server is shut down when the test ends.
func RunHTTPService(t *testing.T, s *Suite) string {
	t.Helper()

	resources := &httpbootstrap.SharedResource{
		Lgr: s.Lgr,
	}
	deps := &httpbootstrap.Dependencies{
		DBConn: s.SqlDB,
	}

	srv, err := httpserver.NewHTTPServer(context.Background(), s.Cfg, deps, resources)
	if err != nil {
		t.Fatalf("suite: create http server: %v", err)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("suite: listen: %v", err)
	}

	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			t.Logf("suite: http server error: %v", err)
		}
	}()

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	})

	return fmt.Sprintf("http://%s", ln.Addr().String())
}

// GenerateToken creates a signed JWT token for testing.
func GenerateToken(s *Suite, userID, deviceID string) (string, error) {
	cp, err := auth.NewClaimsParser(s.Cfg.JWT.ISSUER, []byte(s.Cfg.JWT.Secret))
	if err != nil {
		return "", fmt.Errorf("generate token: create claims parser: %w", err)
	}
	return cp.GenerateJWTToken(auth.User{
		ID:       userID,
		DeviceID: deviceID,
	})
}

// MustGenerateToken creates a signed JWT token and fails the test on error.
func MustGenerateToken(t *testing.T, s *Suite, userID, deviceID string) string {
	t.Helper()
	token, err := GenerateToken(s, userID, deviceID)
	if err != nil {
		t.Fatalf("suite: generate token: %v", err)
	}
	return token
}

// DoRequest performs an HTTP request against the test server and returns the
// response body and status code.
func DoRequest(t *testing.T, method, url string, body io.Reader, headers map[string]string) ([]byte, int) {
	t.Helper()

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatalf("suite: create request: %v", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("suite: do request: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("suite: read response: %v", err)
	}

	return respBody, resp.StatusCode
}
