# goto - a simple, fast, golang based link shortener

As a part of [RaspAPI](https://raspapi.hackclub.com/), I decided to write a link shortener in Golang so I could get more comfortable with the language. It's mainly API based, but I plan to add a frontend to it soon.

## Features

* **Fast**: Close to 12,000 reqs/sec (for a single server) for most endpoints - tested on my own hardware so it may not be accurate.
* **Link Statistics (Hit Tracking)**: Tracks the number of clicks/hits for each link. No unnessecary data is collected.
* **User Accounts/Authentication**: Really simple authentication: an access token generated upon creation of your account, which is hashed in the database.
* **Custom Short Links**: You can provide your own short link name or use a randomly generated one.
* **Private Links**: You can create private links, which means that you need authentication to see link statistics.
* **API Docs**: Poorly written API docs available @ `/api/docs` on any instance of the server.

## Installation

```sh
$ git clone https://github.com/radeeyate/goto.git
$ cd goto
$ go get

# running the server
$ go build .

# linux or macos
$ ./goto

# windows
$ goto.exe
```

Keep in mind that to run the server, you must set up an environment variable (`.env`) file. Please use this format:

```
CONNECTION_STRING="mongodb link of database"
BASE_URL="base URL of the server"
REGISTRATIONS_ENABLED="true or false"
```

## API Documentation

Please see [docs.md](docs.md) for more information.

## License

MIT License (see LICENSE.md)
