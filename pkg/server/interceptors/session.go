package interceptors

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/raystack/frontier/core/authenticate"

	"github.com/lestrrat-go/jwx/v2/jwt"

	"google.golang.org/grpc"

	"github.com/raystack/frontier/pkg/server/consts"

	"github.com/gorilla/securecookie"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
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

// GatewayResponseModifier https://grpc-ecosystem.github.io/grpc-gateway/docs/mapping/customizing_your_gateway/
// called just before RPC server response gets serialized for gateway
func (h Session) GatewayResponseModifier(ctx context.Context, w http.ResponseWriter, _ proto.Message) error {
	md, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		return fmt.Errorf("failed to extract ServerMetadata from context")
	}

	if h.cookieCodec != nil {
		// did the gRPC method set a session ID in the metadata?
		sessionGatewayHeaders := md.HeaderMD.Get(consts.SessionIDGatewayKey)
		if len(sessionGatewayHeaders) == 1 && len(sessionGatewayHeaders[0]) > 0 {
			sessionIDFromGateway := sessionGatewayHeaders[0]

			// delete the gateway headers to not expose any grpc-metadata in http response
			md.HeaderMD.Delete(consts.SessionIDGatewayKey)
			w.Header().Del("grpc-metadata-" + consts.SessionIDGatewayKey)

			// put session id in request cookies
			if encoded, err := h.cookieCodec.Encode(consts.SessionRequestKey, sessionIDFromGateway); err == nil {
				http.SetCookie(w, &http.Cookie{
					Domain:   h.conf.Domain,
					Name:     consts.SessionRequestKey,
					Value:    encoded,
					Path:     "/",
					Expires:  time.Now().UTC().Add(h.conf.Validity),
					MaxAge:   int(h.conf.Validity.Seconds()),
					HttpOnly: true,
					SameSite: CookieSameSite(h.conf.SameSite),
					Secure:   h.conf.Secure,
				})
			}
		}
	}

	// did the gRPC method set a session delete key in the metadata?
	sessionDeleteGatewayHeaders := md.HeaderMD.Get(consts.SessionDeleteGatewayKey)
	if len(sessionDeleteGatewayHeaders) == 1 && sessionDeleteGatewayHeaders[0] == "true" {
		// delete the gateway headers to not expose any grpc-metadata in http response
		md.HeaderMD.Delete(consts.SessionDeleteGatewayKey)
		w.Header().Del("grpc-metadata-" + consts.SessionDeleteGatewayKey)

		// clear session from request
		http.SetCookie(w, &http.Cookie{
			Domain:   h.conf.Domain,
			Name:     consts.SessionRequestKey,
			Value:    "",
			Path:     "/",
			Expires:  time.Now().UTC(),
			MaxAge:   -1,
			HttpOnly: true,
			SameSite: CookieSameSite(h.conf.SameSite),
			Secure:   h.conf.Secure,
		})
	}

	// did the gRPC method set user jwt key in metadata?
	userTokenGatewayHeaders := md.HeaderMD.Get(consts.UserTokenGatewayKey)
	if len(userTokenGatewayHeaders) == 1 && len(userTokenGatewayHeaders[0]) > 0 {
		// delete the gateway headers to not expose any grpc-metadata in http response
		w.Header().Del("grpc-metadata-" + consts.UserTokenGatewayKey)

		w.Header().Set(consts.UserTokenRequestKey, userTokenGatewayHeaders[0])
	}

	// did the gRPC method set location redirect key in metadata?
	locationGatewayHeaders := md.HeaderMD.Get(consts.LocationGatewayKey)
	if len(locationGatewayHeaders) == 1 && len(locationGatewayHeaders[0]) > 0 {
		// delete the gateway headers to not expose any grpc-metadata in http response
		md.HeaderMD.Delete(consts.LocationGatewayKey)
		w.Header().Del("grpc-metadata-" + consts.LocationGatewayKey)

		w.Header().Set(consts.LocationRequestKey, locationGatewayHeaders[0])
		w.WriteHeader(http.StatusSeeOther)
	}
	return nil
}

// UnaryGRPCRequestHeadersAnnotator converts session cookies set in grpc metadata to context
// this requires decrypting the cookie and setting it as context
func (h Session) UnaryGRPCRequestHeadersAnnotator() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// parse and process cookies
		if incomingMD, ok := metadata.FromIncomingContext(ctx); ok {
			if h.cookieCodec != nil {
				if mdCookies := incomingMD.Get("cookie"); len(mdCookies) > 0 {
					header := http.Header{}
					header.Add("Cookie", mdCookies[0])
					request := http.Request{Header: header}
					for _, requestCookie := range request.Cookies() {
						// check if cookie is session cookie
						if requestCookie.Name == consts.SessionRequestKey {
							var sessionID string
							// extract and decode session from cookie
							if err = h.cookieCodec.Decode(requestCookie.Name, requestCookie.Value, &sessionID); err == nil {
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
		}
		return handler(ctx, req)
	}
}
