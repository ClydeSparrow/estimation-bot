package zoom

import "github.com/ClydeSparrow/estimation-bot/pkg/common"

var BLACKLIST_VOTING_IDS = []int{
	16781312, // Alya Makharinsky
	16783360, // Philipp Steinbeck
	16789504, // Rachid Harrassi
	16790528, // Vedika Prakash
	16793600, // Oskar Salmhofer
}

func (session *ZoomSession) IsAllowedToJoin(personID int) bool {
	for _, id := range BLACKLIST_VOTING_IDS {
		if personID == id {
			return false
		}
	}
	return true
}

func (session *ZoomSession) SendChatMessage(destNodeID int, text string) error {
	sendBody := ConferenceChatRequest{
		DestNodeID: destNodeID,
		Sn:         []byte(session.JoinInfo.ZoomID),
		Text:       []byte(text),
	}
	err := session.SendMessage(session.websocketConnection, WS_CONF_CHAT_REQ, sendBody)
	if err != nil {
		return err
	}
	return nil
}

func (session *ZoomSession) SendPrivateMessageToEveryone(text string) error {
	// Send private messages to everyone in chat
	for _, person := range session.peopleJoined {
		err := session.SendChatMessage(person.ID, text)
		if err != nil {
			return err
		}
	}
	return nil
}

func (session *ZoomSession) AddPerson(newPersonID int, newPersonName string) []common.Person {
	for _, inCall := range session.peopleJoined {
		if inCall.ID == newPersonID {
			return session.peopleJoined
		}
	}
	session.peopleJoined = append(session.peopleJoined, common.Person{
		ID:   newPersonID,
		Name: newPersonName,
	})
	return session.peopleJoined
}

func (session *ZoomSession) RemovePerson(leftPersonID int) []common.Person {
	for idx, inCall := range session.peopleJoined {
		if inCall.ID == leftPersonID {
			session.peopleJoined[idx] = session.peopleJoined[len(session.peopleJoined)-1]
			session.peopleJoined = session.peopleJoined[:len(session.peopleJoined)-1]
			return session.peopleJoined
		}
	}
	return session.peopleJoined
}

// func (session *ZoomSession) StartVoting() error {
// 	// Clean voting information if exists
// 	session.voting.Reset()

// 	err := session.SendChatMessage(EVERYONE_CHAT_ID, "╔══════════════╗")
// 	if err != nil {
// 		return err
// 	}
// 	// Send private messages to everyone in chat
// 	for _, person := range session.peopleJoined {
// 		err := session.SendChatMessage(person.ID, "Voting has started, please send me your estimation score. Or type \"skip\" to indicate that you are not joining voting")
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// func (session *ZoomSession) StopVoting() error {
// 	session.voting.started = false

// 	votesByNames := make(map[int][]string)
// 	votesSum := 0

// 	for user, value := range session.voting.voted {
// 		votesByNames[value] = append(votesByNames[value], user)
// 		votesSum += value
// 	}
// 	keys := make([]int, 0, len(votesByNames))
// 	for k := range votesByNames {
// 		keys = append(keys, k)
// 	}
// 	sort.Ints(keys)

// 	b := new(bytes.Buffer)
// 	fmt.Fprint(b, "╚══════════════╝\n\nVoting stopped\n")
// 	fmt.Fprintf(b, "Avg Score: %.2f\n", float32(votesSum)/float32(len(session.voting.voted)))

// 	err := session.SendChatMessage(EVERYONE_CHAT_ID, b.String())
// 	if err != nil {
// 		return err
// 	}

// 	if len(session.voting.voted) == 0 {
// 		return nil
// 	}

// 	// If difference is less than 3, we can output final score for the ticket
// 	if keys[len(keys)-1]-keys[0] < 3 {
// 		maxScore := 0
// 		estimation := 0

// 		// Names don't matter when solution is already here
// 		b = new(bytes.Buffer)
// 		for _, k := range keys {
// 			fmt.Fprintf(b, "%d: %d %s\n", k, len(votesByNames[k]), util.Pluralize(len(votesByNames[k]), "vote"))

// 			if len(votesByNames[k]) >= maxScore {
// 				estimation = k
// 				maxScore = len(votesByNames[k])
// 			}
// 		}

// 		messages := []string{
// 			b.String(),
// 			"Voting is finished because SP difference is less than 3",
// 			fmt.Sprintf("Final estimation: %d", estimation),
// 		}

// 		for _, message := range messages {
// 			if err := session.SendChatMessage(EVERYONE_CHAT_ID, message); err != nil {
// 				return err
// 			}
// 		}
// 	} else {
// 		// If difference is more or equal to 3, discussion should be started, print votes by names
// 		b = new(bytes.Buffer)

// 		for _, k := range keys {
// 			fmt.Fprintf(b, "%d: %d %s %v\n", k, len(votesByNames[k]), util.Pluralize(len(votesByNames[k]), "vote"), votesByNames[k])
// 		}
// 		fmt.Fprint(b, "\nLet's start discussion!")

// 		err = session.SendChatMessage(EVERYONE_CHAT_ID, b.String())
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// func (session *ZoomSession) HandleMessage(body *ConferenceChatIndication) error {
// 	if session.voting.started {
// 		if string(body.Text) == strings.ToLower("skip") {
// 			session.voting.skipped = append(session.voting.skipped, string(body.SenderName))
// 			// Send message how many people skipped voting
// 			err := session.SendChatMessage(EVERYONE_CHAT_ID, fmt.Sprintf("%d %s skipped\n", len(session.voting.skipped), util.Pluralize(len(session.voting.skipped), "person")))
// 			if err != nil {
// 				return err
// 			}
// 		}
// 		// If int received, it's probably new vote
// 		if est, err := strconv.Atoi(string(body.Text)); err == nil {
// 			session.voting.voted[string(body.SenderName)] = est

// 			// Send message how many people voted
// 			err := session.SendChatMessage(EVERYONE_CHAT_ID, fmt.Sprintf("%d %s voted\n", len(session.voting.voted), util.Pluralize(len(session.voting.voted), "person")))
// 			if err != nil {
// 				return err
// 			}
// 		}
// 	}
// 	return nil
// }

// func (voting *Voting) Reset() {
// 	voting.started = true
// 	voting.skipped = nil
// 	voting.voted = make(map[string]int)
// }
