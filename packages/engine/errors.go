package engine

import "errors"

var (
	ErrInvalidTransition = errors.New("invalid simulation lifecycle transition")
	ErrEngineStopped     = errors.New("simulation engine is stopped")
	ErrEngineNotRunning  = errors.New("simulation engine is not running")
	ErrEventNotFound     = errors.New("scheduled event not found")
)
