package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/ichisuke55/translate-bot/config"

	translate "cloud.google.com/go/translate/apiv3"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	translatepb "google.golang.org/genproto/googleapis/cloud/translate/v3"
)

func translateText(projectID string, sourceLang string, targetLang string, text string) (string, error) {
	ctx := context.Background()
	client, err := translate.NewTranslationClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()

	req := &translatepb.TranslateTextRequest{
		Parent:             fmt.Sprintf("projects/%s/locations/global", projectID),
		SourceLanguageCode: sourceLang,
		TargetLanguageCode: targetLang,
		MimeType:           "text/plain", // Mime types: "text/plain", "text/html"
		Contents:           []string{text},
	}

	resp, err := client.TranslateText(ctx, req)
	if err != nil {
		return "[FATAL] cannot translate message", err
	}
	log.Println(resp.GetTranslations())

	// Display the translation for each input text provided
	var msg string
	for _, translation := range resp.GetTranslations() {
		msg = fmt.Sprintf("%v\n", translation.GetTranslatedText())
	}

	return msg, nil

}

func slackVerificationMiddleware(next http.HandlerFunc, signSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		verifier, err := slack.NewSecretsVerifier(r.Header, signSecret)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		bodyReader := io.TeeReader(r.Body, &verifier)
		body, err := ioutil.ReadAll(bodyReader)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := verifier.Ensure(); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		next.ServeHTTP(w, r)
	}
}

func trancateText(msg string) (string, error) {
	// slack's url text style is <URL>
	urlRegexp := `<(http:\/\/|https:\/\/)?[a-z0-9]+([\-\.]{1}[a-z0-9]+)*\.[a-z]{2,5}(:[0-9]{1,5})?(\/.*)?>`
	rep := regexp.MustCompile(urlRegexp)

	log.Printf("original text: %v\n", msg)
	// if URL contains, trancate it.
	match := rep.MatchString(msg)
	if match == true {
		msg = rep.ReplaceAllString(msg, "")
	}
	return msg, nil
}

func main() {

	// Load environment variables
	conf, err := config.NewEnvConfig()
	if err != nil {
		log.Fatal("cannot load environment variables")
	}

	// Set Slack API client
	client := slack.New(
		conf.SlackBotToken,
		slack.OptionAppLevelToken(conf.SlackAppToken),
		slack.OptionDebug(true),
		slack.OptionLog(log.New(os.Stdout, "apiClient: ", log.Lshortfile|log.LstdFlags)),
	)

	socketClient := socketmode.New(
		client,
		socketmode.OptionDebug(true),
		socketmode.OptionLog(log.New(os.Stdout, "socketClient: ", log.Lshortfile|log.LstdFlags)),
	)

	go func() {
		for event := range socketClient.Events {
			switch event.Type {
			case socketmode.EventTypeEventsAPI:
				eventAPIEvent, ok := event.Data.(slackevents.EventsAPIEvent)
				if !ok {
					log.Printf("Ignored %v\n", event)
					continue
				}
				// Prevent from sending duplicate message
				socketClient.Ack(*event.Request)

				switch eventAPIEvent.Type {
				case slackevents.CallbackEvent:
					switch evt := eventAPIEvent.InnerEvent.Data.(type) {
					case *slackevents.MessageEvent:
						if evt.BotID == "" {
							// if URL contains in message, trancate it
							message, err := trancateText(evt.Text)
							if err != nil {
								log.Println(err)
								return
							}
							// translate text via GoogleTranslate API
							message, err = translateText("gcp-dev-ichisuke", "ja-jp", "en-us", message)
							if err != nil {
								log.Println(err)
								return
							}
							_, _, err = client.PostMessage(
								evt.Channel,
								slack.MsgOptionText(message, false),
							)
							if err != nil {
								log.Println(err)
								return
							}
						}
					}

				}

			}
		}
	}()

	socketClient.Run()
}
