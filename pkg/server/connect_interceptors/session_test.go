package connectinterceptors

import (
	"context"
	"net/http"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/userpat"
	"github.com/raystack/frontier/pkg/server/consts"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func TestExtractAuthCredentials(t *testing.T) {
	tests := []struct {
		name   string
		header string
		scheme string
		want   string
	}{
		{name: "exact match", header: "Bearer abc", scheme: "Bearer", want: "abc"},
		{name: "lowercase scheme", header: "bearer abc", scheme: "Bearer", want: "abc"},
		{name: "uppercase scheme", header: "BEARER abc", scheme: "Bearer", want: "abc"},
		{name: "mixed case scheme", header: "BeArEr abc", scheme: "Bearer", want: "abc"},
		{name: "basic exact", header: "Basic Y2lkOnNlYw==", scheme: "Basic", want: "Y2lkOnNlYw=="},
		{name: "basic lowercase", header: "basic Y2lkOnNlYw==", scheme: "Basic", want: "Y2lkOnNlYw=="},
		{name: "no scheme", header: "abc", scheme: "Bearer", want: ""},
		{name: "wrong scheme", header: "Basic abc", scheme: "Bearer", want: ""},
		{name: "scheme without separator", header: "Bearerabc", scheme: "Bearer", want: ""},
		{name: "empty credentials", header: "Bearer ", scheme: "Bearer", want: ""},
		{name: "credentials with surrounding whitespace", header: "Bearer   abc  ", scheme: "Bearer", want: "abc"},
		{name: "header too short", header: "B", scheme: "Bearer", want: ""},
		{name: "empty header", header: "", scheme: "Bearer", want: ""},
		{name: "scheme only no space", header: "Bearer", scheme: "Bearer", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractAuthCredentials(tt.header, tt.scheme)
			assert.Equal(t, tt.want, got)
		})
	}
}

func buildJWT(t *testing.T, jti string, expiresAt time.Time) string {
	t.Helper()
	key := []byte("test-secret-key-for-hmac-signing-32b")
	builder := jwt.NewBuilder().Expiration(expiresAt)
	if jti != "" {
		builder = builder.JwtID(jti)
	}
	tok, err := builder.Build()
	require.NoError(t, err)
	signed, err := jwt.Sign(tok, jwt.WithKey(jwa.HS256, key))
	require.NoError(t, err)
	return string(signed)
}

func TestApplyAuthorizationHeader(t *testing.T) {
	patConf := userpat.Config{Prefix: "fpt"}

	validJWTStr := buildJWT(t, "jti-123", time.Now().Add(time.Hour))
	expiredJWTStr := buildJWT(t, "jti-old", time.Now().Add(-time.Hour))
	jwtNoJTIStr := buildJWT(t, "", time.Now().Add(time.Hour))

	tests := []struct {
		name       string
		header     string
		wantToken  string
		wantSecret string
	}{
		{
			name:      "Bearer with PAT",
			header:    "Bearer fpt_abc123",
			wantToken: "fpt_abc123",
		},
		{
			name:      "lowercase bearer with PAT",
			header:    "bearer fpt_abc123",
			wantToken: "fpt_abc123",
		},
		{
			name:      "uppercase BEARER with PAT",
			header:    "BEARER fpt_abc123",
			wantToken: "fpt_abc123",
		},
		{
			name:   "bare PAT without scheme is rejected",
			header: "fpt_abc123",
		},
		{
			name:   "Bearer with random non-PAT non-JWT value is rejected",
			header: "Bearer randomstring",
		},
		{
			name:      "Bearer with valid JWT",
			header:    "Bearer " + validJWTStr,
			wantToken: validJWTStr,
		},
		{
			name:   "Bearer with expired JWT is rejected",
			header: "Bearer " + expiredJWTStr,
		},
		{
			name:   "Bearer with JWT missing JTI is rejected",
			header: "Bearer " + jwtNoJTIStr,
		},
		{
			name:       "Basic with credentials",
			header:     "Basic Y2lkOnNlYw==",
			wantSecret: "Y2lkOnNlYw==",
		},
		{
			name:       "lowercase basic with credentials",
			header:     "basic Y2lkOnNlYw==",
			wantSecret: "Y2lkOnNlYw==",
		},
		{
			name:   "bare credentials without scheme are rejected",
			header: "Y2lkOnNlYw==",
		},
		{
			name:   "unknown scheme is rejected",
			header: "DPoP fpt_abc123",
		},
		{
			name:   "empty header is rejected",
			header: "",
		},
		{
			name:   "scheme only no credentials is rejected",
			header: "Bearer ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := metadata.MD{}
			applyAuthorizationHeader(md, tt.header, patConf)

			gotToken := md.Get(consts.UserTokenGatewayKey)
			gotSecret := md.Get(consts.UserSecretGatewayKey)

			if tt.wantToken == "" {
				assert.Empty(t, gotToken, "expected no token to be set")
			} else {
				assert.Equal(t, []string{tt.wantToken}, gotToken)
			}
			if tt.wantSecret == "" {
				assert.Empty(t, gotSecret, "expected no secret to be set")
			} else {
				assert.Equal(t, []string{tt.wantSecret}, gotSecret)
			}

			if len(gotToken) > 0 {
				assert.Empty(t, gotSecret, "Bearer must not populate the secret slot")
			}
			if len(gotSecret) > 0 {
				assert.Empty(t, gotToken, "Basic must not populate the token slot")
			}
		})
	}
}

