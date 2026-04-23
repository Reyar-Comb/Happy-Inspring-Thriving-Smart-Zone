package server

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"
)

type Session struct {
	ID         string
	Username   string
	PlayerID   int32
	CreatedAt  time.Time
	LastActive time.Time
}

type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
	}
}

func GenerateSessionID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func (sm *SessionManager) Create(username string) *Session {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session := &Session{
		ID:         GenerateSessionID(),
		Username:   username,
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
	}
	sm.sessions[session.ID] = session
	return session
}

func (sm *SessionManager) Get(sessionID string) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[sessionID]
	return session, exists
}

func (sm *SessionManager) Remove(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.sessions, sessionID)
}

func (sm *SessionManager) GetByPlayerID(playerID int32) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for _, session := range sm.sessions {
		if session.PlayerID == playerID {
			return session, true
		}
	}
	return nil, false
}
