package fsm

import (
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	f := New("xxx", nil)
	if f == nil {
		t.Fatalf("expected not nil")
	}

	if f.initialStateID != "xxx" {
		t.Errorf("expected %q, got %q", "xxx", f.initialStateID)
	}
}

func TestNew_with_options(t *testing.T) {
	calledSet := false
	calledPut := false

	uStore := &UserStateStorageMock{
		SetFunc: func(userID int64, stateID StateID) error {
			calledSet = true
			return nil
		},
	}

	dStore := &DataStorageMock{
		PutFunc: func(userID int64, key any, value any) error {
			calledPut = true
			return nil
		},
	}

	f := New("xxx", nil, WithDataStorage(dStore), WithUserStateStorage(uStore))
	if f == nil {
		t.Fatalf("expected not nil")
	}

	f.userStateStorage.Set(1, "xxx")
	if !calledSet {
		t.Errorf("expected called")
	}

	f.dataStorage.Put(1, "key", "value")
	if !calledPut {
		t.Errorf("expected called")
	}
}

func TestFSM_AddCallback(t *testing.T) {
	f := New("", nil)
	f.AddCallback("xxx", func(f *FSM, args ...any) {})
	if f.callbacks["xxx"] == nil {
		t.Errorf("expected not nil")
	}
}

func TestFSM_AddCallbacks(t *testing.T) {
	f := New("", nil)
	f.AddCallbacks(map[StateID]Callback{
		"xxx": func(f *FSM, args ...any) {},
		"yyy": func(f *FSM, args ...any) {},
	})
	if f.callbacks["xxx"] == nil {
		t.Errorf("expected not nil")
	}
	if f.callbacks["yyy"] == nil {
		t.Errorf("expected not nil")
	}
}

func TestFSM_Transition(t *testing.T) {
	callbackCalled := false
	newStateID := StateID("")

	uStore := &UserStateStorageMock{
		SetFunc: func(userID int64, stateID StateID) error {
			newStateID = stateID
			return nil
		},
	}

	f := New("1", map[StateID]Callback{
		"2": func(f *FSM, args ...any) {
			callbackCalled = true
		},
	})

	f.userStateStorage = uStore

	err := f.Transition(1, "2")
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

	if !callbackCalled {
		t.Errorf("expected callback to be called")
	}

	if newStateID != "2" {
		t.Errorf("expected %q, got %q", "2", newStateID)
	}
}

func TestFSM_Transition_error(t *testing.T) {
	uStore := &UserStateStorageMock{
		SetFunc: func(userID int64, stateID StateID) error {
			return fmt.Errorf("error1")
		},
	}

	f := New("1", map[StateID]Callback{
		"2": func(f *FSM, args ...any) {},
	})

	f.userStateStorage = uStore

	err := f.Transition(1, "2")
	if err == nil {
		t.Fatalf("unexpected nil error")
	}

	if err.Error() != "failed to set user state: error1" {
		t.Errorf("expected %q, got %q", "failed to set user state: error1", err.Error())
	}
}

func TestFSM_Current_exists_error(t *testing.T) {
	uStore := &UserStateStorageMock{
		ExistsFunc: func(userID int64) (bool, error) {
			return false, fmt.Errorf("error1")
		},
	}

	f := New("1", nil)
	f.userStateStorage = uStore

	_, err := f.Current(1)
	if err == nil {
		t.Fatalf("unexpected nil error")
	}

	if err.Error() != "failed to check if user state exists: error1" {
		t.Errorf("expected %q, got %q", "failed to check if user state exists: error1", err.Error())
	}
}

func TestFSM_Current_not_exists_error(t *testing.T) {
	uStore := &UserStateStorageMock{
		ExistsFunc: func(userID int64) (bool, error) {
			return false, nil
		},
		SetFunc: func(userID int64, stateID StateID) error {
			return fmt.Errorf("error2")
		},
	}

	f := New("1", nil)
	f.userStateStorage = uStore

	_, err := f.Current(1)
	if err == nil {
		t.Fatalf("unexpected nil error")
	}

	if err.Error() != "failed to set initial state: error2" {
		t.Errorf("expected %q, got %q", "failed to set initial state: error2", err.Error())
	}
}

func TestFSM_Current_not_exists(t *testing.T) {
	newState := StateID("")

	uStore := &UserStateStorageMock{
		ExistsFunc: func(userID int64) (bool, error) {
			return false, nil
		},
		SetFunc: func(userID int64, stateID StateID) error {
			newState = stateID
			return nil
		},
	}

	f := New("1", nil)
	f.userStateStorage = uStore

	_, err := f.Current(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if newState != "1" {
		t.Errorf("expected %q, got %q", "1", newState)
	}
}

func TestFSM_Current_get_error(t *testing.T) {
	uStore := &UserStateStorageMock{
		ExistsFunc: func(userID int64) (bool, error) {
			return true, nil
		},
		GetFunc: func(userID int64) (StateID, error) {
			return "", fmt.Errorf("error3")
		},
	}

	f := New("1", nil)
	f.userStateStorage = uStore

	_, err := f.Current(1)
	if err == nil {
		t.Fatalf("unexpected nil error")
	}

	if err.Error() != "failed to get user state: error3" {
		t.Errorf("expected %q, got %q", "failed to get user state: error3", err.Error())
	}
}

func TestFSM_Current(t *testing.T) {
	uStore := &UserStateStorageMock{
		ExistsFunc: func(userID int64) (bool, error) {
			return true, nil
		},
		GetFunc: func(userID int64) (StateID, error) {
			return "2", nil
		},
	}

	f := New("1", nil)
	f.userStateStorage = uStore

	s, err := f.Current(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s != "2" {
		t.Errorf("expected %q, got %q", "2", s)
	}
}

func TestFSM_Reset_error(t *testing.T) {
	uStore := &UserStateStorageMock{
		SetFunc: func(userID int64, stateID StateID) error {
			return fmt.Errorf("error4")
		},
	}

	f := New("1", nil)
	f.userStateStorage = uStore

	err := f.Reset(1)
	if err == nil {
		t.Fatalf("unexpected nil error")
	}

	if err.Error() != "error4" {
		t.Errorf("expected %q, got %q", "error4", err.Error())
	}
}

func TestFSM_Reset(t *testing.T) {
	called := false

	uStore := &UserStateStorageMock{
		SetFunc: func(userID int64, stateID StateID) error {
			called = true
			return nil
		},
	}

	f := New("1", nil)
	f.userStateStorage = uStore

	err := f.Reset(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !called {
		t.Errorf("expected called")
	}
}
