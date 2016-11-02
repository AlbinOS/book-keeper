package report

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/AlbinOS/book-keeper/fetcher"

	log "github.com/Sirupsen/logrus"
	"github.com/andygrunwald/go-jira"
	"github.com/serenize/snaker"
	"github.com/spf13/viper"
)

// 64int array type to be used with sort
type int64arr []int64

// 64int array type functions to be used with sort
func (a int64arr) Len() int           { return len(a) }
func (a int64arr) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a int64arr) Less(i, j int) bool { return a[i] < a[j] }

// TimeTracking output for Operations & Analysis To Do List
func TimeTracking(sprint string, jobInputs chan<- fetcher.TicketFetcherJob) ([]string, error) {

	// JIRA credentials
	user := viper.GetString("username")
	password := viper.GetString("password")
	endpoint := viper.GetString("endpoint")

	// Init JIRA client
	jiraClient, err := fetcher.InitJiraClient(endpoint, user, password)
	if err != nil {
		return nil, err
	}

	// In order to use our pool of workers we need a way to get the result of the work
	issues := make(chan *jira.Issue, viper.GetInt("nbWorkers"))

	// Global JQL query
	jqlString := "project=OPS"

	// If no sprint is specified, get all issue for the current sprint
	if sprint == "" {
		jqlString = jqlString + " AND sprint in openSprints(OPS) AND sprint not in futureSprints(OPS)"
	} else {
		jqlString = jqlString + fmt.Sprintf(" AND sprint = %s", sprint)
	}

	// Seach for issue in JIRA using jql query language
	rawIssues, _, err := jiraClient.Issue.Search(jqlString, nil)
	if err != nil {
		return nil, fmt.Errorf("Unable to get issues using JQL='%s', cause: '%s'", jqlString, err)
	}
	log.Infof("There are %d issues selected by the query %s ...", len(rawIssues), jqlString)

	// Do it in a go function to avoid blocking due to queue size
	go fetcher.ScheduleTicket(rawIssues, jobInputs, issues)

	// Hold every time tracking line
	timetracking := make(map[int64]string)
	var keys int64arr

	// Read from workers
	for i := 1; i <= len(rawIssues); i++ {
		// By issue, export worked log by issue
		issue := <-issues
		if issue.Fields.Worklog != nil {
			for _, workLog := range issue.Fields.Worklog.Worklogs {
				name := strings.Split(workLog.Author.Name, ".")[0]
				t := time.Time(workLog.Started)
				date := fmt.Sprintf("%d/%02d/%02d", t.Day(), t.Month(), t.Year())
				timetracking[t.Unix()] = fmt.Sprintf("%s\t%s\t%s\t%.2f\n", issue.Fields.Summary, snaker.SnakeToCamel(name), date, (time.Duration(workLog.TimeSpentSeconds) * time.Second).Hours())
				keys = append(keys, t.Unix())
			}
		} else {
			log.Warningf("Issue %s assigned to %s doesn't have any work logged !", issue.Key, issue.Fields.Assignee)
		}
	}

	// Let's return everything we collected in chronological order
	var sortedTimetracking []string
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
