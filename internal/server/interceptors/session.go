package interceptors

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/raystack/shield/internal/server/consts"

	"github.com/gorilla/securecookie"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

type Session struct {
	// TODO(kushsharma): server should be able to rotate encryption keys of codec
	// use secure cookie EncodeMulti/DecodeMulti
	cookieCodec securecookie.Codec
}

func NewSession(cookieCutter securecookie.Codec) *Session {
	return &Session{cookieCodec: cookieCutter}
}

// GatewayResponseModifier https://grpc-ecosystem.github.io/grpc-gateway/docs/mapping/customizing_your_gateway/
// called just before RPC server response gets serialized for gateway
func (h Session) GatewayResponseModifier(ctx context.Context, w http.ResponseWriter, _ proto.Message) error {
	md, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		return fmt.Errorf("failed to extract ServerMetadata from context")
	}

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
				Name:     consts.SessionRequestKey,
				Value:    encoded,
				Path:     "/",
				Expires:  time.Now().UTC().Add(consts.SessionValidity),
				MaxAge:   86400 * 30, // 30 days
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})
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
			Name:     consts.SessionRequestKey,
			Value:    "",
			Path:     "/",
			Expires:  time.Now().UTC(),
			MaxAge:   -1,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			//Secure:   true,
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

// GatewayRequestMetadataAnnotator look up session header and pass it as context if it exists
// called just before RPC server side execution
func (h Session) GatewayRequestMetadataAnnotator(_ context.Context, r *http.Request) metadata.MD {
	mdMap := map[string]string{}

	// extract cookie and pass it as context
	requestCookie, err := r.Cookie(consts.SessionRequestKey)
	if err == nil && requestCookie.Valid() == nil {
		var sessionID string
		if err = h.cookieCodec.Decode(requestCookie.Name, requestCookie.Value, &sessionID); err == nil {
			mdMap[consts.SessionIDGatewayKey] = sessionID
		}
	}

	// TODO(kushsharma): pass `Refer` header as context value and use it as `redirect_to` field
	// if not provided during registration flow

	return metadata.New(mdMap)
}
