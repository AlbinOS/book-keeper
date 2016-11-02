package server

import (
	"net/http"

	"github.com/AlbinOS/book-keeper/fetcher"
	"github.com/AlbinOS/book-keeper/report"

	"github.com/gin-gonic/gin"
	eztemplate "github.com/michelloworld/ez-gin-template"
	"github.com/spf13/viper"
)

// JobInputs is a channel used to communicate with the pool of TicketFetcherWorker
var JobInputs = make(chan fetcher.TicketFetcherJob, viper.GetInt("nbWorkers"))

// PingResponse is a JSON struct to use when responding to the client
type PingResponse struct {
	Pong string `json:"pong"`
}

// Ping is the handler for the GET /api/ping route.
// This will respond by a pong JSON message if the server is alive
func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, PingResponse{Pong: "OK"})
}

// Home is the handler for the GET / route.
// This will respond by rendering the home html page.
func Home(c *gin.Context) {
	c.HTML(http.StatusOK, "home/root", gin.H{})
}

// About is the handler for the GET /about route.
// This will respond by rendering the about html page.
func About(c *gin.Context) {
	c.HTML(http.StatusOK, "home/about", gin.H{})
}

// TimeTracking is the handler for the GET /timetracking route.
// This will respond by rendering the timetracking html page.
func TimeTracking(c *gin.Context) {
	timetrackings, _ := report.TimeTracking("", JobInputs)
	c.HTML(http.StatusOK, "report/timetracking", gin.H{"timetrackings": timetrackings})
}

// Serve 'Em all
func Serve() {
	// Create a default gin stack
	router := gin.Default()

	// Load templates
	render := eztemplate.New()
	render.TemplatesDir = "views/"
	render.Debug = gin.IsDebugging()
	render.Ext = ".tmpl"
	router.HTMLRender = render.Init()

	// Load all static assets
	router.Static("/static", "./static")

	// API routes
	api := router.Group("/api")
	api.GET("/ping", Ping)

	// Root
	router.GET("/", Home)

	// About
	router.GET("/about", About)

	// TimeTracking
	router.GET("/report/timetracking", TimeTracking)

	// Run the pool of JIRA ticket fetcher
	fetcher.StartWorkers(viper.GetInt("nbWorkers"), viper.GetString("endpoint"), viper.GetString("user"), viper.GetString("password"), JobInputs)

	// By default it serves on :8080 unless a
	// PORT environment variable was defined.
	router.Run()
}
