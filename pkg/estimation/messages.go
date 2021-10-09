package estimation

import (
	"bytes"
	"fmt"

	"github.com/ClydeSparrow/estimation-bot/internal/util"
)

const (
	VOTING_GREETING_MESSAGE  = "Voting started, please send me your estimation score. Or type \"skip\" to indicate that you are not joining estimation"
	VOTING_ESTIMATED_MESSAGE = "Voting is finished because SP difference is less than 3"
	UNKNOWN_COMMAND_MESSAGE  = "Sorry, I don't understand this command"
)

// ====================================

func StartMessages(title string) []Data {
	return []Data{
		{
			Key:     "public",
			Message: StartedMessage(title),
		},
	}
}

func FinishedVotingMessages(result VotingResult) []Data {
	messages := []Data{
		{
			Key:     "public",
			Message: StoppedMessage(result),
		},
	}

	if result.FinalScore > 0 {
		messages = append(messages,
			Data{
				Key:     "public",
				Message: VOTING_ESTIMATED_MESSAGE,
			},
			Data{
				Key:     "public",
				Message: fmt.Sprintf("Final score: %d", result.FinalScore),
			},
		)
	} else {
		messages = append(messages,
			Data{
				Key:     "public",
				Message: ScoreboardMessage(result.Scores),
			},
		)
	}

	return messages
}

// ====================================

func StatusMessage(voted, skipped int) string {
	b := new(bytes.Buffer)

	if voted > 0 && skipped > 0 {
		fmt.Fprintf(b, "%d voted / %d skipped", voted, skipped)
		return b.String()
	}
	if voted > 0 {
		fmt.Fprintf(b, "%d %s voted", voted, util.Pluralize("person", voted))
	}
	if skipped > 0 {
		fmt.Fprintf(b, "%d %s skipped", skipped, util.Pluralize("person", skipped))
	}

	return b.String()
}

func StartedMessage(title string) string {
	if title == "" {
		return "╔══════════════╗"
	} else {
		return fmt.Sprintf("╔═════%s═════╗", title)
	}
}

func StoppedMessage(result VotingResult) string {
	var borderLine string
	if result.Title == "" {
		borderLine = "╚══════════════╝"
	} else {
		borderLine = fmt.Sprintf("╚═════ %s ═════╝", result.Title)
	}
	return fmt.Sprintf("%s\n\nVoting stopped\nAvg Score: %.2f\n", borderLine, result.AvgScore)
}

func ScoreboardMessage(scores map[int][]string) string {
	b := new(bytes.Buffer)

	for score, voted := range scores {
		fmt.Fprintf(b, "%d: %d %s %v\n", score, len(voted), util.Pluralize("vote", len(voted)), voted)
	}
	fmt.Fprint(b, "\nLet's start discussion!")
	return b.String()
}

func ReadyToVoteMessage(readyToVote int) string {
	return fmt.Sprintf("%d %sready to vote", readyToVote, util.HideWhenMultiple("engineer ", readyToVote))
}

func RecapMessage(askedForRecap int) string {
	return fmt.Sprintf("%d %sasked for recap", askedForRecap, util.HideWhenMultiple("engineer ", askedForRecap))
}
