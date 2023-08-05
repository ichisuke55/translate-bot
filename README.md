# translate-bot

## Getting started
1. Set environment
First, download Google Cloud API json file, and set key.json

```
export SLACK_SIGNING_SECRET=<Your slack signing secret>
export SLACK_BOT_TOKEN=<Your slack bot oauth token>
export SLACK_APP_TOKEN=<Your slack app token for socket mode>
export GOOGLE_API_TOKEN=<Your Google Cloud Translate API token>
export GOOGLE_APPLICATION_CREDENTIALS=<Your Google Cloud default credentials>
```

2. Run bot server and integrate with slack
```
go run main.go
```

