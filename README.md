# gator

###

A simple RSS feed aggregator from a course project from boot.dev

## Tools

postgres2 sql database
psql db driver
golang 1.24, the go programming language
goose for database schema migrations
sqlc for sql query generation of go code.

## Creating the database

## Config

Requires a config file in your home directory "gatorconfig.json"
with the following structure:

```
{
    "Db_url": "your_database_connection_string",
    "Username": "your_username_here"
}
```

An example database connection string: "postgres://postgres:@localhost:5432/gator?sslmode=disable"
Username can be blank for the first run, it will be updated each time you log in or register a username.

## Compiling and Installing

After downloading the repo:
run "go build" to compile the program to an executable called 'gator'
You can now run the program from the repo directory by typing ./gator \<command\> \<command args\>

To install for use anywhere on your machine by your user:
run "go install" to install the program to your machine.

run "gator help" to see a list of commands and their descriptions.

```
./gator help
command: help
help
login <username> - logs in the user.
register <username> - adds a user to the database and automatically logs in the user.
reset - drops rows data but keep tables.
users - lists all users.
feeds - lists all feeds.
addfeed <name> <url>
agg - print rss feeds to console.
follow <url> - adds the feed for the url to the users follows.
following - lists all feeds followed by the current user.
unfollow <url> - removes follow for url for current user.
browse <limit> - prints up to limit posts for user feeds.
```

## Running commands

### Adding a user

You can add a user with the register command.
./gator register "username"

### Adding a feed

A feed can be added by a user to the database with the addfeed command.
./gator addfeed "Feed Name" "url of feed rss api"

### Following a feed

A feed is automatically followed by the user that added the feed.
Other users may request to follow a feed by providing the url of the feed.
./gator follow "url of feed"

### Browsing posts

Display posts for users feeds by running the browse command.
The browse command takes an optional limit parameter that limits the number of posts displayed.
The default limit is 2 if the parameter is not provided.

### Aggregating posts

To begin scraping rss feeds, run the agg command.
It requires setting an interval that is a go time duration.
Be careful that querying too fast may cause you to be blocked by the feed provider.
The example below runs once every 5 minutes.
./gator agg 5m

### Resetting the database

You can reset the database for testing by running the reset command.
Be warned this will clear all data for all feeds, users, and posts.
