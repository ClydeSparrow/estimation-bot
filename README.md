# estimation Bot
*Your small helper for Scrum ticket estimations during Zoom calls*


## Overview
Highly inspired and currently heavily copied from Inspired & copied from github.com/chris124567/zoomer. Yeah, simple import is not good enough now, because I re-wrote `MakeWebsocketConnection` function

Overall architecture was developed around two channels
- `commands`. Triggered both by Zoom events (someone joined, chat messages, etc) and time-based events
- `actions`. Subscriber of this channel receives events about ticket estimations and reacts in appropriate way
    - The main idea is to be able to use not only Zoom API, but anything in future

## TODO
### Commands
- `/start TEXT` for better output in chat. `TEXT` can be Jira ticket number (e.g. PE-1234)
- `/engineers N` to set minimal when automatic finishing of estimation is possible
- `/recap` asks for a recap
- `/ready` signals that an engineer is ready to estimate, outputs `user xyz is ready to estimate, X not ready`. If it reaches N set with the `/engineers` command outputs `all engineers are ready to estimate`
- Command to set deadline for estimation. Deadline should be apllied not only for one estimation, but until end of meeting

### Possible integrations
- [JIRA] Set ticket's story points
- [Slack] Forward estimated ticket to a slack channel(?) so that everyone is aware
- [Export] Metrics about votings (# voted/skipped, estimation result and duration, etc)

### Functionality
- Gracefully exit/disconnect
- General refactoring (maybe, for original code as well)
- Support for meetings where you don't have the password but just a Zoom url with the "pwd" parameter in it
- Support audio/video
- Automated tests

### Code Style
- Decision logic function (for example, give a final scope if SP diff < 3) as anonymous functions exported to `.Stop(...)` method
    - Idea: Make easier to define custom rules for estimation
- It's possible to use original library instead of copy-pasting it. Only `MakeWebsocketConnection` is implemented in new & more atomic way
- Refine code with an idea that there might be more clients than just Zoom
    - Separate Voting and ZoomSession objects

## Zoom WebSDK

This project was created by reverse engineering the Zoom Web SDK.  Regular web joins are captcha-gated but web SDK joins [are not](https://devforum.zoom.us/t/remove-recaptcha-on-webinars-websdk1-7-9/23054/25).  I use an API only used by the Web SDK to get tokens needed to join the meeting.  This means you need a Zoom API key/secret, specifically a JWT one.  These can be obtained on the Zoom [App Marketplace](https://marketplace.zoom.us/user/build) site.  The demo at `cmd/zoomer/main.go` reads these from the environment as `ZOOM_JWT_API_KEY` and `ZOOM_JWT_API_SECRET`.

### NOTE
Because the API keys are associated with your account, using this software may get your Zoom account banned (reverse engineering is against the Zoom Terms of Service). Please do not use this on an important account.

### INFORMATION ON PROTOCOL
The protocol used by the Zoom Web client is basically just JSON over Websockets.  The messages look something like this:

```
{"body":{"bCanUnmuteVideo":true},"evt":7938,"seq":44}
{"body":{"add":null,"remove":null,"update":[{"audio":"","bAudioUnencrytped":false,"id":16785408}]},"evt":7937,"seq":47}
{"body":{"add":null,"remove":null,"update":[{"caps":5,"id":16785408,"muted":true}]},"evt":7937,"seq":63}
{"body":{"dc":"the United States(SC)","network":"Zoom Global Network","region":"the United States"},"evt":7954,"seq":3}
```

The "evt" number specifies the event number.  There is a (mostly complete) list of these in `zoom/constant.go` that I extracted from javascript code on the meeting page.

For the above three messages, the types are:
```
WS_CONF_ATTRIBUTE_INDICATION                     = 7938 // ConferenceAttributeIndication
WS_CONF_ROSTER_INDICATION                        = 7937 // ConferenceRosterIndication
WS_CONF_DC_REGION_INDICATION                     = 7954 // ConferenceDCRegionIndication
```
The thing in the comments to the right is the struct type for that message, which can be found in `zoom/message_types.go`.

Also, the server and client both have sequence numbers ("seq") for the messages they send but it doesn't appear to be used for anything (?).


## FEATURES / SUPPORTED MESSAGE TYPES
| Feature                                                                                                            | Send/recv | Message Name                              | Function (if send) / struct type (if recv) | Host Required               | Tested |
| ------------------------------------------------------------------------------------------------------------------ | --------- | ----------------------------------------- | ------------------------------------------ | --------------------------- | ------ |
| Send a chat message                                                                                                | Send      | WS\_CONF\_CHAT\_REQ                       | ZoomSession.SendChatMessage                | No                          | Yes    |
| Pretend to "join audio"                                                                                            | Send      | WS\_AUDIO\_VOIP\_JOIN\_CHANNEL\_REQ       | ZoomSession.JoinAudioVoipChannel           | No                          | Yes    |
| Pretend to turn on/off video (if enabled camera indicator appears to be on but actually just shows a black screen) | Send      | WS\_VIDEO\_MUTE\_VIDEO\_REQ               | ZoomSession.SetVideoMuted                  | No                          | Yes    |
| Pretending to screen share (shows "x" is sharing their screen but is just a black screen)                          | Send      | WS\_CONF\_SET\_SHARE\_STATUS\_REQ         | ZoomSession.SetScreenShareMuted            | Depending on share settings | Yes    |
| Pretend to turn on/off audio (if enabled audio indicator appears to be on but no audio is actually outputted)      | Send      | WS\_AUDIO\_MUTE\_REQ                      | ZoomSession.SetAudioMuted                  | No                          | Yes    |
| Rename self                                                                                                        | Send      | WS\_CONF\_RENAME\_REQ                     | ZoomSession.RenameMe                       | Depending on settings       | Yes    |
| Rename others                                                                                                      | Send      | WS\_CONF\_RENAME\_REQ                     | ZoomSession.RenameById                     | Yes                         | No     |
| Request everyone mutes themselves                                                                                  | Send      | WS\_AUDIO\_MUTEALL\_REQ                   | ZoomSession.RequestAllMute                 | Yes                         | No     |
| Set mute upon entry status                                                                                         | Send      | WS\_CONF\_SET\_MUTE\_UPON\_ENTRY\_REQ     | ZoomSession.SetMuteUponEntry               | Yes                         | No     |
| Set allow unmuting audio                                                                                           | Send      | WS\_CONF\_ALLOW\_UNMUTE\_AUDIO\_REQ       | ZoomSesssion.SetAllowUnmuteAudio           | Yes                         | No     |
| Set allow participant renaming                                                                                     | Send      | WS\_CONF\_ALLOW\_PARTICIPANT\_RENAME\_REQ | ZoomSession.SetAllowParticipantRename      | Yes                         | No     |
| Set chat restrictions level                                                                                        | Send      | WS\_CONF\_CHAT\_PRIVILEDGE\_REQ           | ZoomSession.SetChatLevel                   | Yes                         | Yes    |
| Set screen sharing locked status                                                                                   | Send      | WS\_CONF\_LOCK\_SHARE\_REQ                | ZoomSession.SetShareLockedStatus           | Yes                         | No     |
| End meeting                                                                                                        | Send      | WS\_CONF\_END\_REQ                        | ZoomSession.EndMeeting                     | Yes                         | No     |
| Set allow unmuting video                                                                                           | Send      | WS\_CONF\_ALLOW\_UNMUTE\_VIDEO\_REQ       | ZoomSession.SetAllowUnmuteVideo            | Yes                         | No     |
| Request breakout room join token                                                                                   | Send      | WS\_CONF\_BO\_JOIN\_REQ                   | ZoomSession.RequestBreakoutRoomJoinToken   | No                          | Yes    |
| Breakout room broadcast                                                                                            | Send      | WS\_CONF\_BO\_BROADCAST\_REQ              | ZoomSession.BreakoutRoomBroadcast          | Yes                         | No     |
| Request a token for creation of a breakout room                                                                    | Send      | WS\_CONF\_BO\_TOKEN\_BATCH\_REQ           | ZoomSession.RequestBreakoutRoomToken       | Yes                         | Yes    |
| Create a breakout room                                                                                             | Send      | WS\_CONF\_BO\_START\_REQ                  | ZoomSession.CreateBreakoutRoom             | Yes                         | No     |
| Join information (user ID, participant ID and some other stuff)                                                    | Recv      | WS\_CONF\_JOIN\_RES                       | JoinConferenceResponse                     |                             | Yes    |
| Breakout room creation token response (response to WS\_CONF\_BO\_TOKEN\_BATCH\_REQ)                                | Recv      | WS\_CONF\_BO\_TOKEN\_RES                  | ConferenceBreakoutRoomTokenResponse        |                             | Yes    |
| Breakout room join response                                                                                        | Recv      | WS\_CONF\_BO\_JOIN\_RES                   | ConferenceBreakoutRoomJoinResponse         |                             | Yes    |
| Permission to show avatars changed                                                                                 | Recv      | WS\_CONF\_AVATAR\_PERMISSION\_CHANGED     | ConferenceAvatarPermissionChanged          |                             | Yes    |
| Roster change (mute/unmute, renames, leaves/joins)                                                                 | Recv      | WS\_CONF\_ROSTER\_INDICATION              | ConferenceRosterIndication                 |                             | Yes    |
| Meeting attribute setting (stuff like "is sharing allowed" or "is the meeting locked")                             | Recv      | WS\_CONF\_ATTRIBUTE\_INDICATION           | ConferenceAttributeIndication              |                             | Yes    |
| Host change                                                                                                        | Recv      | WS\_CONF\_HOST\_CHANGE\_INDICATION        | ConferenceHostChangeIndication             |                             | Yes    |
| Cohost change                                                                                                      | Recv      | WS\_CONF\_COHOST\_CHANGE\_INDICATION      | ConferenceCohostChangeIndication           |                             | Yes    |
| "Hold" state (waiting rooms)                                                                                       | Recv      | WS\_CONF\_HOLD\_CHANGE\_INDICATION        | ConferenceHoldChangeIndication             |                             | Yes    |
| Chat message                                                                                                       | Recv      | WS\_CONF\_CHAT\_INDICATION                | ConferenceChatIndication                   |                             | Yes    |
| Meeting "option" parameter (used for waiting room and breakout rooms)                                              | Recv      | WS\_CONF\_OPTION\_INDICATION              | ConferenceOptionIndication                 |                             | Yes    |
| ??? Local Record Indication ???                                                                                    | Recv      | WS\_CONF\_LOCAL\_RECORD\_INDICATION       | ConferenceLocalRecordIndication            |                             | Yes    |
| Breakout room command (forcing you to join a room, broadcasts)                                                     | Recv      | WS\_CONF\_BO\_COMMAND\_INDICATION         | ConferenceBreakoutRoomCommandIndication    |                             | Yes    |
| Breakout room attributes (settings and list of rooms)                                                              | Recv      | WS\_CONF\_BO\_ATTRIBUTE\_INDICATION       | ConferenceBreakoutRoomAttributeIndication  |                             | Yes    |
| Datacenter Region                                                                                                  | Recv      | WS\_CONF\_DC\_REGION\_INDICATION          | ConferenceDCRegionIndication               |                             | Yes    |
| ??? Audio Asn ???                                                                                                  | Recv      | WS\_AUDIO\_ASN\_INDICATION                | AudioAsnIndication                         |                             | Yes    |
| ??? Audio Ssrc ???                                                                                                 | Recv      | WS\_AUDIO\_SSRC\_INDICATION               | AudioSSRCIndication                        |                             | Yes    |
| Someone has enabled video                                                                                          | Recv      | WS\_VIDEO\_ACTIVE\_INDICATION             | VideoActiveIndication                      |                             | Yes    |
| ??? Video Ssrc ???                                                                                                 | Recv      | WS\_VIDEO\_SSRC\_INDICATION               | SSRCIndication                             |                             | Yes    |
| Someone is sharing their screen                                                                                    | Recv      | WS\_SHARING\_STATUS\_INDICATION           | SharingStatusIndication                    |                             | Yes    |


For sending: Look at `zoom/requests.go` and switch out the struct and message type names for your new message type

For receiving: Create a definition for the type and update the getPointerForBody function in `zoom/message.go.`
