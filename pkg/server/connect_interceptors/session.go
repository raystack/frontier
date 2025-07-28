package connectinterceptors

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/raystack/frontier/core/authenticate"

	"github.com/raystack/frontier/internal/api/v1beta1connect"
	"github.com/raystack/frontier/pkg/server/consts"

	"connectrpc.com/connect"
	"github.com/gorilla/securecookie"
	"google.golang.org/grpc/metadata"
)

type SessionInterceptor struct {
	// TODO(kushsharma): server should be able to rotate encryption keys of codec
	// use secure cookie EncodeMulti/DecodeMulti
	cookieCodec securecookie.Codec
	conf        authenticate.SessionConfig
	h           *v1beta1connect.ConnectHandler
}

func (s *SessionInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return connect.StreamingClientFunc(func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		conn := next(ctx, spec)
		return conn
	})
}

func (s *SessionInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return connect.StreamingHandlerFunc(func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		// parse and process cookies
		incomingMD := metadata.MD{}
		if s.cookieCodec != nil {
			if mdCookies := conn.RequestHeader().Get("cookie"); len(mdCookies) > 0 {
				header := http.Header{}
				header.Add("Cookie", mdCookies)
				request := http.Request{Header: header}
				for _, requestCookie := range request.Cookies() {
					// check if cookie is session cookie
					if requestCookie.Name == consts.SessionRequestKey {
						var sessionID string
						// extract and decode session from cookie
						if err := s.cookieCodec.Decode(requestCookie.Name, requestCookie.Value, &sessionID); err == nil {
							// pass cookie in context
							incomingMD.Set(consts.SessionIDGatewayKey, strings.TrimSpace(sessionID))
						}
					}
				}
			}
		}

		// pass user token if in token header as gateway context
		if userToken := conn.RequestHeader().Values(consts.UserTokenRequestKey); len(userToken) > 0 {
			incomingMD.Set(consts.UserTokenGatewayKey, strings.TrimSpace(userToken[0]))
		}
		// check if the same token is part of Authorization header
		if authHeader := conn.RequestHeader().Values("authorization"); len(authHeader) > 0 {
			tokenVal := strings.TrimSpace(strings.TrimPrefix(authHeader[0], "Bearer "))
			if token, err := jwt.ParseInsecure([]byte(tokenVal)); err == nil {
				if token.JwtID() != "" && token.Expiration().After(time.Now().UTC()) {
					incomingMD.Set(consts.UserTokenGatewayKey, tokenVal)
				}
			}
			secretVal := strings.TrimSpace(strings.TrimPrefix(authHeader[0], "Basic "))
			if len(secretVal) > 0 {
				incomingMD.Set(consts.UserSecretGatewayKey, secretVal)
			}
		}

		ctx = metadata.NewIncomingContext(ctx, incomingMD)
		return next(ctx, conn)
	})
}

func (s *SessionInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (resp connect.AnyResponse, err error) {
		// parse and process cookies
		incomingMD := metadata.MD{}
		if s.cookieCodec != nil {
			if mdCookies := req.Header().Get("cookie"); len(mdCookies) > 0 {
				header := http.Header{}
				header.Add("Cookie", mdCookies)
				request := http.Request{Header: header}
				for _, requestCookie := range request.Cookies() {
					// check if cookie is session cookie
					if requestCookie.Name == consts.SessionRequestKey {
						var sessionID string
						// extract and decode session from cookie
						if err = s.cookieCodec.Decode(requestCookie.Name, requestCookie.Value, &sessionID); err == nil {
							// pass cookie in context
							incomingMD.Set(consts.SessionIDGatewayKey, strings.TrimSpace(sessionID))
						}
					}
				}
			}
		}

		// pass user token if in token header as gateway context
		if userToken := req.Header().Values(consts.UserTokenRequestKey); len(userToken) > 0 {
			incomingMD.Set(consts.UserTokenGatewayKey, strings.TrimSpace(userToken[0]))
		}
		// check if the same token is part of Authorization header
		if authHeader := req.Header().Values("authorization"); len(authHeader) > 0 {
			tokenVal := strings.TrimSpace(strings.TrimPrefix(authHeader[0], "Bearer "))
			if token, err := jwt.ParseInsecure([]byte(tokenVal)); err == nil {
				if token.JwtID() != "" && token.Expiration().After(time.Now().UTC()) {
					incomingMD.Set(consts.UserTokenGatewayKey, tokenVal)
				}
			}
			secretVal := strings.TrimSpace(strings.TrimPrefix(authHeader[0], "Basic "))
			if len(secretVal) > 0 {
				incomingMD.Set(consts.UserSecretGatewayKey, secretVal)
			}
		}

		ctx = metadata.NewIncomingContext(ctx, incomingMD)
		return next(ctx, req)
	})
}

