import os
import csv
from jira import JIRA
from dotenv import load_dotenv
from datetime import datetime
import sys

# Load environment variables
load_dotenv()

def validate_environment():
    required_vars = ['JIRA_URL', 'JIRA_USERNAME', 'JIRA_PERSONAL_ACCESS_TOKEN', 'PROJECT_KEY']
    missing_vars = [var for var in required_vars if not os.getenv(var)]
    if missing_vars:
        print(f"Error: Missing required environment variables: {', '.join(missing_vars)}")
        sys.exit(1)

# Validate environment variables before proceeding
validate_environment()

JIRA_URL = os.getenv('JIRA_URL')
JIRA_USERNAME = os.getenv('JIRA_USERNAME')
JIRA_PERSONAL_ACCESS_TOKEN = os.getenv('JIRA_PERSONAL_ACCESS_TOKEN')
PROJECT_KEY = os.getenv('PROJECT_KEY')

try:
    # Connect to Jira using personal access token
    jira = JIRA(JIRA_URL, token_auth=JIRA_PERSONAL_ACCESS_TOKEN)
except Exception as e:
    print(f"Error connecting to JIRA: {str(e)}")
    sys.exit(1)

def get_issues(project_key):
    try:
        jql = f'project={project_key}'
        start_at = 0
        max_results = 100
        all_issues = []
        
        while True:
            issues = jira.search_issues(jql, startAt=start_at, maxResults=max_results, 
                                     fields='key,summary,assignee,status,issuetype,subtasks')
            all_issues.extend(issues)
            
            if len(issues) < max_results:
                break
                
            start_at += max_results
            
        return all_issues
    except Exception as e:
        print(f"Error fetching issues: {str(e)}")
        return []

def get_issue_changelog(issue_key):
    try:
        issue = jira.issue(issue_key, expand='changelog')
        return issue.changelog.histories
    except Exception as e:
        print(f"Error fetching changelog for issue {issue_key}: {str(e)}")
        return []

def calculate_duration(start, end):
    start_dt = datetime.strptime(start, '%Y-%m-%dT%H:%M:%S.%f%z')
    end_dt = datetime.strptime(end, '%Y-%m-%dT%H:%M:%S.%f%z')
    duration = end_dt - start_dt
    return duration

def save_to_csv(results, project_key):
    try:
        timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
        filename = f"All_Issues_History_{project_key}_{timestamp}.csv"
        # Get the absolute path to the project root directory
        project_root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
        filepath = os.path.join(project_root, filename)
        
        with open(filepath, mode='w', newline='', encoding='utf-8') as file:
            writer = csv.writer(file)
            writer.writerow(['Issue Key', 'Summary', 'Assignee', 'Issue Type', 'Parent Issue', 'Created Date', 'Field', 'From', 'To', 'Changed At', 'Duration'])
            for issue_key, issue_data in results.items():
                for change in issue_data['changes']:
                    writer.writerow([
                        issue_key,
                        issue_data['summary'],
                        issue_data['assignee'],
                        issue_data['issue_type'],
                        issue_data['parent_issue'],
                        issue_data['created'],
                        change['field'],
                        change['from'],
                        change['to'],
                        change['changed_at'],
                        change['duration']
                    ])
        print(f"CSV file '{filename}' has been created successfully.")
        return filepath
    except Exception as e:
        print(f"Error saving CSV file: {str(e)}")
        sys.exit(1)

def main(project_key):
    results = {}
    issues = get_issues(project_key)
    for issue in issues:
        issue_key = issue.key
        summary = issue.fields.summary
        assignee = issue.fields.assignee.displayName if issue.fields.assignee else 'Unassigned'
        issue_type = issue.fields.issuetype.name
        parent_issue = issue.fields.parent.key if hasattr(issue.fields, 'parent') else 'None'
        created = issue.fields.created
        changelog = get_issue_changelog(issue_key)
        
        changes = []
        
        # Add creation event
        changes.append({
            'field': 'created',
            'from': 'None',
            'to': issue_type,
            'changed_at': created,
            'duration': 'N/A'  # Duration for creation event is N/A
        })
        
        # Add changelog events
        for history in changelog:
            for item in history.items:
                from_value = item.fromString if item.fromString else 'None'
                to_value = item.toString if item.toString else 'None'
                changed_at = history.created
                next_changed_at = history.created
                
                # Find the next change
                for next_history in changelog:
                    if next_history.created > changed_at:
                        next_changed_at = next_history.created
                        break
                
                duration = calculate_duration(changed_at, next_changed_at)
                changes.append({
                    'field': item.field,
                    'from': from_value,
                    'to': to_value,
                    'changed_at': changed_at,
                    'duration': str(duration)
                })
        
        results[issue_key] = {
            'summary': summary,
            'assignee': assignee,
            'issue_type': issue_type,
            'parent_issue': parent_issue,
            'created': created,
            'changes': changes
        }
    
    return results

if __name__ == "__main__":
    project_key = os.getenv('PROJECT_KEY')
    if not project_key:
        print("Error: PROJECT_KEY environment variable is required")
        sys.exit(1)
    results = main(project_key)
    save_to_csv(results, project_key)
