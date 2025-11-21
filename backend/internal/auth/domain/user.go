package domain

import "time"

type User struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	Email        string    `json:"email" gorm:"uniqueIndex;not null"`
	Password     string    `json:"-"` // Never return password in JSON
	Name         string    `json:"name"`
	AvatarURL    string    `json:"avatar_url,omitempty"`
	Provider     string    `json:"provider"` // "email" or "google"
	AccessToken  string    `json:"-"` // Google access token (not returned in JSON)
	RefreshToken string    `json:"-"` // Google refresh token (not returned in JSON)
	TokenExpiry  time.Time `json:"-"` // When the access token expires
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type RefreshToken struct {
	Token     string    `json:"token" gorm:"primaryKey"`
	UserID    string    `json:"user_id" gorm:"index"`
	ExpiresAt time.Time `json:"expires_at"`
}
