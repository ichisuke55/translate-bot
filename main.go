package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"unicode"

	"github.com/ichisuke55/translate-bot/config"
	"github.com/ichisuke55/translate-bot/logging"
	"go.uber.org/zap"

	translate "cloud.google.com/go/translate/apiv3"
	"cloud.google.com/go/translate/apiv3/translatepb"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

func translateText(projectID string, sourceLang string, targetLang string, text string) (string, error) {
	ctx := context.Background()
	c, err := translate.NewTranslationClient(ctx)
	if err != nil {
		return "", err
	}
	defer c.Close()

	req := &translatepb.TranslateTextRequest{
		Contents:           []string{text},
		MimeType:           "text/plain", // Mime types: "text/plain", default is "text/html"
		Parent:             fmt.Sprintf("projects/%s/locations/global", projectID),
		SourceLanguageCode: sourceLang,
		TargetLanguageCode: targetLang,
	}

	resp, err := c.TranslateText(ctx, req)
	if err != nil {
		return "[ERROR] cannot translate message", err
	}

	// Return the translation result for each input text provided
	var msg string
	for _, translation := range resp.GetTranslations() {
		msg = fmt.Sprintf("%v\n", translation.GetTranslatedText())
	}
	return msg, nil
}

func trancateText(msg string) (string, error) {
	// slack's URL text style is <URL>
	urlRegexp := `<(http:\/\/|https:\/\/)?[a-z0-9]+([\-\.]{1}[a-z0-9]+)*\.[a-z]{2,5}(:[0-9]{1,5})?(\/.*)?>`
	rep := regexp.MustCompile(urlRegexp)

	// if URL contains in message, trancate it.
	if match := rep.MatchString(msg); match {
		msg = rep.ReplaceAllString(msg, "")
	}
	return msg, nil
}

func isJapanese(msg string) bool {
	m := []rune(msg)
	if unicode.In(m[0], unicode.Hiragana) {
		return true
	}
	if unicode.In(m[0], unicode.Katakana) {
		return true
	}
	if unicode.In(m[0], unicode.Han) { // check 漢字
		return true
	}
	return false
}

func main() {
	// Load environment variables
	conf, err := config.NewEnvConfig()
	if err != nil {
		log.Fatal("cannot load environment variables")
	}

	// Create and open logfile
	logfile, err := os.OpenFile(conf.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("cannot create or open log file: %s", err)
	}
	defer logfile.Close()

	// Init logger
	logger, err := logging.NewLogger(logfile)
	if err != nil {
		log.Fatal("cannot initialize logger")
	}
	defer logger.Sync() //nolint:errcheck

	// Set Slack API client
	client := slack.New(
		conf.SlackBotToken,
		slack.OptionAppLevelToken(conf.SlackAppToken),
		// slack.OptionDebug(true),
		slack.OptionLog(log.New(os.Stdout, "apiClient: ", log.Lshortfile|log.LstdFlags)),
	)
	// Test authenticate connection
	_, err = client.AuthTest()
	if err != nil {
		logger.Fatal("failed to authenticate", zap.Error(err))
	}

	socketClient := socketmode.New(
		client,
		// socketmode.OptionDebug(true),
		socketmode.OptionLog(log.New(os.Stdout, "socketClient: ", log.Lshortfile|log.LstdFlags)),
	)

	logger.Info("Bot starting...")

	go func() {
		for event := range socketClient.Events {
			switch event.Type {
			case socketmode.EventTypeEventsAPI:
				eventAPIEvent, ok := event.Data.(slackevents.EventsAPIEvent)
				if !ok {
					logger.Info("Ignored event", zap.Any("event", event))
				}
				// Prevent from sending duplicate message
				socketClient.Ack(*event.Request)

				switch eventAPIEvent.Type {
				case slackevents.CallbackEvent:
					switch evt := eventAPIEvent.InnerEvent.Data.(type) {
					case *slackevents.MessageEvent:
						// skip bot message
						if evt.BotID != "" {
							logger.Debug("this message sent by bot.")
							continue
						}

						logger.Debug("original text", zap.String("message", evt.Text))
						logger.Debug("event message information", zap.Any("type", reflect.TypeOf(evt)), zap.Any("info", evt))

						// skip image only, or delete image message
						if evt.Text == "" {
							logger.Debug("skip translate message, because of image only.")
							continue
						}

						// if URL contains in message, trancate it
						message, err := trancateText(evt.Text)
						if err != nil {
							logger.Error("failed to trancate text", zap.Error(err))
						}

						// if only english in message
						match := isJapanese(message)
						if match {
							// translate text via GoogleTranslate API
							translatedMessage, err := translateText(conf.ProjectID, "ja-jp", "en-us", message)
							if err != nil {
								logger.Error("failed to translate", zap.Error(err))
							}

							logger.Info("translate success!", zap.String("originalText", message), zap.String("translatedText", translatedMessage))

							// post slack message
							_, _, err = client.PostMessage(
								evt.Channel,
								slack.MsgOptionText(translatedMessage, false),
							)
							if err != nil {
								logger.Error("failed to post slack message", zap.Error(err))
							}
						}
					}
				}
			}
		}
	}()
	if err = socketClient.Run(); err != nil {
		logger.Error("failed to reconnect to slack", zap.Error(err))
	}
}
