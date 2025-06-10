# Gator
##### Go Blog Aggregator

### Prerequisites
1. Install Go. You can verify your installation by typing `go version` into your terminal.
2. Install PostgreSQL. You can verify your installation by typing `psql version` into your terminal.

### Installation
Run `go install github.com/curtisbraxdale/blog-gator` in your terminal.

### Config File
We need to create a config file in our home directory in order to make gator run properly.
Manually create a config file in your home directory, `~/.gatorconfig.json`, with the following content:
`{
  "db_url": "connection_string_goes_here",
  "current_user_name": "username_goes_here"
}`

`"current_user_name"` will be updated by the program but we need a `"db_url"`. You will get this by running your Postgres server. I recommend simply using the psql client.
Enter the `psql` shell:
Mac: `psql postgres`
Linux: `sudo -u postgres psql`
You should see a new prompt that looks like this :
`postgres=#`
Create a new database. I called mine `gator`:
`CREATE DATABASE gator;`
Connect to the new database:
`\c gator`
You should see a new prompt look like this:
`gator=#`
You can now type `exit` to leave the psql shell.
Your `"db_url"` will follow the format:
`"postgres://username:@localhost:5432/gator"`
Replace `username` with your username and paste this into you config file.

### Commands
All of the following commands will be used with the `gator` prefix. For example:
`gator register David`

##### Login
This logs the given user in.
Use: `gator login David`
##### Register
This registers a new user to the database.
Use: `gator register Larry`
##### Reset
*WARNING* This resets the database.
Use: `gator reset`
##### Users
This lists the users in the database.
Use: `gator users`
##### Agg
This aggregates posts from the feeds in the databse at a given interval. Meant to run in the background in a separate terminal window.
Use: `gator agg 1h`
##### AddFeed
This adds a feed to the database. Takes a name and a URL.
Use: `gator add TechCrunch https://techcrunch.com/feed/`
##### Feeds
This lists the feeds in the database.
Use: `gator feeds`
##### Follow
This follows the given feed for the current user. Takes a URL.
Use: `gator follow https://techcrunch.com/feed/`
##### Unfollow
This unfolllows the given feed for the current user. Takes a URL.
Use: `gator unfollow https://techcrunch.com/feed/`
##### Following
This lists the followed feeds by the current user.
Use: `gator following`
##### Browse
This shows the given number of recent posts from the followed feeds of the current user. Takes a number.
Use: `gator browse 5`
