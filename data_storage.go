package fsm

import (
	"encoding/json"
	"sync"
)

type dataStorage struct {
	mu   sync.RWMutex
	data map[int64]map[any]any
}

func newDataStorage() *dataStorage {
	return &dataStorage{
		data: make(map[int64]map[any]any),
	}
}

func (m *dataStorage) Put(userID int64, key, value any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.data[userID]; !ok {
		m.data[userID] = make(map[any]any)
	}

	m.data[userID][key] = value

	return nil
}

func (m *dataStorage) Get(userID int64, key any) (any, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, ok := m.data[userID]; !ok {
		return nil, nil
	}

	return m.data[userID][key], nil
}

func (m *dataStorage) Delete(userID int64, key any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.data[userID]; !ok {
		return nil
	}

	delete(m.data[userID], key)

	return nil
}

func (m *dataStorage) MarshalJSON() ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return json.Marshal(m.data)
}

func (m *dataStorage) UnmarshalJSON(data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return json.Unmarshal(data, &m.data)
}
