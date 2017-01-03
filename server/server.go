package server

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/gin-contrib/sessions"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	gin "gopkg.in/gin-gonic/gin.v1"

	"github.com/AlbinOS/book-keeper-ui/dist"
	"github.com/AlbinOS/book-keeper/fetcher"
	"github.com/AlbinOS/book-keeper/report"
)

// JobInputs is a channel used to communicate with the pool of TicketFetcherWorker
var JobInputs = make(chan *fetcher.TicketFetcherJob, viper.GetInt("nbWorkers"))

// PingResponse is a JSON struct to use when responding to the client
type PingResponse struct {
	Pong string `json:"pong"`
}

// ErrorResponse is a JSON struct to use when responding to the client
type ErrorResponse struct {
	Error error `json:"error"`
}

// Ping is the handler for the GET /api/ping route.
// This will respond by a pong JSON message if the server is alive
func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, PingResponse{Pong: "OK"})
}

// TimeTracking is the handler for the GET /api/timetracking/* route.
// This will respond by rendering the timetracking html page.
func TimeTracking(c *gin.Context) {
	delay := c.Param("delay")
	timetrackings, err := report.SortedTimeTracking(delay, JobInputs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err})
	} else {
		c.JSON(http.StatusOK, timetrackings)
	}
}

// LoadHTMLIndex is sup
func LoadHTMLIndex() *template.Template {
	tmpl := template.New("_")

	src := dist.MustAsset(filepath.Join("build", "index.html"))
	tmpl = template.Must(
		tmpl.New("index.html").Parse(string(src)),
	)

	return tmpl
}

// ShowUI serves the main Book Keeper application page.
func ShowUI(c *gin.Context) {
	c.HTML(200, "index.html", gin.H{})
}

///////////
var googleConf = &oauth2.Config{
	ClientID:     "1026734103442-v7iksvb3dj8amhq6u3ufcubh6crdqgp9.apps.googleusercontent.com",
	ClientSecret: "ivNVSJSKKTW_tCv_XwEbBzMr",
	RedirectURL:  "http://127.0.0.1:3001/googleOauth2Callback",
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
	Endpoint:     google.Endpoint,
}

var githubConf = &oauth2.Config{
	ClientID:     "0fd982504c938e7714bc",
	ClientSecret: "3a61e3f52ae1acbc2d66450ed9deafcdea1d565b",
	RedirectURL:  "http://127.0.0.1:3001/githubOauth2Callback",
	Scopes:       []string{},
	Endpoint:     github.Endpoint,
}

func getGoogleLoginURL(state string) string {
	// State can be some kind of random generated hash string.
	// See relevant RFC: http://tools.ietf.org/html/rfc6749#section-10.12
	return googleConf.AuthCodeURL(state)
}

func getGithubLoginURL(state string) string {
	// State can be some kind of random generated hash string.
	// See relevant RFC: http://tools.ietf.org/html/rfc6749#section-10.12
	return githubConf.AuthCodeURL(state)
}

func randToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func loginHandler(c *gin.Context) {
	state := randToken()
	session := sessions.Default(c)
	session.Set("state", state)
	session.Save()
	c.Writer.Write([]byte("<html><title>Golang Google</title> <body> <a href='" + getGoogleLoginURL(state) + "'><button>Login with Google!</button></a><br/><a href='" + getGithubLoginURL(state) + "'><button>Login with Github!</button></a></body></html>"))
}

func googleAuthHandler(c *gin.Context) {
	// Handle the exchange code to initiate a transport.
	session := sessions.Default(c)
	retrievedState := session.Get("state")
	if retrievedState != c.Query("state") {
		c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("Invalid session state: %s", retrievedState))
		return
	}

	tok, err := googleConf.Exchange(oauth2.NoContext, c.Query("code"))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	fmt.Println("Token :", tok)
	client := googleConf.Client(oauth2.NoContext, tok)
	email, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	defer email.Body.Close()
	data, _ := ioutil.ReadAll(email.Body)
	fmt.Println("Email body: ", string(data))
	c.Status(http.StatusOK)
}

func githubAuthHandler(c *gin.Context) {
	// Handle the exchange code to initiate a transport.
	session := sessions.Default(c)
	retrievedState := session.Get("state")
	if retrievedState != c.Query("state") {
		c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("Invalid session state: %s", retrievedState))
		return
	}

	tok, err := githubConf.Exchange(oauth2.NoContext, c.Query("code"))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	fmt.Println("Token :", tok)
	client := githubConf.Client(oauth2.NoContext, tok)
	email, err := client.Get("https://api.github.com/user")
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	defer email.Body.Close()
	data, _ := ioutil.ReadAll(email.Body)
	fmt.Println("Email body: ", string(data))
	c.Status(http.StatusOK)
}

///////////

// Serve 'Em all
func Serve() {
	// Create a default gin stack
	router := gin.Default()
	store := sessions.NewCookieStore([]byte("secret"))
	router.Use(sessions.Sessions("mysession", store))

	// Serve Book Keeper UI
	router.SetHTMLTemplate(LoadHTMLIndex())
	fs := http.FileServer(dist.AssetFS())
	router.GET("/static/*filepath", func(c *gin.Context) { fs.ServeHTTP(c.Writer, c.Request) })
	router.GET("/favicon.ico", func(c *gin.Context) { fs.ServeHTTP(c.Writer, c.Request) })
	router.NoRoute(ShowUI)

	// Serve API
	api := router.Group("/api")
	api.GET("/ping", Ping)
	api.GET("/timetracking", TimeTracking)
	api.GET("/timetracking/:delay", TimeTracking)

	router.GET("/login", loginHandler)
	router.GET("/googleOauth2Callback", googleAuthHandler)
	router.GET("/githubOauth2Callback", githubAuthHandler)

	// Run the pool of JIRA ticket fetcher
	fetcher.StartWorkers(viper.GetInt("nbWorkers"), viper.GetString("endpoint"), viper.GetString("user"), viper.GetString("password"), JobInputs)

	// By default it serves on :8080 unless a
	// PORT environment variable was defined.
	router.Run()
}
