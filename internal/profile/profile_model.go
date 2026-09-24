// Package profile provides user profile viewing, updating, and media uploads.
package profile

import (
	"strings"
	"time"

	"ArifulProtik/TownHall/ent"
)

// Response represents the full public or self profile.
type Response struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Email          string    `json:"email,omitempty"`
	Username       *string   `json:"username,omitempty"`
	Provider       string    `json:"provider"`
	EmailVerified  bool      `json:"email_verified"`
	Bio            string    `json:"bio,omitempty"`
	AvatarURL      string    `json:"avatar_url,omitempty"`
	BannerURL      string    `json:"banner_url,omitempty"`
	Location       string    `json:"location,omitempty"`
	Website        string    `json:"website,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	FollowersCount int       `json:"followers_count"`
	FollowingCount int       `json:"following_count"`
	PostsCount     int       `json:"posts_count"`
	IsSelf         bool      `json:"is_self"`
}

// UpdateRequest holds the editable profile fields.
//
// Length rules count characters (minrunes/maxrunes), matching the service
// and ent validators. Call normalize before Validate so rules apply to the
// trimmed values that are actually persisted.
type UpdateRequest struct {
	Name      *string `json:"name,omitempty" validate:"omitempty,minrunes=2,maxrunes=100"`
	Bio       *string `json:"bio,omitempty" validate:"omitempty,maxrunes=280"`
	AvatarURL *string `json:"avatar_url,omitempty" validate:"omitempty,maxrunes=1000"`
	BannerURL *string `json:"banner_url,omitempty" validate:"omitempty,maxrunes=1000"`
	Location  *string `json:"location,omitempty" validate:"omitempty,maxrunes=100"`
	Website   *string `json:"website,omitempty" validate:"omitempty,maxrunes=200"`
}

// normalize trims editable fields in place so validation and persistence
// agree on the stored values.
func (r *UpdateRequest) normalize() {
	for _, f := range []*string{r.Name, r.Bio, r.AvatarURL, r.BannerURL, r.Location, r.Website} {
		if f != nil {
			*f = strings.TrimSpace(*f)
		}
	}
}

// UploadResponse returns the result of a file upload to UploadThing or local storage.
type UploadResponse struct {
	URL  string `json:"url"`
	Key  string `json:"key,omitempty"`
	Name string `json:"name,omitempty"`
	Size int64  `json:"size,omitempty"`
}

// ToResponse transforms an ent.User into a profile Response.
func ToResponse(u *ent.User, viewerID string) Response {
	var username *string
	if u.Username != "" {
		username = &u.Username
	}
	isSelf := viewerID != "" && viewerID == u.ID
	email := ""
	if isSelf {
		email = u.Email
	}

	return Response{
		ID:             u.ID,
		Name:           u.Name,
		Email:          email,
		Username:       username,
		Provider:       string(u.Provider),
		EmailVerified:  u.EmailVerified,
		Bio:            u.Bio,
		AvatarURL:      u.AvatarURL,
		BannerURL:      u.BannerURL,
		Location:       u.Location,
		Website:        u.Website,
		CreatedAt:      u.CreatedAt,
		FollowersCount: 0,
		FollowingCount: 0,
		PostsCount:     0,
		IsSelf:         isSelf,
	}
}
