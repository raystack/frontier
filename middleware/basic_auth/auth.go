package basic_auth

import (
	"bytes"
	"net/http"
	"strings"
	"text/template"

	goauth "github.com/abbot/go-http-auth"
	"github.com/mitchellh/mapstructure"
	"github.com/odpf/salt/log"
	"github.com/odpf/shield/middleware"
	"github.com/odpf/shield/structs"
)

// BasicAuth make sure the request is allowed to be sent to backend
// Use apache htpasswd utility for generating passwords
// e.g. htpasswd -nB user
// Middleware will look for Authorization header for credentials
// value should be "Basic <base64encoded user:password>"
type BasicAuth struct {
	log  log.Logger
	next http.Handler
}

type Config struct {
	Users []Credentials `yaml:"users" mapstructure:"users"`

	// Scope is optional and used for additional policy based
	// authorization over user
	Scope Scope `yaml:"scope" mapstructure:"scope"`
}

type Credentials struct {
	User string `yaml:"user" mapstructure:"user"`

	// Password must be hashed using MD5, SHA1, or BCrypt(recommended) using htpasswd
	Password string `yaml:"password" mapstructure:"password"`

	// Capabilities are optional and used with scope for applying authz
	Capabilities []string `yaml:"capabilities" mapstructure:"capabilities"`
}

type Scope struct {
	Action     string                          `yaml:"action" mapstructure:"action"`
	Attributes map[string]middleware.Attribute `yaml:"attributes" mapstructure:"attributes"` // auth field -> Attribute
}

func New(logger log.Logger, next http.Handler) *BasicAuth {
	return &BasicAuth{
		log:  logger,
		next: next,
	}
}

func (w BasicAuth) Info() *structs.MiddlewareInfo {
	return &structs.MiddlewareInfo{
		Name:        "basic_auth",
		Description: "basic username password authentication",
	}
}

func (w *BasicAuth) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	wareSpec, ok := middleware.ExtractMiddleware(req, w.Info().Name)
	if !ok {
		w.next.ServeHTTP(rw, req)
		return
	}

	// note: might effect performance, consider caching
	conf := Config{}
	if err := mapstructure.Decode(wareSpec.Config, &conf); err != nil {
		w.log.Error("middleware: invalid config", "config", wareSpec.Config)
		w.notAllowed(rw)
		return
	}
	authenticator := goauth.NewBasicAuthenticator("shield", func(user, realm string) string {
		for _, credential := range conf.Users {
			if credential.User == user {
				return credential.Password
			}
		}
		return ""
	})

	var authedUser string
	if authedUser = authenticator.CheckAuth(req); authedUser != "" {
		req.Header.Set("X-User", authedUser)
	} else {
		w.notAllowed(rw)
		return
	}

	if conf.Scope.Action != "" {
		// basic authorization
		if !w.authorizeRequest(conf, authedUser, req) {
			w.notAllowed(rw)
			return
		}
	}

	w.next.ServeHTTP(rw, req)
}

func (w BasicAuth) notAllowed(rw http.ResponseWriter) {
	rw.WriteHeader(http.StatusUnauthorized)
	return
}

func (w BasicAuth) authorizeRequest(conf Config, user string, req *http.Request) bool {
	var userCapabilities []string
	for _, u := range conf.Users {
		if u.User == user {
			userCapabilities = u.Capabilities
		}
	}
	if len(userCapabilities) == 0 {
		return false
	}
	// check if its superuser
	for _, cap := range userCapabilities {
		if cap == "*" {
			return true
		}
	}

	templateMap := map[string]interface{}{}
	for res, attr := range conf.Scope.Attributes {
		templateMap[res] = ""

		switch attr.Type {
		case middleware.AttributeTypeGRPCPayload:
			// check if grpc request
			if !strings.HasPrefix(req.Header.Get("Content-Type"), "application/grpc") {
				w.log.Error("middleware: not a valid grpc request")
				return false
			}

			// TODO: we can optimise this by parsing all field at once
			payloadField, err := middleware.GRPCPayloadHandler{}.Extract(req, attr.Index)
			if err != nil {
				w.log.Error("middleware: failed to parse grpc payload", "err", err)
				return false
			}

			templateMap[res] = payloadField
			w.log.Info("middleware: extracted", "field", payloadField, "attr", attr)
		case middleware.AttributeTypeJSONPayload:
			if attr.Key == "" {
				w.log.Error("middleware: payload key field empty")
				return false
			}
			payloadField, err := middleware.JSONPayloadHandler{}.Extract(req, attr.Key)
			if err != nil {
				w.log.Error("middleware: failed to parse grpc payload", "err", err)
				return false
			}

			templateMap[res] = payloadField
			w.log.Info("middleware: extracted", "field", payloadField, "attr", attr)
		default:
			w.log.Error("middleware: unknown attribute type", "attr", attr)
			return false
		}
	}

	compiledAction, err := CompileString(conf.Scope.Action, templateMap)
	if err != nil {
		w.log.Error("middleware: action parsing failed", "err", err)
		return false
	}

	var isAllowed = false
	for _, userCap := range userCapabilities {
		if userCap == compiledAction {
			isAllowed = true
			break
		}
	}
	return isAllowed
}

func CompileString(input string, context map[string]interface{}) (string, error) {
	// note: template can be cached
	tmpl, err := template.New("shield_engine").Parse(input)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, context); err != nil {
		return "", err
	}
	return strings.TrimSpace(buf.String()), nil
}
