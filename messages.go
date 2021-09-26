package main

import (
	"bytes"
	"fmt"

	"github.com/ClydeSparrow/estimation-bot/internal/util"
)

const (
	VOTING_GREETING_MESSAGE  = "Voting started, please send me your estimation score. Or type \"skip\" to indicate that you are not joining voting"
	VOTING_ESTIMATED_MESSAGE = "Voting is finished because SP difference is less than 3"
	UNKNOWN_COMMAND_MESSAGE  = "Sorry, I don't understand this command"
)

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
		return fmt.Sprintf("╔═══ %s ═══╗", title)
	}
}

func StoppedMessage(result VotingResult) string {
	fmt.Printf("%+v\n", result)
	var borderLine string
	if result.Title == "" {
		borderLine = "╚══════════════╝"
	} else {
		borderLine = fmt.Sprintf("╚═══ %s ═══╝", result.Title)
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
