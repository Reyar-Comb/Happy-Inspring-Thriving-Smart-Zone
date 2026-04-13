package server

import (
	"fmt"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username     string
	PasswordHash string
}

type UserStore struct {
	mu    sync.RWMutex
	users map[string]*User
}

func NewUserStore() *UserStore {
	return &UserStore{
		users: make(map[string]*User),
	}
}

func (u *UserStore) Register(username, passwordHash string) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	if _, exists := u.users[username]; exists {
		return fmt.Errorf("user already exists")
	}
	u.users[username] = &User{
		Username:     username,
		PasswordHash: passwordHash,
	}
	return nil
}

func (u *UserStore) Verify(username, password string) (*User, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	user, exists := u.users[username]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid password")
	}
	return user, nil
}
