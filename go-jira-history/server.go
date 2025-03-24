package main

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Serve static files
	r.Static("/static", "./static")

	// Load HTML template
	r.LoadHTMLGlob("templates/*")

	// Routes
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.POST("/fetch-issues", func(c *gin.Context) {
		projectKey := c.PostForm("project_key")
		if projectKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Project key is required"})
			return
		}

		results, err := processIssues(projectKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Save to CSV
		if err := saveToCSV(results, projectKey); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Prepare data for the table
		var tableData []map[string]interface{}
		for issueKey, issueData := range results {
			for _, change := range issueData.Changes {
				tableData = append(tableData, map[string]interface{}{
					"issue_key":    issueKey,
					"summary":      issueData.Summary,
					"assignee":     issueData.Assignee,
					"issue_type":   issueData.IssueType,
					"parent_issue": issueData.ParentIssue,
					"created":      issueData.Created.Format(time.RFC3339),
					"field":        change.Field,
					"from":         change.From,
					"to":           change.To,
					"changed_at":   change.ChangedAt.Format(time.RFC3339),
					"duration":     change.Duration,
				})
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    tableData,
			"message": "Successfully fetched issues",
		})
	})

	r.GET("/download-csv", func(c *gin.Context) {
		// Find the most recent CSV file
		files, err := filepath.Glob("All_Issues_History_*.csv")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error finding CSV files"})
			return
		}

		if len(files) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "No CSV file found"})
			return
		}

		// Sort files by modification time
		sort.Slice(files, func(i, j int) bool {
			infoI, _ := os.Stat(files[i])
			infoJ, _ := os.Stat(files[j])
			return infoI.ModTime().After(infoJ.ModTime())
		})

		latestFile := files[0]
		c.File(latestFile)
	})

	r.Run(":5000")
} 