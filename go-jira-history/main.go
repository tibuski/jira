package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/andygrunwald/go-jira"
	"github.com/joho/godotenv"
)

type IssueChange struct {
	Field      string
	From       string
	To         string
	ChangedAt  time.Time
	Duration   string
}

type IssueData struct {
	Summary     string
	Assignee    string
	IssueType   string
	ParentIssue string
	Created     time.Time
	Changes     []IssueChange
}

func getJiraClient() (*jira.Client, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	jiraURL := os.Getenv("JIRA_URL")
	token := os.Getenv("JIRA_PERSONAL_ACCESS_TOKEN")

	// Validate environment variables
	if jiraURL == "" {
		return nil, fmt.Errorf("JIRA_URL is not set in .env file")
	}
	if token == "" {
		return nil, fmt.Errorf("JIRA_PERSONAL_ACCESS_TOKEN is not set in .env file")
	}

	// Ensure the URL is properly formatted
	if !strings.HasSuffix(jiraURL, "/") {
		jiraURL = jiraURL + "/"
	}

	// Log the URL for debugging
	log.Printf("Connecting to JIRA at: %s", jiraURL)

	// Create a custom HTTP client with token authentication
	tp := jira.PATAuthTransport{
		Token: token,
	}

	// Create the JIRA client with the custom transport
	client, err := jira.NewClient(tp.Client(), jiraURL)
	if err != nil {
		return nil, fmt.Errorf("error creating JIRA client: %v", err)
	}

	// Test the connection with a simpler API call
	_, _, err = client.User.GetSelf()
	if err != nil {
		// Log more details about the error
		log.Printf("Authentication error details: %v", err)
		return nil, fmt.Errorf("error authenticating with JIRA: %v", err)
	}

	log.Printf("Successfully authenticated with JIRA")
	return client, nil
}

func getIssues(client *jira.Client, projectKey string) ([]jira.Issue, error) {
	jql := fmt.Sprintf("project=%s", projectKey)
	issues, _, err := client.Issue.Search(jql, &jira.SearchOptions{
		MaxResults: 100,
		Fields:     []string{"key", "summary", "assignee", "status", "issuetype", "subtasks"},
	})
	if err != nil {
		return nil, fmt.Errorf("error fetching issues: %v", err)
	}
	return issues, nil
}

func getIssueChangelog(client *jira.Client, issueKey string) ([]jira.ChangelogHistory, error) {
	issue, _, err := client.Issue.Get(issueKey, &jira.GetQueryOptions{
		Expand: "changelog",
	})
	if err != nil {
		return nil, fmt.Errorf("error fetching changelog for issue %s: %v", issueKey, err)
	}
	return issue.Changelog.Histories, nil
}

func calculateDuration(start, end time.Time) string {
	if end.IsZero() {
		return "N/A"
	}
	duration := end.Sub(start)
	return duration.String()
}

func saveToCSV(results map[string]IssueData, projectKey string) error {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("All_Issues_History_%s_%s.csv", projectKey, timestamp)

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"Issue Key", "Summary", "Assignee", "Issue Type", "Parent Issue", "Created Date", "Field", "From", "To", "Changed At", "Duration"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("error writing CSV header: %v", err)
	}

	// Write data
	for issueKey, issueData := range results {
		for _, change := range issueData.Changes {
			row := []string{
				issueKey,
				issueData.Summary,
				issueData.Assignee,
				issueData.IssueType,
				issueData.ParentIssue,
				issueData.Created.Format(time.RFC3339),
				change.Field,
				change.From,
				change.To,
				change.ChangedAt.Format(time.RFC3339),
				change.Duration,
			}
			if err := writer.Write(row); err != nil {
				return fmt.Errorf("error writing CSV row: %v", err)
			}
		}
	}

	return nil
}

func processIssues(projectKey string) (map[string]IssueData, error) {
	client, err := getJiraClient()
	if err != nil {
		return nil, err
	}

	issues, err := getIssues(client, projectKey)
	if err != nil {
		return nil, err
	}

	results := make(map[string]IssueData)

	for _, issue := range issues {
		changelog, err := getIssueChangelog(client, issue.Key)
		if err != nil {
			return nil, err
		}

		changes := make([]IssueChange, 0)
		created := time.Time{}

		// Get creation date from the first changelog entry
		if len(changelog) > 0 {
			created, err = time.Parse("2006-01-02T15:04:05.000-0700", changelog[0].Created)
			if err != nil {
				log.Printf("Error parsing creation date for issue %s: %v", issue.Key, err)
			}
			changes = append(changes, IssueChange{
				Field:     "created",
				From:      "None",
				To:        issue.Fields.Type.Name,
				ChangedAt: created,
				Duration:  "N/A",
			})
		}

		// Add changelog events
		for i, history := range changelog {
			// Parse the history creation date
			historyCreated, err := time.Parse("2006-01-02T15:04:05.000-0700", history.Created)
			if err != nil {
				log.Printf("Error parsing history date for issue %s: %v", issue.Key, err)
				continue
			}

			// Parse the next history creation date if available
			var nextHistoryCreated time.Time
			if i < len(changelog)-1 {
				nextHistoryCreated, err = time.Parse("2006-01-02T15:04:05.000-0700", changelog[i+1].Created)
				if err != nil {
					log.Printf("Error parsing next history date for issue %s: %v", issue.Key, err)
				}
			}

			for _, item := range history.Items {
				fromValue := "None"
				if item.FromString != "" {
					fromValue = item.FromString
				}

				toValue := "None"
				if item.ToString != "" {
					toValue = item.ToString
				}

				changes = append(changes, IssueChange{
					Field:     item.Field,
					From:      fromValue,
					To:        toValue,
					ChangedAt: historyCreated,
					Duration:  calculateDuration(historyCreated, nextHistoryCreated),
				})
			}
		}

		assignee := "Unassigned"
		if issue.Fields.Assignee != nil {
			assignee = issue.Fields.Assignee.DisplayName
		}

		parentIssue := "None"
		if issue.Fields.Parent != nil {
			parentIssue = issue.Fields.Parent.Key
		}

		results[issue.Key] = IssueData{
			Summary:     issue.Fields.Summary,
			Assignee:    assignee,
			IssueType:   issue.Fields.Type.Name,
			ParentIssue: parentIssue,
			Created:     created,
			Changes:     changes,
		}
	}

	return results, nil
} 