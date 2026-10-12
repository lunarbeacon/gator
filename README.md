# Gator
## Requirements
You need to install the SQL database postgresql and the programming 
language Go to run this software.

## Installation
Use the command 
```Go
go install gator
```
## Set up
Create a configuration file in the gator directory called 
1. Create a JSON configuration file called .gatorconfig.json in your home directory, 
in the following format. You'll need to change the database connection string and the username.
```json
{
  "db_url": "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable",
  "current_user_name": "alan"
}
```
2. Install the go application goose
```Go
go install goose
```
3. Run the command
```shell
goose postgres <db_connection_string> up
```
## Commands
The following commands are available:
- gator register \<name> (register a user name)
- gator login \<name> (login as \<name>)
- gator addfeed "Feed name" "feed url" (add a feed and follow it)
- gator follow "feed url" (follow a feed someone else has added)
- gator feeds (list feeds)
- gator unfollow "feed url" (unfollow a feed)
- gator browse \<limit> (browse up to \<limit> posts from the feeds you follow)
- gator agg (retrieve the latest posts from the feeds)