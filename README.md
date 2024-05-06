# translate-bot

## Getting started
1. Set environment
First, download Google Cloud API json file, and set key.json at project root directory.

```
export SLACK_SIGNING_SECRET=<Your slack signing secret>
export SLACK_BOT_TOKEN=<Your slack bot oauth token>
export SLACK_APP_TOKEN=<Your slack app token for socket mode>
export GOOGLE_API_TOKEN=<Your Google Cloud Translate API token>
export GOOGLE_APPLICATION_CREDENTIALS=<Your Google Cloud default credentials>
export PROJECT_ID=<Your Google Cloud Project ID>
```

or create .env file.

```
SLACK_SIGNING_SECRET=~
SLACK_BOT_TOKEN=~
SLACK_APP_TOKEN=~
GOOGLE_API_TOKEN=~
GOOGLE_APPLICATION_CREDENTIALS=~
PROJECT_ID=~
```

2. Run bot server and check integration with slack
- exec file
```
go run main.go
```

- docker
```
docker compose up --build
```

