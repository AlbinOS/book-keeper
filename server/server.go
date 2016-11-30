package server

import (
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/AlbinOS/book-keeper-ui/dist"
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
	timetrackings, err := report.SortedTimeTracking(sprint, JobInputs)
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

// Serve 'Em all
func Serve() {
	// Create a default gin stack
	router := gin.Default()

	// Serve Book Keeper UI
	router.SetHTMLTemplate(LoadHTMLIndex())
	fs := http.FileServer(dist.AssetFS())
	router.GET("/static/*filepath", func(c *gin.Context) { fs.ServeHTTP(c.Writer, c.Request) })
	router.GET("/favicon.ico", func(c *gin.Context) { fs.ServeHTTP(c.Writer, c.Request) })
	router.NoRoute(ShowUI)

	// Serve API
	api := router.Group("/api")
	api.GET("/ping", Ping)
	api.GET("/timetracking", TimeTracking)                 // All users, current sprints
	api.GET("/timetracking/sprints/:sprint", TimeTracking) // All users, one sprint

	// Run the pool of JIRA ticket fetcher
	fetcher.StartWorkers(viper.GetInt("nbWorkers"), viper.GetString("endpoint"), viper.GetString("user"), viper.GetString("password"), JobInputs)

	// By default it serves on :8080 unless a
	// PORT environment variable was defined.
	router.Run()
}
