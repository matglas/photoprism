package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/photoprism/photoprism/internal/form"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/sirupsen/logrus"
)

// NewApiTest returns new API test helper.
func NewApiTest() (app *gin.Engine, router *gin.RouterGroup, conf *config.Config) {
	conf = service.Config()
	gin.SetMode(gin.TestMode)
	app = gin.New()
	app.LoadHTMLGlob(conf.TemplatesPath() + "/*")
	router = app.Group("/api/v1")
	return app, router, service.Config()
}

// NewApiTestWithOIDC returns new API test helper configured for OIDC.
func NewApiTestWithOIDC() (app *gin.Engine, router *gin.RouterGroup, conf *config.Config) {
	app, router, conf = NewApiTest()
	conf.Options().OidcIssuerUrl = "http://dummy-oidc:9998"
	conf.Options().OidcClientID = "native"
	conf.Options().OidcClientSecret = "random"
	return app, router, conf
}

// AuthenticateAdmin Register session routes and returns valid SessionId.
// Call this func after registering other routes and before performing other requests.
func AuthenticateAdmin(app *gin.Engine, router *gin.RouterGroup) (sessId string) {
	return AuthenticateUser(app, router, "admin", "photoprism")
}

// AuthenticateUser Register session routes and returns valid SessionId.
// Call this func after registering other routes and before performing other requests.
func AuthenticateUser(app *gin.Engine, router *gin.RouterGroup, username string, password string) (sessId string) {
	CreateSession(router)
	f := form.Login{
		UserName: username,
		Password: password,
	}
	loginStr, err := json.Marshal(f)
	if err != nil {
		log.Fatal(err)
	}
	r0 := PerformRequestWithBody(app, http.MethodPost, "/api/v1/session", string(loginStr))
	sessId = r0.Header().Get("X-Session-ID")
	return
}

// Performs API request with empty request body.
// See https://medium.com/@craigchilds94/testing-gin-json-responses-1f258ce3b0b1
func PerformRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// Performs API request with empty request body and Cookie.
func PerformRequestWithCookie(r http.Handler, method, path string, cookie string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	req.Header.Add("Cookie", cookie)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// Performs authenticated API request with empty request body.
func AuthenticatedRequest(r http.Handler, method, path, sess string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	req.Header.Add("X-Session-ID", sess)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// Performs API request including request body as string.
func PerformRequestWithBody(r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	reader := strings.NewReader(body)
	req, _ := http.NewRequest(method, path, reader)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// Performs authenticated API request including request body as string.
func AuthenticatedRequestWithBody(r http.Handler, method, path, body string, sessionId string) *httptest.ResponseRecorder {
	reader := strings.NewReader(body)
	req, _ := http.NewRequest(method, path, reader)
	req.Header.Add("X-Session-ID", sessionId)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestMain(m *testing.M) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.DebugLevel)

	c := config.TestConfig()
	service.SetConfig(c)

	code := m.Run()

	_ = c.CloseDb()

	os.Exit(code)
}
