package main

import (
	"sort"
	"time"

	"github.com/ClydeSparrow/estimation-bot/internal/util"
)

type Voting struct {
	Title     string         // Ticket numebr
	Voted     map[string]int // Name -> Score
	Skipped   []string       // List of names who decided to skip
	createdAt int64          // When voting started
	updatedAt int64          // When latest vote was received
}

type VotingResult struct {
	Title      string
	AvgScore   float32
	Scores     map[int][]string
	FinalScore int
}

func NewVoting(title string) (*Voting, error) {
	v := Voting{
		Title:     title,
		Voted:     map[string]int{},
		Skipped:   []string{},
		createdAt: time.Now().Unix(),
		updatedAt: time.Now().Unix(),
	}
	return &v, nil
}

func (voting *Voting) IsStarted() bool {
	return voting.createdAt > 0
}

func (voting *Voting) AddVote(person string, vote int) {
	if !voting.IsStarted() {
		return
	}
	// TODO: Should we accept votes from persons who decided to skip before?
	voting.Voted[person] = vote
	voting.updatedAt = time.Now().Unix()
}

func (voting *Voting) SkippedVote(person string) {
	// TODO: If person decided to skip, remove his vote from estimations - ???
	if !voting.IsStarted() {
		return
	}
	if util.StringInSlice(person, voting.Skipped) {
		return
	}

	voting.Skipped = append(voting.Skipped, person)
	voting.updatedAt = time.Now().Unix()
}

func (voting *Voting) StatusChanged(since int64) bool {
	return voting.updatedAt >= since
}

func (voting *Voting) Status() (int, int) {
	return len(voting.Voted), len(voting.Skipped)
}

func (voting *Voting) Finish() *VotingResult {
	if len(voting.Voted) == 0 {
		return &VotingResult{}
	}

	result := VotingResult{
		Title: voting.Title,
	}
	result.Scores = make(map[int][]string)
	voteSum := 0

	for user, score := range voting.Voted {
		result.Scores[score] = append(result.Scores[score], user)
		voteSum += score
	}

	keys := make([]int, 0, len(result.Scores))
	for k := range result.Scores {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	result.AvgScore = float32(voteSum) / float32(len(voting.Voted))

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

	voting.Reset()
	return &result
}

func (voting *Voting) Reset() {
	voting.Title = ""
	voting.Voted = make(map[string]int)
	voting.Skipped = make([]string, 0)
	voting.createdAt = 0
	voting.updatedAt = 0
}
