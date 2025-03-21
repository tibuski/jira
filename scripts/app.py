from flask import Flask, render_template, request, jsonify, send_file
from main import get_issues, get_issue_changelog, calculate_duration, save_to_csv
import os
from datetime import datetime

app = Flask(__name__)

@app.route('/')
def index():
    return render_template('index.html')

@app.route('/fetch-issues', methods=['POST'])
def fetch_issues():
    try:
        project_key = request.form.get('project_key')
        if not project_key:
            return jsonify({'error': 'Project key is required'}), 400
        
        # Get issues and process them
        issues = get_issues(project_key)
        results = {}
        
        for issue in issues:
            issue_key = issue.key
            summary = issue.fields.summary
            assignee = issue.fields.assignee.displayName if issue.fields.assignee else 'Unassigned'
            issue_type = issue.fields.issuetype.name
            parent_issue = issue.fields.parent.key if hasattr(issue.fields, 'parent') else 'None'
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
                'issue_type': issue_type,
                'parent_issue': parent_issue,
                'changes': changes
            }

        # Save to CSV
        save_to_csv(results)
        
        # Prepare data for the table
        table_data = []
        for issue_key, issue_data in results.items():
            for change in issue_data['changes']:
                table_data.append({
                    'issue_key': issue_key,
                    'summary': issue_data['summary'],
                    'assignee': issue_data['assignee'],
                    'issue_type': issue_data['issue_type'],
                    'parent_issue': issue_data['parent_issue'],
                    'field': change['field'],
                    'from': change['from'],
                    'to': change['to'],
                    'changed_at': change['changed_at'],
                    'duration': change['duration']
                })

        return jsonify({
            'success': True,
            'data': table_data,
            'message': f'Successfully fetched {len(issues)} issues'
        })

    except Exception as e:
        return jsonify({'error': str(e)}), 500

@app.route('/download-csv')
def download_csv():
    try:
        # Get the most recent CSV file in the current directory
        csv_files = [f for f in os.listdir('.') if f.startswith('All_Issues_History_') and f.endswith('.csv')]
        if not csv_files:
            return jsonify({'error': 'No CSV file found'}), 404
        
        # Get the most recent file
        latest_file = max(csv_files, key=os.path.getctime)
        
        # Use send_file with the correct path
        return send_file(
            latest_file,
            as_attachment=True,
            download_name=latest_file,
            mimetype='text/csv'
        )
    except Exception as e:
        return jsonify({'error': str(e)}), 500

if __name__ == '__main__':
    app.run(debug=True) 