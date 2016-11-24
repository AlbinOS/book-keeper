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
var JobInputs = make(chan *fetcher.TicketFetcherJob, viper.GetInt("nbWorkers"))

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
	sprint := c.Param("sprint")
	user := c.Param("user")
	timetrackings, err := report.SortedTimeTracking(sprint, user, JobInputs)

	switch c.NegotiateFormat(gin.MIMEHTML, gin.MIMEJSON) {
	case gin.MIMEHTML:
		if err == nil {
			c.HTML(http.StatusOK, "report/timetracking", gin.H{"timetrackings": timetrackings})
		} else {
			c.HTML(http.StatusInternalServerError, "errors/500", gin.H{"message": err})
		}
	case gin.MIMEJSON:
		c.JSON(200, timetrackings)
	}
}

// Serve 'Em all
func Serve() {
	// Create a default gin stack
	router := gin.Default()
	router.Use(func(context *gin.Context) {
		// Add header Access-Control-Allow-Origin
		context.Writer.Header().Add("Access-Control-Allow-Origin", "*")
		context.Next()
	})

	// Load templates
	render := eztemplate.New()
	render.TemplatesDir = "views/"
	render.Debug = gin.IsDebugging()
	render.Ext = ".tmpl"
	router.HTMLRender = render.Init()

	// Load all static assets
	router.Static("/static", "./static")
	router.StaticFile("/favicon.ico", "./static/favicon.ico")

	// API routes
	api := router.Group("/api")
	api.GET("/ping", Ping)

	// Root
	router.GET("/", Home)

	// About
	router.GET("/about", About)

	// TimeTracking
	report := router.Group("/report")
	report.GET("/timetracking", TimeTracking) // All users, current sprint

	report.GET("/timetracking/users/:user", TimeTracking) // One user, current sprint

	report.GET("/timetracking/sprints/:sprint", TimeTracking) // All users, one sprint

	report.GET("/timetracking/sprints/:sprint/:user", TimeTracking) // One user, one sprint

	// Run the pool of JIRA ticket fetcher
	fetcher.StartWorkers(viper.GetInt("nbWorkers"), viper.GetString("endpoint"), viper.GetString("user"), viper.GetString("password"), JobInputs)

	// By default it serves on :8080 unless a
	// PORT environment variable was defined.
	router.Run()
}