// UnaryConnectResponseInterceptor adds session cookie to response if session id is present in header
func (s SessionInterceptor) UnaryConnectResponseInterceptor() connect.UnaryInterceptorFunc {
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			// Let the handler and other interceptors run first
			resp, err := next(ctx, req)
			if err != nil {
				return nil, err
			}

			// Only attempt to set cookie if we have a codec and session ID
			if sessionID := resp.Header().Get(consts.SessionIDGatewayKey); sessionID != "" && s.cookieCodec != nil {
				// encode session id into cookie
				encodedSession, err := s.cookieCodec.Encode(consts.SessionRequestKey, sessionID)
				if err != nil {
					return nil, connect.NewError(connect.CodeInternal, err)
				}

				// set cookie in response with all required attributes
				cookie := fmt.Sprintf("%s=%s; Path=/; Domain=%s; HttpOnly; SameSite=%v; Secure",
					consts.SessionRequestKey,
					encodedSession,
					s.conf.Domain,
					CookieSameSite(s.conf.SameSite))
				resp.Header().Set("Set-Cookie", cookie)

				// delete the gateway headers to not expose any grpc-metadata in http response
				resp.Header().Del(consts.SessionIDGatewayKey)
				resp.Header().Del("grpc-metadata-" + consts.SessionIDGatewayKey)
			}

			// Check if we need to delete the session cookie (after any set operations)
			if deleteSession := resp.Header().Get(consts.SessionDeleteGatewayKey); deleteSession == "true" {
				// Remove the gateway header
				resp.Header().Del(consts.SessionDeleteGatewayKey)

				// Set an expired cookie to clear it
				cookie := fmt.Sprintf("%s=; Path=/; Domain=%s; Expires=%s; MaxAge=-1; HttpOnly; SameSite=%v; Secure",
					consts.SessionRequestKey,
					s.conf.Domain,
					time.Now().UTC().Format(time.RFC1123),
					CookieSameSite(s.conf.SameSite))
				resp.Header().Set("Set-Cookie", cookie)
			}

			// did the gRPC method set location redirect key in metadata?
			locationGatewayHeaders := req.Header().Values(consts.LocationGatewayKey)
			if len(locationGatewayHeaders) == 1 && len(locationGatewayHeaders[0]) > 0 {
				// delete the gateway headers to not expose any grpc-metadata in http response
				resp.Header().Del((consts.LocationGatewayKey))
				resp.Header().Del("grpc-metadata-" + consts.LocationGatewayKey)

				resp.Header().Set(consts.LocationRequestKey, locationGatewayHeaders[0])
			}

			return resp, nil
		})
	}
	return connect.UnaryInterceptorFunc(interceptor)
}

func NewSessionInterceptor(cookieCutter securecookie.Codec, conf authenticate.SessionConfig, h *v1beta1connect.ConnectHandler) *SessionInterceptor {
	return &SessionInterceptor{
		// could be nil if not configured by user
		cookieCodec: cookieCutter,
		conf:        conf,
		h:           h,
	}
}

func CookieSameSite(name string) http.SameSite {
	switch strings.ToLower(name) {
	case "lax":
		return http.SameSiteLaxMode
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteDefaultMode
	}
}

// UnaryConnectRequestHeadersAnnotator converts session cookies set in grpc metadata to context
// this requires decrypting the cookie and setting it as context
func (s SessionInterceptor) UnaryConnectRequestHeadersAnnotator() connect.UnaryInterceptorFunc {
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (resp connect.AnyResponse, err error) {
			// parse and process cookies
			incomingMD := metadata.MD{}
			if s.cookieCodec != nil {
				if mdCookies := req.Header().Get("cookie"); len(mdCookies) > 0 {
					header := http.Header{}
					header.Add("Cookie", mdCookies)
					request := http.Request{Header: header}
					for _, requestCookie := range request.Cookies() {
						// check if cookie is session cookie
						if requestCookie.Name == consts.SessionRequestKey {
							var sessionID string
							// extract and decode session from cookie
							if err = s.cookieCodec.Decode(requestCookie.Name, requestCookie.Value, &sessionID); err == nil {
								// pass cookie in context
								incomingMD.Set(consts.SessionIDGatewayKey, strings.TrimSpace(sessionID))
							}
						}
					}
				}
			}

			// pass user token if in token header as gateway context
			if userToken := req.Header().Values(consts.UserTokenRequestKey); len(userToken) > 0 {
				incomingMD.Set(consts.UserTokenGatewayKey, strings.TrimSpace(userToken[0]))
			}
			// check if the same token is part of Authorization header
			if authHeader := req.Header().Values("authorization"); len(authHeader) > 0 {
				tokenVal := strings.TrimSpace(strings.TrimPrefix(authHeader[0], "Bearer "))
				if token, err := jwt.ParseInsecure([]byte(tokenVal)); err == nil {
					if token.JwtID() != "" && token.Expiration().After(time.Now().UTC()) {
						incomingMD.Set(consts.UserTokenGatewayKey, tokenVal)
					}
				}
				secretVal := strings.TrimSpace(strings.TrimPrefix(authHeader[0], "Basic "))
				if len(secretVal) > 0 {
					incomingMD.Set(consts.UserSecretGatewayKey, secretVal)
				}
			}

			ctx = metadata.NewIncomingContext(ctx, incomingMD)
			return next(ctx, req)
		})
	}
	return connect.UnaryInterceptorFunc(interceptor)
}
