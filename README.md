# Jira Data Center

## User rights
Keep in mind that Jira container is running as `(uid:guid)` `2001:2001`.   
Adapt `./data/shared` folder rights accordingly

## Compose options
`ports` and `JIRA_NODE_ID` are not mandatory in nodes definition but :
* `ports` allows to contact the node directly if needed
* `JIRA_NODE_ID` keeps clustering table clean 
