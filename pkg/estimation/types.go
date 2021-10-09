package estimation

type Person struct {
	Name  string
	ID    int
	Score int

	Skipped       bool
	Ready         bool
	AskedForRecap bool
}

type Data struct {
	Key     string
	Author  Person
	Message string
}

// ==========================================

type Voting struct {
	Title string // Ticket numebr
	// Voted     map[string]int // Name -> Score
	// Skipped   []string       // List of names who decided to skip
	CreatedAt int64 // When estimation started
	UpdatedAt int64 // When latest vote was received

	peopleJoined map[int]Person
}

type VotingResult struct {
	Title      string
	AvgScore   float32
	Scores     map[int][]string
	FinalScore int
}
