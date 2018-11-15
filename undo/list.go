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
	action ReversibleAction
	next, previous *node
}

func (l *undoList) addAction(a ReversibleAction) {
	n := node{action: a, previous: l.current}
	l.current.next = &n;
	l.current = &n
}

func (l *undoList) undo() error {
	if isHead(l.current) {
		return nil
	}
	err := l.current.action.Undo()
	l.current = l.current.previous
	return err
}

func (l *undoList) redo() error {
	if l.current.next != nil {
		l.current = l.current.next
		err := l.current.action.Do()
		return err
	}
	return nil
}

func isHead(n *node) bool {
	return n.action == nil || n.previous == nil
}
