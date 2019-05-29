package undo

type undoList struct {
	current *node
}

func newUndoList() undoList {
	head := node{}
	h := undoList{&head}
	return h
}

type node struct {
	action         ReversibleAction
	next, previous *node
}

func (l *undoList) addAction(a ReversibleAction) error {
	err := a.Do()
	if err != nil {
		return err
	}
	n := node{action: a, previous: l.current}
	l.current.next = &n
	l.current = &n
	return nil
}

func (l *undoList) undo() (bool, error) {
	if isHead(l.current) {
		return false, nil
	}
	if err := l.current.action.Undo(); err != nil {
		return false, err
	}
	l.current = l.current.previous
	return true, nil
}

func (l *undoList) redo() (bool, error) {
	if l.current.next == nil {
		return false, nil
	}
	l.current = l.current.next
	if err := l.current.action.Do(); err != nil {
		return false, err
	}
	return true, nil
}

func isHead(n *node) bool {
	return n.action == nil || n.previous == nil
}
