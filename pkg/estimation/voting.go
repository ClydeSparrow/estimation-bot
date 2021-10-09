package estimation

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"
)

var BLACKLIST_VOTING_NAMES = []string{
	"Alya Makharinsky",
	"Philipp Steinbeck",
	"Rachid Harrassi",
	"Vedika Prakash",
	"Oskar Salmhofer",
}

func NewVoting(args []string, people []Person) (*Voting, error) {
	title := ""
	if len(args) > 0 {
		title = args[0]
	}

	v := Voting{
		Title:     strings.ToUpper(title),
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	v.peopleJoined = make(map[int]Person)
	for _, person := range people {
		v.peopleJoined[person.ID] = person
	}
	return &v, nil
}

func (v *Voting) isAllowedToJoin(person string) bool {
	for _, name := range BLACKLIST_VOTING_NAMES {
		if person == name {
			return false
		}
	}
	return true
}

func (v *Voting) AddPerson(ID int, name string) error {
	if !v.isAllowedToJoin(name) {
		return fmt.Errorf("user %s is not allowed to join estimation", name)
	}
	if _, alreadyJoined := v.peopleJoined[ID]; alreadyJoined {
		return fmt.Errorf("%s already joined estimation", name)
	}

	v.peopleJoined[ID] = Person{ID: ID, Name: name}
	log.Printf("%+v\n", v.peopleJoined)
	return nil
}

func (v *Voting) RemovePerson(personLeft int) error {
	delete(v.peopleJoined, personLeft)
	log.Printf("%+v\n", v.peopleJoined)
	return nil
}

func (v *Voting) IsStarted() bool {
	return v.CreatedAt > 0
}

func (v *Voting) HasUpdates(since int64) bool {
	return v.UpdatedAt >= since
}

func (v *Voting) AddVote(personVoted int, vote int) error {
	// TODO: Should we accept votes from persons who decided to skip before?
	if !v.IsStarted() {
		return errors.New("voting is not started yet")
	}
	if person, joined := v.peopleJoined[personVoted]; joined {
		person.Score = vote
		v.peopleJoined[personVoted] = person
	} else {
		return fmt.Errorf("user with ID %d didn't join voting", personVoted)
	}

	v.UpdatedAt = time.Now().Unix()
	return nil
}

func (v *Voting) Skipped(personSkipped int) error {
	// TODO: If person decided to skip, remove his vote from estimations - ???
	if !v.IsStarted() {
		return errors.New("voting is not started yet")
	}
	if person, joined := v.peopleJoined[personSkipped]; joined {
		person.Skipped = true
		v.peopleJoined[personSkipped] = person
	} else {
		return fmt.Errorf("user with ID %d didn't join voting", personSkipped)
	}

	v.UpdatedAt = time.Now().Unix()
	return nil
}

func (v *Voting) Status() (int, int) {
	voted := 0
	skipped := 0

	for _, person := range v.peopleJoined {
		if person.Skipped {
			skipped++
		}
		if person.Score > 0 {
			voted++
		}
	}
	return voted, skipped
}

func (v *Voting) ListPeople() []Person {
	people := make([]Person, 0, len(v.peopleJoined))
	for _, value := range v.peopleJoined {
		people = append(people, value)
	}
	return people
}

func (v *Voting) Finish() VotingResult {
	result := VotingResult{Title: v.Title}

	voted, _ := v.Status()
	if voted == 0 {
		return result
	}

	result.Scores = make(map[int][]string)
	voteSum := 0

	for _, person := range v.peopleJoined {
		result.Scores[person.Score] = append(result.Scores[person.Score], person.Name)
		voteSum += person.Score
	}

	keys := make([]int, 0, len(result.Scores))
	for k := range result.Scores {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	result.AvgScore = float32(voteSum) / float32(voted)

	// TODO: Make this algorythm argument of .Finish()
	// If score difference is less than 3, we can make a final decision
	if keys[len(keys)-1]-keys[0] < 3 {
		count := 0
		score := 0

		for _, k := range keys {
			if len(result.Scores[k]) >= count {
				score = k
				count = len(result.Scores[k])
			}
		}

		result.FinalScore = score
	}

	v.Reset()
	return result
}

func (v *Voting) Ready(personReady int) (int, error) {
	peopleReady := 0
	for _, person := range v.peopleJoined {
		if person.Ready {
			peopleReady++
		}
	}

	if person, joined := v.peopleJoined[personReady]; joined {
		if person.Ready {
			return peopleReady, fmt.Errorf("user %s already ready to vote", person.Name)
		}

		person.Ready = true
		v.peopleJoined[personReady] = person
		peopleReady++
	} else {
		return peopleReady, fmt.Errorf("user with ID %d didn't join voting", personReady)
	}

	// This event doesn't affect status of voting, messages will be sent to chat independantly
	return peopleReady, nil
}

func (v *Voting) AskedForRecap(personAsked int) (int, error) {
	peopleAsked := 0
	for _, person := range v.peopleJoined {
		if person.AskedForRecap {
			peopleAsked++
		}
	}

	if person, joined := v.peopleJoined[personAsked]; joined {
		if person.AskedForRecap {
			return peopleAsked, fmt.Errorf("user %s already asked for recap", person.Name)
		}

		person.AskedForRecap = true
		v.peopleJoined[personAsked] = person
		peopleAsked++
	} else {
		return peopleAsked, fmt.Errorf("user with ID %d didn't join voting", personAsked)
	}

	// This event doesn't affect status of voting, messages will be sent to chat independantly
	return peopleAsked, nil
}

func (v *Voting) Reset() {
	for ID, person := range v.peopleJoined {
		v.peopleJoined[ID] = Person{
			ID:   ID,
			Name: person.Name,
		}
	}

	v.CreatedAt = 0
	v.UpdatedAt = 0
}
