# AdventLeaderBot

After the success of the [Advent of Code](https://adventofcode.com/) bot [of last
year](https://github.com/michielappelman/adventleader) I decided to this year up the ante and
actually get it to listen for commands and be able to host multiple [Webex
Teams](https://www.webex.com/products/teams/index.html) rooms and associated Advent of Code
leaderboards.

The address of the bot currently running in production is: AdventLeaderBot@webex.bot You can add the
Bot to a room or to a space, but remember to always @-mention it to give it any commands.

The bot uses the custom
[github.com/michielappelman/leaderboard](https://github.com/michielappelman/leaderboard) Go package
to retrieve the leaderboard JSON files and updates the rooms that subscribed to those leaderboards
every 5 minutes, if there are changes.

## Bot Commands

There are four commands defined currently to control the behaviour of the bot once it's added to a
Webex Teams room:

- ` register 12345 xyz` which will register the room with a certain leaderboard and the required
session Cookie for the adventofcode.com website.
- `poll`, polls the current standing of the leaderboard and updates the room.
- `year 2018` will set the year to poll for this room.
- `help` will give an overview of the commands and the current settings for the room.

## Google App Engine

The bot is built to be run on GAE and makes use of the GCP Datastore to store Room subscriptions.
Files related to the GAE deployment:

- The `app.yaml` file requires the Environment Variable `WEBEX_TEAMS_TOKEN` set for the Webex Teams
Bot auth token..
- The `cron.yaml` specifies the 5 minute interval for polling the leaderboards for all the rooms.

## TODO

- [ ] Set a timer for invalid API failures, to stop the bot from spamming when the cookie expires.
- [ ] Match Webex Teams user ID's with player names, to know who's who.

