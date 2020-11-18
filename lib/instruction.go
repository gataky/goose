package lib

import "strconv"

type action int

const (
	up action = iota
	down
	pending
	executed
	rollback
	redo
)

type Instructions struct {
	BatchHash   string
	ExcludeHash string
	LastHash    string
	Action      action

	Steps     int
	Direction int
}

func NewInstructions(action action, args ...string) *Instructions {

	var steps int
	if len(args) > 0 {
		steps, _ = strconv.Atoi(args[0])
	}

	var direction int
	switch action {
	case up, pending:
		direction = Up
	case down, executed, rollback, redo:
		direction = Down
	}

	return &Instructions{
		Steps:     steps,
		Direction: direction,
		Action:    action,
	}
}
