package undo

// Handler encapsulates all undo/redo functionality. Use NewHandler() to create an initialized Handler
type Handler struct {
	lists map[string]*undoList
}

// NewHandler creates an initialized Handler
func NewHandler() Handler {
	l := make(map[string]*undoList)
	h := Handler{l}
	return h
}

// AddAction appends a new action onto the current node and updates the current node. Will destroy any history ahead of the current node.
func (h *Handler) AddAction(id string , a ReversibleAction) {
	l := h.getList(id)
	l.addAction(a)
}

// Undo reverses out the current action and moves the current pointer back one action. If current action is the head, do nothing
func (h *Handler) Undo(id string , a ReversibleAction) error {
	l := h.getList(id)
	err := l.undo()
	return err
}

// Redo moves the current pointer ahead one action and performs it. If current action is the tail, do nothing
func (h *Handler) Redo(id string, a ReversibleAction) error {
	l := h.getList(id)
	err := l.redo()
	return err
}

func (h *Handler) getList(id string) *undoList {
	if list, exists := h.lists[id]; exists {
		return list
	}
	l := newUndoList()
	h.lists[id] = &l
	return &l
}
