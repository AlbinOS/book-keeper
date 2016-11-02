package fetcher

import (
	log "github.com/Sirupsen/logrus"
	"github.com/andygrunwald/go-jira"
)

// TicketFetcherJob represent a job to fetch a JIRA ticket
type TicketFetcherJob struct {
	Key       string
	Responses chan *jira.Issue
}

// TicketFetcherWorker fetch a complete JIRA ticket using its key and a decicated JIRA client.
// It will send the result to the issues channel
func TicketFetcherWorker(id int, jiraEndpoint string, user string, password string, inputs <-chan TicketFetcherJob) {
	log.WithFields(log.Fields{
		"workerID": id,
	}).Info("ticketFetcherWorker starting ...")

	// Init JIRA client connected to Amoobi's JIRA instance
	jiraClient, err := InitJiraClient(jiraEndpoint, user, password)
	if err != nil {
		log.Error(err)
		return
	}

	// For each issue key
	for ticketJob := range inputs {
		contextLogger := log.WithFields(log.Fields{
			"workerID": id,
			"issueKey": ticketJob.Key,
		})
		contextLogger.Info("ticketFetcherWorker processing ...")

		// Fetch all info about issueKey and send them through the channel
		issue, _, err := jiraClient.Issue.Get(ticketJob.Key)
		if err != nil {
			contextLogger.Errorf("Unable to get issue; cause: '%s'", err)
		} else {
			// Send fetched issue for further treatment
			ticketJob.Responses <- issue
		}
	}

	log.WithFields(log.Fields{
		"workerID": id,
	}).Info("ticketFetcherWorker work done !")
}

// ScheduleTicket schedules our issues for fetching
func ScheduleTicket(rawIssues []jira.Issue, inputs chan<- TicketFetcherJob, issues chan *jira.Issue) {
	// Schedule our issues for fetching
	for _, rawIssue := range rawIssues {
		// Send some work to our little workers that we love so much
		inputs <- TicketFetcherJob{Key: rawIssue.Key, Responses: issues}
	}
}

// StartWorkers will start numberOfWorkerToStart TicketFetcherWorker
func StartWorkers(numberOfWorkerToStart int, endPoint string, user string, password string, inputs <-chan TicketFetcherJob) {
	// This starts up CYCLOP_NUMBER_OF_WORKERS workers, we do it now to
	// take advantage of the time needed by the issue search to init our workers
	for w := 1; w <= numberOfWorkerToStart; w++ {
		go TicketFetcherWorker(w, endPoint, user, password, inputs)
	}
}
