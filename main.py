import os
import csv
from jira import JIRA
from dotenv import load_dotenv
from datetime import datetime

# Load environment variables
load_dotenv()

JIRA_URL = os.getenv('JIRA_URL')
JIRA_USERNAME = os.getenv('JIRA_USERNAME')
JIRA_PERSONAL_ACCESS_TOKEN = os.getenv('JIRA_PERSONAL_ACCESS_TOKEN')
PROJECT_KEY = os.getenv('PROJECT_KEY')

# Connect to Jira using personal access token
jira = JIRA(JIRA_URL, token_auth=JIRA_PERSONAL_ACCESS_TOKEN)

def get_issues():
    jql = f'project={PROJECT_KEY}'
    issues = jira.search_issues(jql, fields='key,summary,assignee,status')
    print(f"Fetched {len(issues)} issues")
    return issues

def get_issue_changelog(issue_key):
    issue = jira.issue(issue_key, expand='changelog')
    print(f"Fetched changelog for issue {issue_key}")
    return issue.changelog.histories

def calculate_duration(start, end):
    start_dt = datetime.strptime(start, '%Y-%m-%dT%H:%M:%S.%f%z')
    end_dt = datetime.strptime(end, '%Y-%m-%dT%H:%M:%S.%f%z')
    duration = end_dt - start_dt
    return duration

def save_to_csv(results):
    filename = "All_Issues_History.csv"
    with open(filename, mode='w', newline='') as file:
        writer = csv.writer(file)
        writer.writerow(['Issue Key', 'Summary', 'Assignee', 'Field', 'From', 'To', 'Changed At', 'Duration'])
        for issue_key, issue_data in results.items():
            for change in issue_data['changes']:
                writer.writerow([
                    issue_key,
                    issue_data['summary'],
                    issue_data['assignee'],
                    change['field'],
                    change['from'],
                    change['to'],
                    change['changed_at'],
                    change['duration']
                ])
    print("CSV file has been created successfully.")

def main():
    results = {}
    issues = get_issues()
    for issue in issues:
        issue_key = issue.key
        summary = issue.fields.summary
        assignee = issue.fields.assignee.displayName if issue.fields.assignee else 'Unassigned'
        changelog = get_issue_changelog(issue_key)
        
        changes = []
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
            'changes': changes
        }
        print(f"Processed issue {issue_key}")
    
    return results

if __name__ == "__main__":
    results = main()
    save_to_csv(results)