func TestApplyAuthorizationHeader_EmptyPATPrefix(t *testing.T) {
	md := metadata.MD{}
	applyAuthorizationHeader(md, "Bearer fpt_abc123", userpat.Config{Prefix: ""})
	assert.Empty(t, md.Get(consts.UserTokenGatewayKey), "PAT path must be disabled when prefix is empty")
}

// fakeStreamingHandlerConn is a minimal StreamingHandlerConn used to exercise
// SessionInterceptor.WrapStreamingHandler. Only RequestHeader() needs to do real work.
type fakeStreamingHandlerConn struct {
	requestHeader http.Header
}

func (f *fakeStreamingHandlerConn) Spec() connect.Spec           { return connect.Spec{} }
func (f *fakeStreamingHandlerConn) Peer() connect.Peer           { return connect.Peer{} }
func (f *fakeStreamingHandlerConn) Receive(_ any) error          { return nil }
func (f *fakeStreamingHandlerConn) RequestHeader() http.Header   { return f.requestHeader }
func (f *fakeStreamingHandlerConn) Send(_ any) error             { return nil }
func (f *fakeStreamingHandlerConn) ResponseHeader() http.Header  { return http.Header{} }
func (f *fakeStreamingHandlerConn) ResponseTrailer() http.Header { return http.Header{} }

func newTestSessionInterceptor() *SessionInterceptor {
	return NewSessionInterceptor(nil, authenticate.SessionConfig{}, nil, userpat.Config{Prefix: "fpt"})
}

// captureMetadataUnary runs the SessionInterceptor's WrapUnary wrapper and returns
// the metadata captured inside the wrapped UnaryFunc's context.
func captureMetadataUnary(t *testing.T, s *SessionInterceptor, headers http.Header) metadata.MD {
	t.Helper()
	var captured metadata.MD
	next := connect.UnaryFunc(func(ctx context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		captured = md
		return connect.NewResponse(&frontierv1beta1.GetCurrentUserResponse{}), nil
	})
	req := connect.NewRequest(&frontierv1beta1.GetCurrentUserRequest{})
	for k, vs := range headers {
		for _, v := range vs {
			req.Header().Add(k, v)
		}
	}
	_, err := s.WrapUnary(next)(context.Background(), req)
	require.NoError(t, err)
	return captured
}

// captureMetadataStreaming exercises WrapStreamingHandler the same way.
func captureMetadataStreaming(t *testing.T, s *SessionInterceptor, headers http.Header) metadata.MD {
	t.Helper()
	var captured metadata.MD
	next := connect.StreamingHandlerFunc(func(ctx context.Context, _ connect.StreamingHandlerConn) error {
		md, _ := metadata.FromIncomingContext(ctx)
		captured = md
		return nil
	})
	conn := &fakeStreamingHandlerConn{requestHeader: headers.Clone()}
	require.NoError(t, s.WrapStreamingHandler(next)(context.Background(), conn))
	return captured
}

