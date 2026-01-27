package sessions

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/reflective-technologies/kiosk-cli/internal/config"
)

// Store manages persistent session IDs per app.
type Store struct {
	path     string
	mu       sync.Mutex
	sessions map[string]string
}

// Load reads the session store from disk (or initializes an empty store if missing).
func Load() (*Store, error) {
	path := config.SessionsPath()
	sessions := make(map[string]string)

	data, err := os.ReadFile(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("read sessions: %w", err)
		}
		return &Store{path: path, sessions: sessions}, nil
	}

	if err := json.Unmarshal(data, &sessions); err != nil {
		return nil, fmt.Errorf("parse sessions: %w", err)
	}

	return &Store{path: path, sessions: sessions}, nil
}

// Get returns the session ID for an app key.
func (s *Store) Get(appKey string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id, ok := s.sessions[appKey]
	return id, ok
}

// GetOrCreate returns the session ID for an app key, creating one if needed.
func (s *Store) GetOrCreate(appKey string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if id, ok := s.sessions[appKey]; ok {
		return id, false, nil
	}

	id, err := newUUID()
	if err != nil {
		return "", false, err
	}

	s.sessions[appKey] = id
	if err := s.saveLocked(); err != nil {
		delete(s.sessions, appKey)
		return "", false, err
	}

	return id, true, nil
}

// Set stores a session ID for an app key.
func (s *Store) Set(appKey, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[appKey] = id
	return s.saveLocked()
}

// Delete removes the session ID for an app key.
func (s *Store) Delete(appKey string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.sessions[appKey]; !ok {
		return nil
	}

	delete(s.sessions, appKey)
	return s.saveLocked()
}

func (s *Store) saveLocked() error {
	if err := os.MkdirAll(config.KioskDir(), 0755); err != nil {
		return fmt.Errorf("create kiosk dir: %w", err)
	}

	data, err := json.MarshalIndent(s.sessions, "", "  ")
	if err != nil {
		return fmt.Errorf("encode sessions: %w", err)
	}

	if err := os.WriteFile(s.path, data, 0644); err != nil {
		return fmt.Errorf("write sessions: %w", err)
	}

	return nil
}
