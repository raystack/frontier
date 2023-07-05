package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"html/template"
	"io"
	"net/http"
	"net/url"
)

const (
	shieldRouteListConn    = "/v1beta1/auth"
	shieldRegister         = "/v1beta1/auth/register"
	shieldRegisterCallback = "/v1beta1/auth/callback"
	shieldLogout           = "/v1beta1/auth/logout"
	shieldAccessToken      = "/v1beta1/auth/token"
	shieldUserProfile      = "/v1beta1/users/self" // protected endpoint
	jwksPath               = "/.well-known/jwks.json"

	mailotpStrategy = "mailotp"
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
	r.GET("/oauth", oauth())
	r.GET("/mailauth", mailauth())
	r.GET("/token", token())
	r.POST("/token", token())
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
		if shieldResp.StatusCode != http.StatusOK {
			ctx.Error(fmt.Errorf("shield returned status code %d", shieldResp.StatusCode))
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
			if strategy.Name == mailotpStrategy {
				content += `<article>`
				content += `<form action="/mailauth" method="get">`
				content += `<div>` + strategy.Name + `</div>`
				content += `Email: <input type="text" name="email" placeholder="email">`
				content += `<input type="submit" value="Submit">`
				content += `</form>`
				content += `</article>`
			} else {
				content += `<article>`
				content += `<a href="/oauth?strategy=` + strategy.Name
				//content += `?redirect=1`
				//content += `&return_to=` + returnAfterAuthURL
				content += `">` + strategy.Name + `</a>`
				content += "</article>"
			}
		}
		content += `</div>`
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"title":   "Authentication demo",
			"page":    "Login/Register",
			"content": template.HTML(content),
		})
	}
}

func oauth() func(ctx *gin.Context) {
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
		if shieldResp.StatusCode != http.StatusOK {
			ctx.Error(fmt.Errorf("shield returned status code %d", shieldResp.StatusCode))
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

func mailauth() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		userEmail := ctx.Query("email")
		if len(userEmail) == 0 {
			ctx.Redirect(http.StatusSeeOther, "/login")
		}
		shieldURL, _ := url.JoinPath(shieldHost, shieldRegister, mailotpStrategy)
		shieldResp, err := http.Get(shieldURL + "?email=" + userEmail)
		if err != nil {
			ctx.Error(err)
			return
		}
		if shieldResp.StatusCode != http.StatusOK {
			ctx.Error(fmt.Errorf("shield returned status code %d", shieldResp.StatusCode))
			return
		}
		defer shieldResp.Body.Close()

		type Response struct {
			Endpoint string `json:"endpoint"`
			State    string `json:"state"`
		}
		var response Response
		if err = json.NewDecoder(shieldResp.Body).Decode(&response); err != nil {
			ctx.Error(err)
			return
		}
		content := ""
		content += `<div>`
		content += `<form action="/callback" method="get">`
		content += `<div>Enter OTP sent to you email: ` + userEmail + `</div>`
		content += `OTP: <input type="text" name="code" placeholder="otp sent in the mail">`
		content += `State: <input readonly type="text" name="state" value="` + response.State + `">`
		content += `<input type="hidden" name="strategy_name" value="` + mailotpStrategy + `">`
		content += `<input type="submit" value="Submit">`
		content += `</form>`
		content += `</div>`
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"title":   "Authentication demo",
			"page":    "Mail OTP verify",
			"content": template.HTML(content),
		})
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
		if resp.StatusCode != http.StatusOK {
			ctx.Error(fmt.Errorf("shield returned status code %d", resp.StatusCode))
			respBody, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			ctx.Error(fmt.Errorf("shield returned %s", string(respBody)))
			return
		}
		// clone response headers for cookie
		ctx.Writer.Header().Set("set-cookie", resp.Header.Get("set-cookie"))

		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"title":   "Authentication demo",
			"page":    "Callback",
			"content": "Authentication successful. Check profile section now.",
		})
	}
}

func token() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		if ctx.Request.Method == http.MethodGet {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"title": "Authentication demo",
				"page":  "Stateless Profile via Client Credentials",
				"content": template.HTML(`
<div>
	<form action="/token" method="post">
	<input type="hidden" name="grant_type" value="client_credentials">
	<input type="text" name="client_id" placeholder="client id">
	<input type="text" name="client_secret" placeholder="client secret">
	<input type="submit" value="Get Profile">
	</form>
</div>
`),
			})
			return
		}

		type tokenRequest struct {
			GrantType    string `json:"grant_type"`
			ClientID     string `json:"client_id"`
			ClientSecret string `json:"client_secret"`
		}
		bodyBuf := &bytes.Buffer{}
		if err := json.NewEncoder(bodyBuf).Encode(tokenRequest{
			GrantType:    ctx.PostForm("grant_type"),
			ClientID:     ctx.PostForm("client_id"),
			ClientSecret: ctx.PostForm("client_secret"),
		}); err != nil {
			ctx.Error(err)
			return
		}

		// build a pass through request to shield
		req, err := http.NewRequest(http.MethodPost, shieldHost+shieldAccessToken, bodyBuf)
		if err != nil {
			ctx.Error(err)
			return
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			ctx.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			ctx.Error(fmt.Errorf("shield returned status code %d", resp.StatusCode))
			respBody, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			ctx.Error(fmt.Errorf("shield returned %s", string(respBody)))
			return
		}

		// render token
		tokenHTML := "Access token is disabled by auth server"

		// get access token from shield
		var tokenResp struct {
			AccessToken string `json:"access_token"`
			TokenType   string `json:"token_type"`
		}
		if err = json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
			ctx.Error(fmt.Errorf("failed to decode token response: %s", err))
			return
		}

		// parse & verify jwt with shield public keys if provided
		if tokenResp.AccessToken != "" {
			jwks, err := jwk.Fetch(
				ctx,
				shieldHost+jwksPath,
			)
			if err != nil {
				ctx.Error(fmt.Errorf("failed to fetch JWK: %s", err))
				return
			}
			verifiedToken, err := jwt.Parse([]byte(tokenResp.AccessToken), jwt.WithKeySet(jwks))
			if err != nil {
				fmt.Printf("failed to verify JWS: %s\n", err)
				return
			}
			tokenClaims, err := verifiedToken.AsMap(ctx)
			if err != nil {
				ctx.Error(err)
				return
			}
			// token ready to use
			tokenHTML = `Authentication successful via client credentials. 
<h4>User Token:</h4>
<article>` + tokenResp.AccessToken + `</article>
<h4>Claims:</h4>
<article>` + fmt.Sprintf("%v", tokenClaims) + `</article>`
		}

		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"title":   "Authentication demo",
			"page":    "Callback",
			"content": template.HTML(tokenHTML),
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

		content := "Hello <b>" + response.User.Email + "</b>, you are logged in!"
		content += "<article>Cookie: " + ctx.Request.Header.Get("Cookie") + "</article>"
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"title":   "Authentication demo",
			"page":    "Profile",
			"content": template.HTML(content),
		})
	}
}
