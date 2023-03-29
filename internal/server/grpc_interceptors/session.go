package grpc_interceptors

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

const (
	// const keys user to pass values from gRPC methods to http mux interface
	SessionIDGatewayKey     = "gateway-session-id"
	SessionDeleteGatewayKey = "gateway-session-delete"
	UserTokenGatewayKey     = "gateway-user-token"
	LocationGatewayKey      = "gateway-location"

	// UserTokenRequestKey is returned from the application to client containing user details in
	// response headers
	UserTokenRequestKey = "x-user-token"

	// LocationRequestKey is used to set location response header for redirecting browser
	LocationRequestKey = "location"

	// SessionRequestKey is the key to store session value in browser
	SessionRequestKey = "sid"
	// SessionValidity defines the age of a session
	// TODO(kushsharma): should we expose this in config?
	SessionValidity = time.Hour * 24 * 30 // 30 days
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
	sessionGatewayHeaders := md.HeaderMD.Get(SessionIDGatewayKey)
	if len(sessionGatewayHeaders) == 1 && len(sessionGatewayHeaders[0]) > 0 {
		sessionIDFromGateway := sessionGatewayHeaders[0]

		// delete the gateway headers to not expose any grpc-metadata in http response
		md.HeaderMD.Delete(SessionIDGatewayKey)
		w.Header().Del("grpc-metadata-" + SessionIDGatewayKey)

		// put session id in request cookies
		if encoded, err := h.cookieCodec.Encode(SessionRequestKey, sessionIDFromGateway); err == nil {
			http.SetCookie(w, &http.Cookie{
				Name:     SessionRequestKey,
				Value:    encoded,
				Path:     "/",
				Expires:  time.Now().UTC().Add(SessionValidity),
				MaxAge:   86400 * 30, // 30 days
				HttpOnly: true,
				//Secure:   true,
			})
		}
	}

	// did the gRPC method set a session delete key in the metadata?
	sessionDeleteGatewayHeaders := md.HeaderMD.Get(SessionDeleteGatewayKey)
	if len(sessionDeleteGatewayHeaders) == 1 && sessionDeleteGatewayHeaders[0] == "true" {
		// delete the gateway headers to not expose any grpc-metadata in http response
		md.HeaderMD.Delete(SessionDeleteGatewayKey)
		w.Header().Del("grpc-metadata-" + SessionDeleteGatewayKey)

		// clear session from request
		http.SetCookie(w, &http.Cookie{
			Name:     SessionRequestKey,
			Value:    "",
			Path:     "/",
			Expires:  time.Now().UTC(),
			MaxAge:   -1,
			HttpOnly: true,
			//Secure:   true,
		})
	}

	// did the gRPC method set user jwt key in metadata?
	userTokenGatewayHeaders := md.HeaderMD.Get(UserTokenGatewayKey)
	if len(userTokenGatewayHeaders) == 1 && len(userTokenGatewayHeaders[0]) > 0 {
		// delete the gateway headers to not expose any grpc-metadata in http response
		w.Header().Del("grpc-metadata-" + UserTokenGatewayKey)

		w.Header().Set(UserTokenRequestKey, userTokenGatewayHeaders[0])
	}

	// did the gRPC method set location redirect key in metadata?
	locationGatewayHeaders := md.HeaderMD.Get(LocationGatewayKey)
	if len(locationGatewayHeaders) == 1 && len(locationGatewayHeaders[0]) > 0 {
		// delete the gateway headers to not expose any grpc-metadata in http response
		md.HeaderMD.Delete(LocationGatewayKey)
		w.Header().Del("grpc-metadata-" + LocationGatewayKey)

		w.Header().Set(LocationRequestKey, locationGatewayHeaders[0])
		w.WriteHeader(http.StatusSeeOther)
	}
	return nil
}

// GatewayRequestMetadataAnnotator look up session header and pass it as context if it exists
// called just before RPC server side execution
func (h Session) GatewayRequestMetadataAnnotator(_ context.Context, r *http.Request) metadata.MD {
	mdMap := map[string]string{}

	// extract cookie and pass it as context
	requestCookie, err := r.Cookie(SessionRequestKey)
	if err == nil && requestCookie.Valid() == nil {
		var sessionID string
		if err = h.cookieCodec.Decode(requestCookie.Name, requestCookie.Value, &sessionID); err == nil {
			mdMap[SessionIDGatewayKey] = sessionID
		}
	}

	// TODO(kushsharma): pass `Refer` header as context value and use it as `redirect_to` field
	// if not provided during registration flow

	return metadata.New(mdMap)
}
