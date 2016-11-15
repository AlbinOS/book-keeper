package report

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/AlbinOS/book-keeper/fetcher"
	"github.com/serenize/snaker"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
)

// UserWorkLog describe work logged on one ticket by one user
type UserWorkLog struct {
	Ticket     string
	TicketURL  string
	User       string
	Date       string
	Timestamp  int64
	Duration   float64
	WorklogURL string
}

// UserWorkLogs is a slice of UserWorkLog
type UserWorkLogs []*UserWorkLog

func (slice UserWorkLogs) Len() int {
	return len(slice)
}

func (slice UserWorkLogs) Less(i, j int) bool {
	return slice[i].Timestamp < slice[j].Timestamp
}

func (slice UserWorkLogs) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

// SortedTimeTracking return timetracking by user sorted chronologically
func SortedTimeTracking(sprint string, worker string, jobInputs chan<- *fetcher.TicketFetcherJob) (UserWorkLogs, error) {
	// JIRA credentials
	user := viper.GetString("user")
	password := viper.GetString("password")
	endpoint := viper.GetString("endpoint")

	// Seach for issue in JIRA using jql query language
	jql := fmt.Sprintf("%s", fetcher.SprintJql(sprint))
	log.WithFields(log.Fields{
		"endpoint": endpoint,
		"worker":   worker,
	}).Infof("Fetching JIRA tickets using JQL: '%s'", jql)

	// Get all tickets for analysis
	rawIssues, err := fetcher.Tickets(endpoint, user, password, jql)
	if err != nil {
		return nil, err
	}

	// Fetch details for all selected issues
	fetchedTickets := fetcher.FetchTicketsDetail(rawIssues, jobInputs)

	// Hold every time tracking line
	var timetracking UserWorkLogs

	for i := 1; i <= len(rawIssues); i++ {
		ticket := <-fetchedTickets

		// By issue, export worked log by issue
		if ticket.Fields.Worklog != nil {
			for _, worklog := range ticket.Fields.Worklog.Worklogs {
				// If user is specified and it is not the right one, ignore the log
				if worker != "" && worker != worklog.Author.Name {
					continue
				}

				t := time.Time(worklog.Started)
				date := fmt.Sprintf("%d/%02d/%02d", t.Day(), t.Month(), t.Year())
				timetracking = append(timetracking, &UserWorkLog{Ticket: ticket.Fields.Summary, TicketURL: fetcher.TicketURL(endpoint, ticket.Key), User: snaker.SnakeToCamel(strings.Split(worklog.Author.Name, ".")[0]), Date: date, Timestamp: t.Unix(), Duration: (time.Duration(worklog.TimeSpentSeconds) * time.Second).Hours(), WorklogURL: fetcher.WorklogURL(endpoint, ticket.Key, worklog.ID)})
			}
		} else {
			log.Warningf("Issue %s assigned to %s doesn't have any work logged !", ticket.Key, ticket.Fields.Assignee)
		}
	}

	// Let's return everything we collected in chronological order
	log.Infof("There are %d timetracking line generated.", len(timetracking))
	sort.Sort(timetracking)

	return timetracking, nil
}

// TODO: CostDriver blabla
// func CostDriver(project string) error {
// 	// Prompt user for JIRA credentials
// 	user, password := jiraCredentialsPrompt()
//
// 	// Init JIRA client
// 	jiraClient, err := initJiraClient(jiraEndpoint, user, password)
// 	if err != nil {
// 		return err
// 	}
//
// 	// TODO
//
// 	return nil
// }
