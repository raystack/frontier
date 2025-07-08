package connectinterceptors

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/raystack/frontier/core/authenticate"

	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/raystack/frontier/pkg/server/consts"

	"connectrpc.com/connect"
	"github.com/gorilla/securecookie"
	"google.golang.org/grpc/metadata"
)

type Session struct {
	// TODO(kushsharma): server should be able to rotate encryption keys of codec
	// use secure cookie EncodeMulti/DecodeMulti
	cookieCodec securecookie.Codec
	conf        authenticate.SessionConfig
}

func NewSession(cookieCutter securecookie.Codec, conf authenticate.SessionConfig) *Session {
	return &Session{
		// could be nil if not configured by user
		cookieCodec: cookieCutter,
		conf:        conf,
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
func (s Session) UnaryConnectRequestHeadersAnnotator() connect.UnaryInterceptorFunc {
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (resp connect.AnyResponse, err error) {
			// parse and process cookies
			incomingMD := metadata.MD{}
			if s.cookieCodec != nil {
				if mdCookies := req.Header().Values("cookie"); len(mdCookies) > 0 && mdCookies[0] != "" {
					header := http.Header{}
					header.Add("Cookie", mdCookies[0])
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
			if userToken := incomingMD.Get(consts.UserTokenRequestKey); len(userToken) > 0 {
				incomingMD.Set(consts.UserTokenGatewayKey, strings.TrimSpace(userToken[0]))
			}
			// check if the same token is part of Authorization header
			if authHeader := incomingMD.Get("authorization"); len(authHeader) > 0 {
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
