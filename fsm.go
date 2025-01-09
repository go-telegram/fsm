package fsm

import (
	"encoding/json"
	"fmt"
)

// StateID is a type for state identifier
type StateID string

// Callback is a function that will be called on state transition
type Callback func(f *FSM, args ...any)

// FSM is a finite state machine
type FSM struct {
	initialStateID   StateID
	callbacks        map[StateID]Callback
	userStateStorage UserStateStorage
	dataStorage      DataStorage
}

// UserStateStorage is an interface for user state storage
type UserStateStorage interface {
	Set(userID int64, stateID StateID) error
	Exists(userID int64) (bool, error)
	Get(userID int64) (StateID, error)
	MarshalJSON() ([]byte, error)
	UnmarshalJSON(data []byte) error
}

// DataStorage is an interface for data storage
type DataStorage interface {
	Put(userID int64, key, value any) error
	Get(userID int64, key any) (any, error)
	Delete(userID int64, key any) error
	MarshalJSON() ([]byte, error)
	UnmarshalJSON(data []byte) error
}

// New creates a new FSM
func New(initialStateName StateID, callbacks map[StateID]Callback, opts ...Option) *FSM {
	s := &FSM{
		initialStateID:   initialStateName,
		callbacks:        make(map[StateID]Callback),
		userStateStorage: newUserStateStorage(),
		dataStorage:      newDataStorage(),
	}

	for stateID, callback := range callbacks {
		s.callbacks[stateID] = callback
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// AddCallback adds a callback for a state
func (f *FSM) AddCallback(stateID StateID, callback Callback) {
	f.callbacks[stateID] = callback
}

// AddCallbacks adds callbacks for states
func (f *FSM) AddCallbacks(cb map[StateID]Callback) {
	for stateID, callback := range cb {
		f.callbacks[stateID] = callback
	}
}

// Transition transitions the user to a new state
func (f *FSM) Transition(userID int64, stateID StateID, args ...any) error {
	errSetState := f.userStateStorage.Set(userID, stateID)
	if errSetState != nil {
		return fmt.Errorf("failed to set user state: %w", errSetState)
	}

	cb, okCb := f.callbacks[stateID]
	if okCb {
		cb(f, args...)
	}

	return nil
}

// Current returns the current state of the user
func (f *FSM) Current(userID int64) (StateID, error) {
	stateExists, errCheckExists := f.userStateStorage.Exists(userID)
	if errCheckExists != nil {
		return "", fmt.Errorf("failed to check if user state exists: %w", errCheckExists)
	}

	if !stateExists {
		errSetState := f.userStateStorage.Set(userID, f.initialStateID)
		if errSetState != nil {
			return "", fmt.Errorf("failed to set initial state: %w", errSetState)
		}

		return f.initialStateID, nil
	}

	stateID, errGetState := f.userStateStorage.Get(userID)
	if errGetState != nil {
		return "", fmt.Errorf("failed to get user state: %w", errGetState)
	}

	return stateID, nil
}

// Reset resets the state of the user to the initial state
func (f *FSM) Reset(userID int64) error {
	return f.userStateStorage.Set(userID, f.initialStateID)
}

// MarshalJSON marshals the FSM to JSON
func (f *FSM) MarshalJSON() ([]byte, error) {
	type response struct {
		InitialStateID StateID `json:"initial_state_id"`
		UserStates     []byte  `json:"user_states_storage"`
		Storage        []byte  `json:"data_storage"`
	}

	dataStorageData, errMarshalData := f.dataStorage.MarshalJSON()
	if errMarshalData != nil {
		return nil, fmt.Errorf("failed to marshal data storage: %w", errMarshalData)
	}

	userStatesStorageData, errMarshalUserStates := f.userStateStorage.MarshalJSON()
	if errMarshalUserStates != nil {
		return nil, fmt.Errorf("failed to marshal user states: %w", errMarshalUserStates)
	}

	return json.Marshal(response{
		InitialStateID: f.initialStateID,
		UserStates:     userStatesStorageData,
		Storage:        dataStorageData,
	})
}

// UnmarshalJSON unmarshals the FSM from JSON
func (f *FSM) UnmarshalJSON(data []byte) error {
	type response struct {
		InitialStateID StateID `json:"initial_state_id"`
		UserStates     []byte  `json:"user_states_storage"`
		Storage        []byte  `json:"data_storage"`
	}

	var r response
	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}

	f.initialStateID = r.InitialStateID

	errUnmarshalData := f.dataStorage.UnmarshalJSON(r.Storage)
	if errUnmarshalData != nil {
		return fmt.Errorf("failed to unmarshal data storage: %w", errUnmarshalData)
	}

	errUnmarshalUserStates := f.userStateStorage.UnmarshalJSON(r.UserStates)
	if errUnmarshalUserStates != nil {
		return fmt.Errorf("failed to unmarshal user states: %w", errUnmarshalUserStates)
	}

	return nil
}

// Data returns the data storage
func (f *FSM) Data() DataStorage {
	return f.dataStorage
}
