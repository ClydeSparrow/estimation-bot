package main

import (
	"strings"
	"time"

	"github.com/ClydeSparrow/estimation-bot/pkg/common"
)

func NewVoting(title string) (*common.Voting, error) {
	v := common.Voting{
		Title:     strings.ToUpper(title),
		Voted:     map[string]int{},
		Skipped:   []string{},
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}
	return &v, nil
}
