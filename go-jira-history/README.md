# JIRA Issue History Tracker (Go Version)

This is a Go implementation of the JIRA Issue History Tracker. It provides a web interface to fetch and display JIRA issue history, including creation dates, changes, and durations.

## Features

- Fetch issues from a JIRA project
- Display issue history in a sortable and searchable table
- Export data to CSV
- Modern, responsive UI with Bootstrap
- DataTables integration for enhanced table functionality

## Prerequisites

- Go 1.21 or later
- JIRA instance with API access
- Personal Access Token for JIRA authentication

## Environment Variables

Create a `.env` file in the project root with the following variables:

```
JIRA_URL=your_jira_url
JIRA_USERNAME=your_username
JIRA_PERSONAL_ACCESS_TOKEN=your_token
```

## Installation

1. Clone the repository
2. Navigate to the project directory:
   ```bash
   cd go-jira-history
   ```
3. Install dependencies:
   ```bash
   go mod download
   ```

## Running the Application

1. Start the server:
   ```bash
   go run .
   ```
2. Open your browser and navigate to:
   ```
   http://localhost:5000
   ```

## Usage

1. Enter a JIRA project key in the input field
2. Click "Fetch Issues" to retrieve the issue history
3. Use the table's search and sort functionality to find specific issues
4. Click "Download CSV" to export the data

## Project Structure

- `main.go`: Core JIRA API interaction functions
- `server.go`: Web server implementation using Gin
- `templates/index.html`: Web interface template
- `.env`: Environment variables (create this file)

## Dependencies

- `github.com/andygrunwald/go-jira`: JIRA API client
- `github.com/gin-gonic/gin`: Web framework
- `github.com/joho/godotenv`: Environment variable management

## License

MIT License 