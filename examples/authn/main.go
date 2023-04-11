package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

const (
	shieldRouteListConn    = "/admin/v1beta1/auth"
	shieldRegister         = "/admin/v1beta1/auth/register"
	shieldRegisterCallback = "/admin/v1beta1/auth/callback"
	shieldLogout           = "/admin/v1beta1/auth/logout"
	shieldUserProfile      = "/admin/v1beta1/users/self"
)

var (
	// shieldHost which is running locally and configured with oidc parameters
	// it should have client id, secret, issuer and an oidc callback endpoint
	// for this example we are using ourselves as a frontend to shield backend
	shieldHost = "http://localhost:7400"
	appHost    = "localhost:8888"

	returnAfterAuthURL = url.QueryEscape("http://" + appHost + "/profile")
)

func main() {
	flag.StringVar(&shieldHost, "shieldhost", shieldHost, "shield host endpoint, e.g. http://localhost:7400")
	flag.StringVar(&appHost, "apphost", appHost, "app host, e.g. localhost:8888")
	flag.StringVar(&returnAfterAuthURL, "returnto", returnAfterAuthURL, "where should shield return the call after successful auth, e.g. http://localhost:8888/profile")
	flag.Parse()

	r := gin.Default()
	r.LoadHTMLFiles("static/index.html")
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.GET("/", home())
	r.GET("/login", login())
	r.GET("/auth", auth())
	r.GET("/callback", callback())
	r.GET("/logout", logout())
	r.GET("/profile", profile())
	r.GET("/test", test())
	r.Run(appHost) // listen and serve
}

func test() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"headers": ctx.Request.Header,
		})
	}
}

func home() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"title":   "Demo",
			"page":    "Home",
			"content": "Welcome to authentication demo",
		})
	}
}

func login() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		shieldResp, err := http.Get(shieldHost + shieldRouteListConn)
		if err != nil {
			ctx.Error(err)
			return
		}
		defer shieldResp.Body.Close()

		type Response struct {
			Strategies []struct {
				Name   string `json:"name"`
				Params any
			} `json:"strategies"`
		}
		var response Response
		if err = json.NewDecoder(shieldResp.Body).Decode(&response); err != nil {
			ctx.Error(err)
			return
		}

		content := `<div><h3>Supported Providers:</h3>`
		for _, strategy := range response.Strategies {
			content += `<div><a href="/auth?strategy=` + strategy.Name
			//content += `?redirect=1`
			//content += `&return_to=` + returnAfterAuthURL
			content += `">` + strategy.Name + `</a></div>`
		}
		content += `</div>`
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"title":   "Authentication demo",
			"page":    "Login/Register",
			"content": template.HTML(content),
		})
	}
}

func auth() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		authStrategy := ctx.Query("strategy")
		if len(authStrategy) == 0 {
			ctx.Redirect(http.StatusSeeOther, "/login")
		}
		shieldURL, _ := url.JoinPath(shieldHost, shieldRegister, authStrategy)
		shieldResp, err := http.Get(shieldURL)
		if err != nil {
			ctx.Error(err)
			return
		}
		defer shieldResp.Body.Close()

		type Response struct {
			Endpoint string `json:"endpoint"`
		}
		var response Response
		if err = json.NewDecoder(shieldResp.Body).Decode(&response); err != nil {
			ctx.Error(err)
			return
		}
		ctx.Redirect(http.StatusSeeOther, response.Endpoint)
	}
}

func callback() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		// build a pass through request to shield
		req, err := http.NewRequest(http.MethodGet, shieldHost+shieldRegisterCallback, nil)
		if err != nil {
			ctx.Error(err)
			return
		}
		req.URL.RawQuery = ctx.Request.URL.RawQuery
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			ctx.Error(err)
			return
		}

		// parse & verify jwt with shield public keys
		userToken := resp.Header.Get("x-user-token")
		jwks, err := jwk.Fetch(
			ctx,
			shieldHost+"/jwks.json",
		)
		if err != nil {
			ctx.Error(err)
			return
		}
		verifiedToken, err := jwt.Parse([]byte(userToken), jwt.WithKeySet(jwks))
		if err != nil {
			fmt.Printf("failed to verify JWS: %s\n", err)
			return
		}
		tokenClaims, err := verifiedToken.AsMap(ctx)
		if err != nil {
			ctx.Error(err)
			return
		}

		// clone response headers for cookie
		ctx.Writer.Header().Set("set-cookie", resp.Header.Get("set-cookie"))

		// render token
		tokenHTML := "<article>" + userToken + "</article><article>" + fmt.Sprintf("%v", tokenClaims) + "</article>"
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"title":   "Authentication demo",
			"page":    "Callback",
			"content": "Callback successful for auth. Check profile section now.",
			"token":   template.HTML(tokenHTML),
		})
	}
}

func logout() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		req, err := http.NewRequest(http.MethodGet, shieldHost+shieldLogout, nil)
		if err != nil {
			ctx.Error(err)
			return
		}
		// set cookie
		for _, c := range ctx.Request.Cookies() {
			req.AddCookie(c)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			ctx.Error(err)
			return
		}

		// clone response headers for cookie
		ctx.Writer.Header().Set("set-cookie", resp.Header.Get("set-cookie"))
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"title":   "Authentication demo",
			"page":    "Logout",
			"content": "You are logged out",
		})
	}
}

func profile() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		req, err := http.NewRequest(http.MethodGet, shieldHost+shieldUserProfile, nil)
		if err != nil {
			ctx.Error(err)
			return
		}
		// set cookie
		for _, c := range ctx.Request.Cookies() {
			req.AddCookie(c)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			ctx.Error(err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			// fail early
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"title":   "Authentication demo",
				"page":    "Profile",
				"content": "Please login to fetch profile",
			})
			return
		}

		// deserialize response
		type Response struct {
			User struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			} `json:"user"`
		}
		var response Response
		if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
			ctx.Error(err)
			return
		}

		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"title":   "Authentication demo",
			"page":    "Profile",
			"content": "Hello " + response.User.Email + ", you are logged in!",
		})
	}
}
