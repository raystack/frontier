package basic_auth

import (
	"bytes"
	"net/http"
	"regexp"
	"strings"
	"text/template"

	"github.com/raystack/frontier/internal/proxy/middleware"
	"github.com/raystack/frontier/pkg/body_extractor"
	"github.com/raystack/frontier/pkg/httputil"

	goauth "github.com/abbot/go-http-auth"
	"github.com/mitchellh/mapstructure"
	"github.com/raystack/salt/log"
)

const (
	RegexPrefix = "r#"
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

	// TODO take a file instead of embedded users in yaml
	UserDB string `yaml:"userdb" mapstructure:"userdb"`

	// Scope is optional and used for additional policy based
	// authorization over user
	Scope Scope `yaml:"scope" mapstructure:"scope"`
}

type Credentials struct {
	User string `yaml:"user" mapstructure:"user"`

	// Password must be hashed using MD5, SHA1, or BCrypt(recommended) using htpasswd
	Password string `yaml:"password" mapstructure:"password"`

	// Capabilities are optional and used with scope for applying authz
	// Supports regular expr if marked with r# as a prefix
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

func (w BasicAuth) Info() *middleware.MiddlewareInfo {
	return &middleware.MiddlewareInfo{
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
	authenticator := goauth.NewBasicAuthenticator("frontier", func(user, realm string) string {
		for _, credential := range conf.Users {
			if credential.User == user {
				return credential.Password
			}
		}
		return ""
	})

	var authedUser string
	if authedUser = authenticator.CheckAuth(req); authedUser != "" {
		req.Header.Set(httputil.HeaderXUser, authedUser)
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
			payloadField, err := body_extractor.GRPCPayloadHandler{}.Extract(&req.Body, attr.Index)
			if err != nil {
				w.log.Error("middleware: failed to parse grpc payload", "err", err)
				return false
			}

			templateMap[res] = payloadField
			w.log.Debug("middleware: extracted", "field", payloadField, "attr", attr)
		case middleware.AttributeTypeJSONPayload:
			if attr.Key == "" {
				w.log.Error("middleware: payload key field empty")
				return false
			}
			payloadField, err := body_extractor.JSONPayloadHandler{}.Extract(&req.Body, attr.Key)
			if err != nil {
				w.log.Error("middleware: failed to parse json payload", "err", err)
				return false
			}

			templateMap[res] = payloadField
			w.log.Debug("middleware: extracted", "field", payloadField, "attr", attr)
		default:
			w.log.Error("middleware: unknown attribute type", "attr", attr)
			return false
		}
	}

	var isAllowed = false
	compiledAction, err := CompileString(conf.Scope.Action, templateMap)
	if err != nil {
		w.log.Error("middleware: action parsing failed", "err", err)
		return false
	}
	for _, userCap := range userCapabilities {
		if w.matchAction(userCap, compiledAction) {
			isAllowed = true
			break
		}
	}
	return isAllowed
}

func (w BasicAuth) matchAction(cap, action string) bool {
	// do regex compare if required
	if strings.HasPrefix(cap, RegexPrefix) {
		cap = strings.TrimPrefix(cap, RegexPrefix)
		rxAction, err := regexp.Compile(cap)
		if err != nil {
			w.log.Warn("failed to compile regex", "exp", cap, "err", err)
			return false
		}
		if rxAction.MatchString(action) {
			return true
		}
	}
	return cap == action
}

func CompileString(input string, context map[string]interface{}) (string, error) {
	// note: template can be cached
	tmpl, err := template.New("frontier_engine").Parse(input)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, context); err != nil {
		return "", err
	}
	return strings.TrimSpace(buf.String()), nil
}
