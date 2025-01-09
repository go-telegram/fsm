package fsm

// Option is a type for FSM options
type Option func(*FSM)

// WithDataStorage sets dataStorage for FSM
func WithDataStorage(storage DataStorage) Option {
	return func(f *FSM) {
		f.dataStorage = storage
	}
}

// WithUserStateStorage sets userStateStorage for FSM
func WithUserStateStorage(storage UserStateStorage) Option {
	return func(f *FSM) {
		f.userStateStorage = storage
	}
}