// captureMetadataAnnotator exercises UnaryConnectRequestHeadersAnnotator.
func captureMetadataAnnotator(t *testing.T, s *SessionInterceptor, headers http.Header) metadata.MD {
	t.Helper()
	var captured metadata.MD
	next := connect.UnaryFunc(func(ctx context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		captured = md
		return connect.NewResponse(&frontierv1beta1.GetCurrentUserResponse{}), nil
	})
	req := connect.NewRequest(&frontierv1beta1.GetCurrentUserRequest{})
	for k, vs := range headers {
		for _, v := range vs {
			req.Header().Add(k, v)
		}
	}
	wrapped := s.UnaryConnectRequestHeadersAnnotator()(next)
	_, err := wrapped(context.Background(), req)
	require.NoError(t, err)
	return captured
}

// TestSessionInterceptor_AuthFlow_EndToEnd is the main regression test: it verifies that
// across all three transports (unary, streaming, headers annotator), Authorization headers
// are routed to the correct gateway metadata slot and surface through
// authenticate.GetTokenFromContext / authenticate.GetSecretFromContext exactly as expected.
func TestSessionInterceptor_AuthFlow_EndToEnd(t *testing.T) {
	s := newTestSessionInterceptor()

	validJWT := buildJWT(t, "jti-end2end", time.Now().Add(time.Hour))
	expiredJWT := buildJWT(t, "jti-expired", time.Now().Add(-time.Hour))
	jwtNoJTI := buildJWT(t, "", time.Now().Add(time.Hour))
	basicCreds := "Y2lkOnNlYw==" // cid:sec

	tests := []struct {
		name       string
		header     string
		wantToken  string
		wantSecret string
	}{
		{name: "Bearer JWT routes to token only", header: "Bearer " + validJWT, wantToken: validJWT},
		{name: "Bearer PAT routes to token only", header: "Bearer fpt_pat_token", wantToken: "fpt_pat_token"},
		{name: "Basic creds route to secret only", header: "Basic " + basicCreds, wantSecret: basicCreds},
		{name: "Bearer with expired JWT yields nothing", header: "Bearer " + expiredJWT},
		{name: "Bearer with JWT missing JTI yields nothing", header: "Bearer " + jwtNoJTI},
		{name: "Bearer with non-PAT non-JWT garbage yields nothing", header: "Bearer randomjunk"},
		{name: "lowercase bearer scheme accepted", header: "bearer fpt_pat_token", wantToken: "fpt_pat_token"},
		{name: "lowercase basic scheme accepted", header: "basic " + basicCreds, wantSecret: basicCreds},
		{name: "unknown scheme rejected", header: "DPoP fpt_pat_token"},
		{name: "bare PAT without scheme rejected", header: "fpt_pat_token"},
		{name: "bare creds without scheme rejected", header: basicCreds},
		{name: "empty header rejected", header: ""},
	}

	transports := []struct {
		name    string
		capture func(*testing.T, *SessionInterceptor, http.Header) metadata.MD
	}{
		{name: "WrapUnary", capture: captureMetadataUnary},
		{name: "WrapStreamingHandler", capture: captureMetadataStreaming},
		{name: "UnaryConnectRequestHeadersAnnotator", capture: captureMetadataAnnotator},
	}

	for _, transport := range transports {
		for _, tc := range tests {
			t.Run(transport.name+"/"+tc.name, func(t *testing.T) {
				h := http.Header{}
				if tc.header != "" {
					h.Set("Authorization", tc.header)
				}
				md := transport.capture(t, s, h)

				ctx := metadata.NewIncomingContext(context.Background(), md)
				gotToken, hasToken := authenticate.GetTokenFromContext(ctx)
				gotSecret, hasSecret := authenticate.GetSecretFromContext(ctx)

				if tc.wantToken == "" {
					assert.False(t, hasToken, "expected no token via GetTokenFromContext")
					assert.Empty(t, gotToken)
				} else {
					assert.True(t, hasToken, "expected token via GetTokenFromContext")
					assert.Equal(t, tc.wantToken, gotToken)
				}

				if tc.wantSecret == "" {
					assert.False(t, hasSecret, "expected no secret via GetSecretFromContext")
					assert.Empty(t, gotSecret)
				} else {
					assert.True(t, hasSecret, "expected secret via GetSecretFromContext")
					assert.Equal(t, tc.wantSecret, gotSecret)
				}

				// Cross-slot regression: Bearer must never populate the secret slot,
				// and Basic must never populate the token slot. This was the central
				// bug fixed by scheme-routed Authorization parsing.
				if tc.wantToken != "" {
					assert.False(t, hasSecret, "Bearer must not populate the secret slot")
				}
				if tc.wantSecret != "" {
					assert.False(t, hasToken, "Basic must not populate the token slot")
				}
			})
		}
	}
}

