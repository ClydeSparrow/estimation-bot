package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/ClydeSparrow/estimation-bot/pkg/estimation"
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
	commands := make(chan estimation.Data)
	// create the output channel that sends results back to the main function
	messages := make(chan estimation.Data)

	// Goroutine responsible for command handling
	// `commands` is a channel for received messages from Zoom, `messages` - for messages to be sent back
	go func(commands <-chan estimation.Data, messages chan<- estimation.Data) {
		voting, _ := estimation.NewVoting([]string{}, []estimation.Person{})
		lastUpdate := time.Now().Unix()

		for data := range commands {
			// Specific commands
			switch data.Key {
			case "zoom:add":
				voting.AddPerson(data.Author.ID, data.Author.Name)
				continue
			case "zoom:remove":
				voting.RemovePerson(data.Author.ID)
				continue
			case "timer:status":
				if voting.IsStarted() && voting.HasUpdates(lastUpdate) {
					if msg := estimation.StatusMessage(voting.Status()); msg != "" {
						messages <- estimation.Data{
							Key:     "public",
							Message: msg,
						}
					}
					lastUpdate = time.Now().Unix()
				}
				continue
			}

			// Perform action based on user non-command input
			if !strings.HasPrefix(data.Message, MESSAGE_PREFIX) {
				switch strings.ToLower(data.Message) {
				case "skip":
					if err := voting.Skipped(data.Author.ID); err != nil {
						log.Println(err)
					}
				default:
					est, err := strconv.Atoi(data.Message)
					if err != nil {
						log.Println("can't convert message to integer: ", data)
					}
					if err := voting.AddVote(data.Author.ID, est); err != nil {
						log.Println(err)
					}
				}
				continue
			}

			// COMMANDS
			data.Message = strings.TrimPrefix(data.Message, MESSAGE_PREFIX)

			words := strings.Fields(data.Message)
			if len(words) < 1 {
				// No command after "/" exist
				messages <- estimation.Data{
					Key:     "whisper",
					Author:  data.Author,
					Message: estimation.UNKNOWN_COMMAND_MESSAGE,
				}
				continue
			}
			args := words[1:]

			switch strings.ToLower(words[0]) {
			case "start":
				voting, _ = estimation.NewVoting(args, voting.ListPeople())

				for _, msg := range estimation.StartMessages(voting.Title) {
					messages <- msg
				}
				for _, person := range voting.ListPeople() {
					messages <- estimation.Data{
						Key: "whisper",
						Author: estimation.Person{
							ID: person.ID,
						},
						Message: estimation.VOTING_GREETING_MESSAGE,
					}
				}
			case "stop":
				result := voting.Finish()
				for _, msg := range estimation.FinishedVotingMessages(result) {
					messages <- msg
				}
			case "ready":
				peopleReady, err := voting.Ready(data.Author.ID)
				if err != nil {
					log.Println(err)
					continue
				}

				messages <- estimation.Data{
					Key:     "public",
					Message: estimation.ReadyToVoteMessage(peopleReady),
				}
			case "recap":
				peopleAsked, err := voting.AskedForRecap(data.Author.ID)
				if err != nil {
					log.Println(err)
					continue
				}

				messages <- estimation.Data{
					Key:     "public",
					Message: estimation.RecapMessage(peopleAsked),
				}
			default:
				messages <- estimation.Data{
					Key:     "whisper",
					Author:  data.Author,
					Message: estimation.UNKNOWN_COMMAND_MESSAGE,
				}
			}
		}
	}(commands, messages)

	if err := MakeConnection(*meetingNumber, *meetingPassword, apiKey, apiSecret, commands, messages); err != nil {
		log.Fatal(err)
	}

	close(commands)
	close(messages)
}

func MakeConnection(meetingNumber, meetingPassword, zoomApiKey, zoomApiSecret string, command chan estimation.Data, action <-chan estimation.Data) error {
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
	go func(command chan<- estimation.Data) {
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
			// case "private":
			// 	session.SendPrivateMessageToEveryone(data.Message)
			case "whisper":
				session.SendChatMessage(data.Author.ID, data.Message)
			default:
				log.Fatalf("That's very bad: %+v\n", data)
			}
		case <-votingStatusTicket.C:
			// TODO: dsdas
			command <- estimation.Data{Key: "timer:status"}
		case <-done:
			log.Println("+ Received done signal")
			return nil
		case <-interrupt:
			log.Println("+ Received interupt signal")
			return nil
		}
	}
}
