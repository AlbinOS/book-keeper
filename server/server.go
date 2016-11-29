package server

import (
	"net/http"
    "html/template"

	"github.com/AlbinOS/book-keeper/fetcher"
	"github.com/AlbinOS/book-keeper/report"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
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
	sprint := c.Param("sprint")
	user := c.Param("user")
	timetrackings, err := report.SortedTimeTracking(sprint, user, JobInputs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err})
	} else {
		c.JSON(http.StatusOK, timetrackings)
	}
}

// ShowUI serves the main Book Keeper application page.
func ShowUI(c *gin.Context) {
	c.HTML(200, "index.html", gin.H{})
}

// Serve 'Em all
func Serve() {
	// Create a default gin stack
	router := gin.Default()
    html := template.Must(template.ParseFiles("./dist/index.html"))
    router.SetHTMLTemplate(html)
    // Book Keeper UI
    router.Static("/static", "./dist/static")
    router.StaticFile("/favicon.ico", "./dist/favicon.ico")



    router.NoRoute(ShowUI)

	// API routes
	api := router.Group("/api")
	api.GET("/ping", Ping)
	api.GET("/timetracking", TimeTracking)                       // All users, current sprints
	api.GET("/timetracking/users/:user", TimeTracking)           // One user, current sprints
	api.GET("/timetracking/sprints/:sprint", TimeTracking)       // All users, one sprint
	api.GET("/timetracking/sprints/:sprint/:user", TimeTracking) // One user, one sprint

	// Run the pool of JIRA ticket fetcher
	fetcher.StartWorkers(viper.GetInt("nbWorkers"), viper.GetString("endpoint"), viper.GetString("user"), viper.GetString("password"), JobInputs)

	// By default it serves on :8080 unless a
	// PORT environment variable was defined.
	router.Run()
}
