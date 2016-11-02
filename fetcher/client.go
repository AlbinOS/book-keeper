package fetcher

import (
	"fmt"

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
