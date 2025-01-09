package fsm

import (
	"encoding/json"
	"sync"
)

type userStateStorage struct {
	mu   sync.RWMutex
	data map[int64]StateID
}

func newUserStateStorage() *userStateStorage {
	return &userStateStorage{
		data: make(map[int64]StateID),
	}
}

func (u *userStateStorage) Set(userID int64, stateID StateID) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.data[userID] = stateID

	return nil
}

func (u *userStateStorage) Exists(userID int64) (bool, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	_, ok := u.data[userID]

	return ok, nil
}

func (u *userStateStorage) Get(userID int64) (StateID, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	stateID, ok := u.data[userID]
	if !ok {
		return "", nil
	}

	return stateID, nil
}

func (u *userStateStorage) MarshalJSON() ([]byte, error) {
	u.mu.Lock()
	defer u.mu.Unlock()

	return json.Marshal(u.data)
}

func (u *userStateStorage) UnmarshalJSON(data []byte) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	return json.Unmarshal(data, &u.data)
}
