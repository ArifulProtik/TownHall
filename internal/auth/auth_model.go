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
	Bio           string    `json:"bio,omitempty"`
	AvatarURL     string    `json:"avatar_url,omitempty"`
	BannerURL     string    `json:"banner_url,omitempty"`
	Location      string    `json:"location,omitempty"`
	Website       string    `json:"website,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// CheckUsernameResponse reports whether a username is available.
type CheckUsernameResponse struct {
	Available bool   `json:"available"`
	Reason    string `json:"reason,omitempty"`
}

// SetupUsernameRequest holds the request to set an initial username.
type SetupUsernameRequest struct {
	Username string `json:"username" validate:"required,min=3,max=30,alphanum_underscore"`
}

// LoginRequest is the login request body.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

// TokenPair is an issued access + refresh token set with expiries.
// RefreshRaw is the only copy of the raw refresh value — hash before storing.
type TokenPair struct {
	AccessToken string
	AccessExp   time.Time
	RefreshRaw  string
	RefreshExp  time.Time
}

// ChangePasswordRequest holds current and new passwords for updating credentials.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required,min=8,max=72"`
	NewPassword     string `json:"new_password" validate:"required,min=8,max=72"`
}

// ToUserResponse maps an Ent user to its public response shape.
func ToUserResponse(u *ent.User) UserResponse {
	var username *string
	if u.Username != "" {
		username = &u.Username
	}
	return UserResponse{
		ID:            u.ID,
		Name:          u.Name,
		Email:         u.Email,
		Username:      username,
		Provider:      string(u.Provider),
		EmailVerified: u.EmailVerified,
		Bio:           u.Bio,
		AvatarURL:     u.AvatarURL,
		BannerURL:     u.BannerURL,
		Location:      u.Location,
		Website:       u.Website,
		CreatedAt:     u.CreatedAt,
	}
}

// TokenResponse carries an access token; the refresh token travels by cookie only.
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}
