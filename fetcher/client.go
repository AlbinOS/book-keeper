package fetcher

import (
	"fmt"

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

// Tickets fetchs tickets according to defined criterias
func Tickets(endpoint string, user string, password string, jql string) ([]jira.Issue, error) {

	// Init JIRA client
	jiraClient, err := InitJiraClient(endpoint, user, password)
	if err != nil {
		return nil, err
	}

	// Seach for issue in JIRA using jql query language
	tickets, _, err := jiraClient.Issue.Search(jql, nil)
	if err != nil {
		return nil, fmt.Errorf("Unable to get issues using JQL='%s', cause: '%s'", jql, err)
	}

	log.Infof("There are %d issues selected by the query %s ...", len(tickets), jql)
	return tickets, nil
}

// SprintJql construct a valid sprint related JIRA Jql condition
func SprintJql(sprint string) string {

	// If no sprint is specified, get all issue for the current sprint
	if sprint == "" {
		return "sprint in openSprints(OPS) AND sprint not in futureSprints(OPS)"
	}
	return fmt.Sprintf("sprint = %s", sprint)
}

// ProjectJql construct a valid project related JIRA Jql condition
func ProjectJql(project string) string {
	return fmt.Sprintf("project=%s", project)
}
