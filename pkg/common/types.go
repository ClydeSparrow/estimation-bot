package common

// import "github.com/ClydeSparrow/estimation-bot/pkg/zoom"

type Person struct {
	Name string
	ID   int
}

type Data struct {
	Key     string
	Author  Person
	Message string
}

// ==========================================

// type ScrumZoomSession zoom.ZoomSession

// ==========================================

type Voting struct {
	Title     string         // Ticket numebr
	Voted     map[string]int // Name -> Score
	Skipped   []string       // List of names who decided to skip
	CreatedAt int64          // When voting started
	UpdatedAt int64          // When latest vote was received
}

type VotingResult struct {
	Title      string
	AvgScore   float32
	Scores     map[int][]string
	FinalScore int
}
