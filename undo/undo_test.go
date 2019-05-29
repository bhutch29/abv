package undo

import (
	"reflect"
	"testing"
)

func TestAddAction(t *testing.T) {
	a := NewActor()
	d := &dummyAction{}
	err := a.AddAction("", d)
	if err != nil {
		t.Error(err)
	}
	checkResult(d.Calls, []string{"Do"}, t)
}

func TestUndoAction(t *testing.T) {
	a := NewActor()
	d := &dummyAction{}
	a.AddAction("", d)
	acted, err := a.Undo("")
	if err != nil {
		t.Error(err)
	}
	if !acted {
		t.Error("undo did not report that it acted")
	}
	checkResult(d.Calls, []string{"Do", "Undo"}, t)
}

func TestMultipleUndoLists(t *testing.T) {
	a := NewActor()
	d1 := &dummyAction{}
	d2 := &dummyAction{}
	a.AddAction("1", d1)
	a.AddAction("2", d2)
	a.Undo("1")
	checkResult(d1.Calls, []string{"Do", "Undo"}, t)
	checkResult(d2.Calls, []string{"Do"}, t)
}

func checkResult(have []string, want []string, t *testing.T) {
	if !reflect.DeepEqual(have, want) {
		t.Errorf("wanted calls %v got %v", want, have)
	}
}

type dummyAction struct {
	Calls []string
}

func (a *dummyAction) Do() error {
	a.Calls = append(a.Calls, "Do")
	return nil
}

func (a *dummyAction) Undo() error {
	a.Calls = append(a.Calls, "Undo")
	return nil
}
