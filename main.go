package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/ClydeSparrow/estimation-bot/pkg/common"
	"github.com/ClydeSparrow/estimation-bot/pkg/zoom"
)

const (
	BOT_NAME  = "EstimationBot"
	DEVICE_ID = "ad8ffee7-d47c-4357-9ac8-965ed64e96fc"

	MESSAGE_PREFIX     = "/"
	STATUS_UPDATE_TIME = 2 * time.Second
)

func main() {
	var meetingNumber = flag.String("meetingNumber", "", "Meeting number")
	var meetingPassword = flag.String("password", "", "Meeting password")

	flag.Parse()

	// get keys from environment
	apiKey := os.Getenv("ZOOM_JWT_API_KEY")
	apiSecret := os.Getenv("ZOOM_JWT_API_SECRET")

	// create the input channel that sends work to the goroutines
	commands := make(chan common.Data)
	// create the output channel that sends results back to the main function
	actions := make(chan common.Data)

	// Goroutine responsible for command handling
	go func(commands <-chan common.Data, actions chan<- common.Data) {
		voting := &Voting{}
		lastUpdate := time.Now().Unix()

		for data := range commands {
			// scheduled event to get voting status was received
			if data.Key == "timer" && data.Message == "status" && voting.IsStarted() && voting.StatusChanged(lastUpdate) {
				msg := StatusMessage(voting.Status())
				if msg != "" {
					actions <- common.Data{
						Key:     "public",
						Message: StatusMessage(voting.Status()),
					}
				}
				lastUpdate = time.Now().Unix()
				continue
			}

			if !strings.HasPrefix(data.Message, MESSAGE_PREFIX) {
				// USER REPLIES
				switch strings.ToLower(data.Message) {
				case "skip":
					voting.SkippedVote(data.Author.Name)
				default:
					if est, err := strconv.Atoi(data.Message); err == nil {
						voting.AddVote(data.Author.Name, est)
					}
				}
			} else {
				data.Message = strings.TrimPrefix(data.Message, MESSAGE_PREFIX)

				words := strings.Fields(data.Message)
				wordsCount := len(words)
				if wordsCount < 1 {
					// TODO: Error message?
					continue
				}
				args := words[1:]
				fmt.Printf("%+v\n", args)
				argsCount := len(args)

				// COMMANDS
				switch strings.ToLower(words[0]) {
				case "start":
					title := ""
					if argsCount > 0 {
						title = args[0]
					}
					voting, _ = NewVoting(title)

					actions <- common.Data{
						Key:     "public",
						Message: StartedMessage(voting.Title),
					}

					actions <- common.Data{
						Key:     "private",
						Message: VOTING_GREETING_MESSAGE,
					}
				case "stop":
					result := voting.Finish()

					actions <- common.Data{
						Key:     "public",
						Message: StoppedMessage(*result),
					}

					if result.FinalScore > 0 {
						actions <- common.Data{
							Key:     "public",
							Message: VOTING_ESTIMATED_MESSAGE,
						}

						actions <- common.Data{
							Key:     "public",
							Message: fmt.Sprintf("Final score: %d", result.FinalScore),
						}
					} else {
						actions <- common.Data{
							Key:     "public",
							Message: ScoreboardMessage(result.Scores),
						}
					}
				default:
					actions <- common.Data{
						Key:     "whisper",
						Author:  data.Author,
						Message: UNKNOWN_COMMAND_MESSAGE,
					}
				}
			}
		}
	}(commands, actions)

	if err := MakeConnection(*meetingNumber, *meetingPassword, apiKey, apiSecret, commands, actions); err != nil {
		log.Fatal(err)
	}

	close(commands)
	close(actions)
}

func MakeConnection(meetingNumber, meetingPassword, zoomApiKey, zoomApiSecret string, command chan common.Data, action <-chan common.Data) error {
	// TODO: I need another explanation
	done := make(chan struct{})

	// TODO: Good enough explanation
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	// Send time-based events to consumer
	votingStatusTicket := time.NewTicker(STATUS_UPDATE_TIME)
	defer votingStatusTicket.Stop()

	// create new session
	// meetingNumber, meetingPassword, username, hardware uuid (can be random but should be relatively constant or it will appear to zoom that you have many many many devices), proxy url, jwt api key, jwt api secret)
	session, err := zoom.NewZoomSession(meetingNumber, meetingPassword, BOT_NAME, DEVICE_ID, "", zoomApiKey, zoomApiSecret)
	if err != nil {
		log.Fatal(err)
	}
	// get the rwc token and other info needed to construct the websocket url for the meeting
	meetingInfo, cookieString, err := session.GetMeetingInfoData()
	if err != nil {
		log.Fatal(err)
	}

	// get the url for the websocket connection.  always pass false for the second parameter (its used internally to keep track of some parameters used for getting out of waiting rooms)
	websocketUrl, err := session.GetWebsocketUrl(meetingInfo, false)
	if err != nil {
		log.Fatal(err)
	}
	log.Print(websocketUrl)

	if err = session.MakeWebsocketConnection(websocketUrl, cookieString); err != nil {
		log.Fatal(err)
	}

	// Goroutine where all the action takes place
	go func(command chan<- common.Data) {
		defer close(done)

		if err := session.Listen(command); err != nil {
			log.Fatal(err)
		}
	}(command)

	for {
		select {
		case data := <-action:
			switch data.Key {
			case "public":
				session.SendChatMessage(zoom.EVERYONE_CHAT_ID, data.Message)
			case "private":
				session.SendPrivateMessageToEveryone(data.Message)
			case "whisper":
				session.SendChatMessage(data.Author.ID, data.Message)
			default:
				log.Fatal("That's very bad")
			}
		case <-votingStatusTicket.C:
			command <- common.Data{
				Key:     "timer",
				Message: "status",
			}
		case <-done:
			log.Println("+ Received done signal")
			return nil
		case <-interrupt:
			log.Println("+ Received interupt signal")
			return nil
		}
	}
}
