package domain

import "time"

const (
	StateNone = iota
	StateAwaitingProblemReport
	StateAwaitingSuggestion
)

type UserState struct {
	State      int
	Text       string
	Photos     []interface{}
	PhotoCount int
	CreatedAt  time.Time
}
