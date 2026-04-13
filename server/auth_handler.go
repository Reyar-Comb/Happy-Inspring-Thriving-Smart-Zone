package server

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success   bool   `json:"success"`
	SessionID string `json:"session_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *Server) HandleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	user, err := s.Users.Verify(req.Username, req.Password)
	if err != nil {
		json.NewEncoder(w).Encode(LoginResponse{
			Success: false,
			Error:   "Invalid username or password",
		})
		return
	}

	session := s.Sessions.Create(user.Username)
	json.NewEncoder(w).Encode(LoginResponse{
		Success:   true,
		SessionID: session.ID,
	})
}

func (s *Server) HandleRegister(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error processing password", http.StatusInternalServerError)
		return
	}

	err = s.Users.Register(req.Username, string(hash))
	if err != nil {
		json.NewEncoder(w).Encode(LoginResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	session := s.Sessions.Create(req.Username)

	json.NewEncoder(w).Encode(LoginResponse{
		Success:   true,
		SessionID: session.ID,
	})
}
