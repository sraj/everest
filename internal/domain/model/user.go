package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// User represents a user in the system.
type User struct {
	ID        string    `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// UserProfile represents app-specific user data, keyed by Zitadel subject.
// Identity fields (name, email) are managed by Zitadel and loaded from /auth/me.
type UserProfile struct {
	UserID      string    `json:"user_id" db:"user_id"`
	Nickname    string    `json:"nickname" db:"nickname"`
	Preferences Prefs     `json:"preferences" db:"preferences"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Prefs is a key-value map stored as JSONB in PostgreSQL.
type Prefs map[string]string

func (p *Prefs) Scan(value any) error {
	if value == nil {
		*p = make(Prefs)
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("prefs: unsupported scan type")
	}
	return json.Unmarshal(b, (*map[string]string)(p))
}

func (p Prefs) Value() (driver.Value, error) {
	if p == nil {
		return []byte("{}"), nil
	}
	return json.Marshal((map[string]string)(p))
}
