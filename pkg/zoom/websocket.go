package zoom

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"

	"net/http"
	"net/url"
	"strconv"

	"github.com/ClydeSparrow/estimation-bot/pkg/estimation"
)

func (session *ZoomSession) GetWebsocketUrl(meetingInfo *MeetingInfo, wasInWaitingRoom bool) (string, error) {
	pingRwcServer := getRwgPingServer(meetingInfo)
	rwgInfo, err := session.getRwgPingData(meetingInfo, pingRwcServer)
	if err != nil {
		return "", err
	}

	if len(meetingInfo.Result.EncryptedRWC) < 1 {
		return "", errors.New("No RWC hosts found")
	}

	// query string for websocket url
	values := url.Values{}

	values.Set("rwcAuth", rwgInfo.RwcAuth)
	values.Set("dn2", base64.StdEncoding.EncodeToString([]byte(meetingInfo.Result.UserName)))
	values.Set("auth", meetingInfo.Result.Auth)
	values.Set("sign", meetingInfo.Result.Sign)
	values.Set("browser", USER_AGENT_SHORTHAND)
	values.Set("trackAuth", meetingInfo.Result.TrackAuth)
	values.Set("mid", meetingInfo.Result.Mid)
	values.Set("tid", meetingInfo.Result.Tid)
	values.Set("lang", "en")
	values.Set("ts", strconv.FormatInt(meetingInfo.Result.Ts, 10))
	values.Set("ZM-CID", session.HardwareID.String()) // this is a hardware id.  you shouldnt have it change a bunch of times per ip or you will look highly suspicious
	values.Set("_ZM_MTG_TRACK_ID", "")
	values.Set("jscv", "1.8.6")
	values.Set("fromNginx", "false")
	values.Set("zak", "")
	values.Set("mpwd", meetingInfo.Result.Password)
	values.Set("as_type", "1")

	// unknown
	values.Set("tk", "")
	values.Set("cfs", "0")
	// "opt" is a parameter to specify a meeting within a meeting, for instance breakout rooms or the main meeting in a meeting with waiting room enabled
	if wasInWaitingRoom {
		values.Set("opt", session.meetingOpt)
		values.Set("zoomid", session.JoinInfo.ZoomID)
		values.Set("participantID", strconv.Itoa(session.JoinInfo.ParticipantID))
	}

	return (&url.URL{
		Scheme:   "wss",
		Host:     rwgInfo.Rwg,
		Path:     fmt.Sprintf("/wc/api/%s", meetingInfo.Result.MeetingNumber),
		RawQuery: values.Encode(),
	}).String(), nil
}

func (session *ZoomSession) MakeWebsocketConnection(websocketUrl string, cookieString string) error {
	websocketHeaders := http.Header{}
	websocketHeaders.Set("Accept-Language", "en-US,en;q=0.9")
	websocketHeaders.Set("Cache-Control", "no-cache")
	websocketHeaders.Set("Origin", "http://localhost:9999")
	websocketHeaders.Set("Pragma", "no-cache")
	websocketHeaders.Set("User-Agent", USER_AGENT)
	websocketHeaders.Set("Cookie", cookieString)

	dialer := websocket.Dialer{
		// TODO: REMOVE -- DEV ONLY FOR CHARLES PROXY
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	if session.ProxyURL != nil {
		dialer.Proxy = http.ProxyURL(session.ProxyURL)
	}

	connection, _, err := dialer.Dial(websocketUrl, websocketHeaders)
	if err != nil {
		return err
	}

	session.websocketConnection = connection
	return nil
}

func (session *ZoomSession) Listen(command chan<- estimation.Data) error {
	defer session.websocketConnection.Close()
	done := make(chan struct{})

	var message *GenericZoomMessage
	go func() {
		defer close(done)
		for {
			// reset struct
			message = &GenericZoomMessage{}

			err := session.websocketConnection.ReadJSON(&message)
			if err != nil {
				log.Print("failed to read:", err)
				return
			}
			// log.Printf("Received message (Evt: %s = %d; Seq: %d): %s", MessageNumberToName[message.Evt], message.Evt, message.Seq, string(message.Body))

			if message.Evt == WS_CONF_JOIN_RES {
				bodyData := JoinConferenceResponse{}
				err := json.Unmarshal(message.Body, &bodyData)
				if err != nil {
					log.Print("Failed to unmarshal json: %+v", err)
					return
				}
				session.JoinInfo = bodyData
			}

			// convert generic json message to go type
			m, err := GetMessageBody(message)
			if err != nil {
				// log.Printf("Decoding message failed: %+v", err)
				continue
			}

			switch m := m.(type) {
			case *ConferenceRosterIndication:
				// we want to have correct list of all people in meeting to send private messages
				for _, person := range m.Add {
					// don't add ourself
					if person.ID != session.JoinInfo.UserID {
						command <- estimation.Data{
							Key: "zoom:add",
							Author: estimation.Person{
								ID:   person.ID,
								Name: string(person.Dn2),
							},
						}
					}
				}
				for _, person := range m.Remove {
					command <- estimation.Data{
						Key: "zoom:remove",
						Author: estimation.Person{
							ID: person.ID,
						},
					}
				}
				continue
			case *ConferenceChatIndication:
				command <- estimation.Data{
					Key: "zoom",
					Author: estimation.Person{
						Name: string(m.SenderName),
						ID:   m.DestNodeID,
					},
					Message: string(m.Text),
				}
			default:
				continue
			}
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	// zoom sends pings (aside from regular websocket ones) approximately every minute of the form "{"evt":0,"seq":74}"
	minutelyJsonPingTicker := time.NewTicker(60 * time.Second)
	defer minutelyJsonPingTicker.Stop()

	for {
		select {
		case <-minutelyJsonPingTicker.C:
			session.SendMessage(session.websocketConnection, WS_CONN_KEEPALIVE, nil)
		case <-done:
			return nil
		case <-interrupt:
			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := session.websocketConnection.WriteMessage(websocket.CloseMessage, []byte(""))
			if err != nil {
				return err
			}
			<-done
			return nil
		}
	}
}
