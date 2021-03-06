package fetcher

import (
	"fmt"
	"path"

	log "github.com/Sirupsen/logrus"
	jira "github.com/andygrunwald/go-jira"
)

// InitJiraClient init a JIRA client and check against errors
func InitJiraClient(endpoint string, user string, password string) (*jira.Client, error) {
	// Amoobi's JIRA instance
	jiraClient, _ := jira.NewClient(nil, endpoint)

	// Get an authentication cookie
	res, err := jiraClient.Authentication.AcquireSessionCookie(user, password)
	if !res {
		return nil, fmt.Errorf("Error authenticating with JIRA endpoint (%s), cause: '%s'", endpoint, "Wrong JIRA username or associated password !")
	} else if err != nil {
		return nil, fmt.Errorf("Error authenticating with JIRA endpoint (%s), cause: '%s'", endpoint, err)
	}

	return jiraClient, nil
}

// Tickets fetchs tickets according to defined jql
func Tickets(endpoint string, user string, password string, jql string) ([]jira.Issue, error) {

	// Init JIRA client
	jiraClient, err := InitJiraClient(endpoint, user, password)
	if err != nil {
		return nil, err
	}

	// Seach for issue in JIRA using jql query language
	tickets, _, err := jiraClient.Issue.Search(jql, &jira.SearchOptions{MaxResults: 100})
	if err != nil {
		return nil, fmt.Errorf("Unable to get issues using JQL='%s', cause: '%s'", jql, err)
	}

	log.Infof("There are %d issues selected by the query %s ...", len(tickets), jql)
	return tickets, nil
}

// SprintJql construct a valid sprint related JIRA Jql condition
func SprintJql(sprint string) string {

	// If no sprint is specified, get all issue for the current sprints
	if sprint == "" {
		return "sprint in openSprints() AND sprint not in futureSprints()"
	}
	return fmt.Sprintf("sprint=\"%s\"", sprint)
}

// ProjectJql construct a valid project related JIRA Jql condition
func ProjectJql(project string) string {
	return fmt.Sprintf("project=\"%s\"", project)
}

// UpdatedJql construct a valid updated date related JIRA Jql condition
func UpdatedJql(delay string) string {
	// If no sprint is specified, get all issue for the current sprints
	if delay == "" {
		return "updated>-30d"
	}
	return fmt.Sprintf("updated>%s", delay)
}

// TicketURL construct a valid JIRA cloud ticket browsing url
func TicketURL(endpoint string, ticketKey string) string {
	return endpoint + path.Join("browse", ticketKey)
}

// WorklogURL construct a valid JIRA cloud permanent worklog browsing url
func WorklogURL(endpoint string, ticketKey string, worklogID string) string {
	worklogFocusedEndpoint := fmt.Sprintf("%s%s%s%s%s", ticketKey, "?focusedWorklogId=", worklogID, "&page=com.atlassian.jira.plugin.system.issuetabpanels%3Aworklog-tabpanel#worklog-", worklogID)
	return endpoint + path.Join("browse", worklogFocusedEndpoint)
}
