package undo

// ReversibleAction defines an undo-able database change
type ReversibleAction interface {
	Do() error
	Undo() error
}
