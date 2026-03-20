package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/bugfixes/go-bugfixes/logs"
	ConfigBuilder "github.com/keloran/go-config"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	Config *ConfigBuilder.Config
	CTX    context.Context
}

type Session struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expiresAt"`
}

func New(config *ConfigBuilder.Config) *Auth {
	return &Auth{
		Config: config,
		CTX:    context.Background(),
	}
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (a *Auth) Login(username, password string) (*Session, error) {
	db, err := a.Config.Database.GetPGXClient(a.CTX)
	if err != nil {
		return nil, logs.Error(err)
	}
	defer db.Close(a.CTX)

	var userID int
	var passwordHash string
	err = db.QueryRow(a.CTX,
		"SELECT id, password_hash FROM admin_users WHERE username = $1",
		username,
	).Scan(&userID, &passwordHash)
	if err != nil {
		return nil, logs.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return nil, logs.Errorf("invalid credentials")
	}

	token, err := generateToken()
	if err != nil {
		return nil, logs.Error(err)
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	_, err = db.Exec(a.CTX,
		"INSERT INTO admin_sessions (user_id, token, expires_at) VALUES ($1, $2, $3)",
		userID, token, expiresAt,
	)
	if err != nil {
		return nil, logs.Error(err)
	}

	return &Session{
		Token:     token,
		ExpiresAt: expiresAt.Format(time.RFC3339),
	}, nil
}

func (a *Auth) ValidateToken(token string) (bool, error) {
	db, err := a.Config.Database.GetPGXClient(a.CTX)
	if err != nil {
		return false, logs.Error(err)
	}
	defer db.Close(a.CTX)

	var expiresAt time.Time
	err = db.QueryRow(a.CTX,
		"SELECT expires_at FROM admin_sessions WHERE token = $1",
		token,
	).Scan(&expiresAt)
	if err != nil {
		return false, nil
	}

	if time.Now().After(expiresAt) {
		// Clean up expired session
		_, _ = db.Exec(a.CTX, "DELETE FROM admin_sessions WHERE token = $1", token)
		return false, nil
	}

	return true, nil
}
