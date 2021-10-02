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
