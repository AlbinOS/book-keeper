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

// 64int array type to be used with sort
type int64arr []int64

// 64int array type functions to be used with sort
func (a int64arr) Len() int           { return len(a) }
func (a int64arr) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a int64arr) Less(i, j int) bool { return a[i] < a[j] }

// UserWorkLog describe work logged on one ticket by one user
type UserWorkLog struct {
	Ticket   string
	User     string
	Date     string
	Duration float64
}

// SortedTimeTracking return timetracking by user sorted chronologically
func SortedTimeTracking(project string, sprint string, worker string, jobInputs chan<- *fetcher.TicketFetcherJob) ([]*UserWorkLog, error) {
	// JIRA credentials
	user := viper.GetString("username")
	password := viper.GetString("password")
	endpoint := viper.GetString("endpoint")

	// Seach for issue in JIRA using jql query language
	jql := fmt.Sprintf("%s AND %s", fetcher.ProjectJql(project), fetcher.SprintJql(sprint))

	// Get all tickets for analysis
	rawIssues, err := fetcher.Tickets(endpoint, user, password, jql)
	if err != nil {
		return nil, err
	}

	// Fetch details for all selected issues
	fetchedTickets := fetcher.FetchTicketsDetail(rawIssues, jobInputs)

	// Hold every time tracking line
	timetracking := make(map[int64]*UserWorkLog)
	var keys int64arr

	for i := 1; i <= len(rawIssues); i++ {
		ticket := <-fetchedTickets

		// By issue, export worked log by issue
		if ticket.Fields.Worklog != nil {
			for _, workLog := range ticket.Fields.Worklog.Worklogs {
				// If user is specified and it is not the right one, ignore the log
				if worker != "" && worker != workLog.Author.Name {
					continue
				}

				t := time.Time(workLog.Started)
				keys = append(keys, t.Unix())

				date := fmt.Sprintf("%d/%02d/%02d", t.Day(), t.Month(), t.Year())
				timetracking[t.Unix()] = &UserWorkLog{Ticket: ticket.Fields.Summary, User: snaker.SnakeToCamel(strings.Split(workLog.Author.Name, ".")[0]), Date: date, Duration: (time.Duration(workLog.TimeSpentSeconds) * time.Second).Hours()}
				//	timetracking[t.Unix()] = &UserWorkLog{User: workLog.Author.Name, Ticket: ticket, WorkLog: &workLog}
			}
		} else {
			log.Warningf("Issue %s assigned to %s doesn't have any work logged !", ticket.Key, ticket.Fields.Assignee)
		}
	}

	// Let's return everything we collected in chronological order
	var sortedTimetracking []*UserWorkLog
	sort.Sort(keys)
	log.Infof("There are %d timetracking line generated :", len(timetracking))
	for _, key := range keys {
		sortedTimetracking = append(sortedTimetracking, timetracking[key])
	}

	return sortedTimetracking, nil
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
