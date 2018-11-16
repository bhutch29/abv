package undo

// Actor encapsulates all undo/redo functionality. Use NewHandler() to create an initialized Actor
type Actor struct {
	lists map[string]*undoList
}

// NewActor creates an initialized Actor
func NewActor() Actor {
	l := make(map[string]*undoList)
	h := Actor{l}
	return h
}

// AddAction performs the action and appends it onto the current node and updates the current node. Will destroy any history ahead of the current node.
func (h *Actor) AddAction(id string , a ReversibleAction) error {
	l := h.getList(id)
	err := l.addAction(a)
	return err
}

// Undo reverses out the current action and moves the current pointer back one action. If current action is the head, do nothing
func (h *Actor) Undo(id string) (bool, error) {
	l := h.getList(id)
	acted, err := l.undo()
	return acted, err
}

// Redo moves the current pointer ahead one action and performs it. If current action is the tail, do nothing
func (h *Actor) Redo(id string) error {
	l := h.getList(id)
	err := l.redo()
	return err
}

func (h *Actor) getList(id string) *undoList {
	if list, exists := h.lists[id]; exists {
		return list
	}
	l := newUndoList()
	h.lists[id] = &l
	return &l
}