// TestSessionInterceptor_XUserTokenHeader_StillWorks ensures the existing X-User-Token
// header path (consts.UserTokenRequestKey) is untouched by this refactor.
func TestSessionInterceptor_XUserTokenHeader_StillWorks(t *testing.T) {
	s := newTestSessionInterceptor()
	h := http.Header{}
	h.Set(consts.UserTokenRequestKey, "explicit-token")

	md := captureMetadataUnary(t, s, h)
	ctx := metadata.NewIncomingContext(context.Background(), md)
	tok, ok := authenticate.GetTokenFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, "explicit-token", tok)
}

// TestSessionInterceptor_XUserTokenAndAuthorization_CoexistWithoutClobber confirms that
// when both X-User-Token and Authorization are present, the Authorization value (if valid)
// overrides — preserving prior behavior. The header is processed *after* the X-User-Token
// branch, so the same order must hold.
func TestSessionInterceptor_XUserTokenAndAuthorization_CoexistWithoutClobber(t *testing.T) {
	s := newTestSessionInterceptor()
	validJWT := buildJWT(t, "jti-priority", time.Now().Add(time.Hour))
	h := http.Header{}
	h.Set(consts.UserTokenRequestKey, "explicit-token")
	h.Set("Authorization", "Bearer "+validJWT)

	md := captureMetadataUnary(t, s, h)
	ctx := metadata.NewIncomingContext(context.Background(), md)
	tok, ok := authenticate.GetTokenFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, validJWT, tok, "Authorization header should win when both are set")
}

// TestSessionInterceptor_NoAuthHeaders_ProducesEmptyMetadata ensures the interceptor
// produces clean incoming metadata when nothing auth-related is present.
func TestSessionInterceptor_NoAuthHeaders_ProducesEmptyMetadata(t *testing.T) {
	s := newTestSessionInterceptor()

	md := captureMetadataUnary(t, s, http.Header{})
	ctx := metadata.NewIncomingContext(context.Background(), md)
	_, hasToken := authenticate.GetTokenFromContext(ctx)
	_, hasSecret := authenticate.GetSecretFromContext(ctx)
	assert.False(t, hasToken)
	assert.False(t, hasSecret)
}

// TestSessionInterceptor_BearerAndBasic_BothPresent_BearerWins documents the behavior when
// a request carries two Authorization headers — only the first is considered (matches
// Values()[0] read), and Bearer wins if it comes first.
func TestSessionInterceptor_BearerAndBasic_BothPresent_BearerWins(t *testing.T) {
	s := newTestSessionInterceptor()
	h := http.Header{}
	h.Add("Authorization", "Bearer fpt_first")
	h.Add("Authorization", "Basic Y2lkOnNlYw==")

	md := captureMetadataUnary(t, s, h)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	tok, hasToken := authenticate.GetTokenFromContext(ctx)
	_, hasSecret := authenticate.GetSecretFromContext(ctx)
	assert.True(t, hasToken)
	assert.Equal(t, "fpt_first", tok)
	assert.False(t, hasSecret, "second Authorization header value must be ignored")
}
