# Jira Data Center

## Docker

### User rights
Keep in mind that Jira container is running as `(uid:guid)` `2001:2001`.   
Adapt `./data/shared` folder rights accordingly.

### Compose options
`ports` and `JIRA_NODE_ID` are not mandatory in nodes definition but:
* `ports` allows to contact the node directly if needed.
* `JIRA_NODE_ID` keeps clustering table clean.

### Starting the containers
1. Rename docker-compose.yml.example to docker-compose.yml
2. Edit the `x-anchors` in docker compose-compose.yml to adapt to your needs.
3. Run the following command to start the containers:
   ```bash
   docker compose up -d && docker compose logs -f
   ```


## Python

* Copy .env.example to .env and adapt variables :
  ```bash
  JIRA_URL=https://jira.mydomain.com
  JIRA_USERNAME=MyUser
  JIRA_PERSONAL_ACCESS_TOKEN=my_very_long_token_GeneratedInJiraUserProfile
  PROJECT_KEY=MyProject
  ```

* Install python requirements :
  ```bash
  python -m pip install -r requirements.txt
  ```

* run `main.py` (CLI) or `app.py` (Flask Web Interface)
