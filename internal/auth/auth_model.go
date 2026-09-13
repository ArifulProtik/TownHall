package auth

import (
	"time"

	"ArifulProtik/TownHall/ent"
)

// SignupEmail is the signup-with-email request body.
type SignupEmail struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

// UserResponse is the safe public user shape (never includes the hash).
type UserResponse struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	Username      *string   `json:"username,omitempty"`
	Provider      string    `json:"provider"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
}

// ToUserResponse maps an Ent user to its public response shape.
func ToUserResponse(u *ent.User) UserResponse {
	return UserResponse{
		ID:            u.ID,
		Name:          u.Name,
		Email:         u.Email,
		Username:      &u.Username,
		Provider:      string(u.Provider),
		EmailVerified: u.EmailVerified,
		CreatedAt:     u.CreatedAt,
	}
}
