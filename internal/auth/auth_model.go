package auth

import (
	"time"

	"ArifulProtik/TownHall/ent"
)

// max=72 on passwords is the bcrypt limit, not a product choice.
type SignupEmail struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

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

type CheckUsernameResponse struct {
	Available bool   `json:"available"`
	Reason    string `json:"reason,omitempty"`
}

type SetupUsernameRequest struct {
	Username string `json:"username" validate:"required,min=3,max=30,alphanum_underscore"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

// RefreshRaw is the only copy of the raw refresh value; only its hash is stored.
type TokenPair struct {
	AccessToken string
	AccessExp   time.Time
	RefreshRaw  string
	RefreshExp  time.Time
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required,min=8,max=72"`
	NewPassword     string `json:"new_password" validate:"required,min=8,max=72"`
}

// ToUserResponse drops the password hash; it must never reach JSON.
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

// TokenResponse carries the access token; refresh travels by cookie only.
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}
