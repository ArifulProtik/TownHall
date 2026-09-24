package profile

import (
	"strings"
	"time"

	"ArifulProtik/TownHall/ent"
	"ArifulProtik/TownHall/internal/auth"
)

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

// Length rules count characters, so normalize (trim) before Validate runs.
type UpdateRequest struct {
	Name      *string `json:"name,omitempty" validate:"omitempty,minrunes=2,maxrunes=100"`
	Bio       *string `json:"bio,omitempty" validate:"omitempty,maxrunes=280"`
	AvatarURL *string `json:"avatar_url,omitempty" validate:"omitempty,maxrunes=1000"`
	BannerURL *string `json:"banner_url,omitempty" validate:"omitempty,maxrunes=1000"`
	Location  *string `json:"location,omitempty" validate:"omitempty,maxrunes=100"`
	Website   *string `json:"website,omitempty" validate:"omitempty,maxrunes=200"`
}

// Normalize trims editable fields in place. response.Bind calls it between
// binding and validation so rune rules apply to the stored values.
func (r *UpdateRequest) Normalize() {
	for _, f := range []*string{r.Name, r.Bio, r.AvatarURL, r.BannerURL, r.Location, r.Website} {
		if f != nil {
			*f = strings.TrimSpace(*f)
		}
	}
}

// ToResponse is ToUserResponse plus profile extras. Email shows only to self.
func ToResponse(u *ent.User, viewerID string) Response {
	base := auth.ToUserResponse(u)
	isSelf := viewerID != "" && viewerID == u.ID
	email := ""
	if isSelf {
		email = u.Email
	}

	return Response{
		ID:             base.ID,
		Name:           base.Name,
		Email:          email,
		Username:       base.Username,
		Provider:       base.Provider,
		EmailVerified:  base.EmailVerified,
		Bio:            base.Bio,
		AvatarURL:      base.AvatarURL,
		BannerURL:      base.BannerURL,
		Location:       base.Location,
		Website:        base.Website,
		CreatedAt:      base.CreatedAt,
		FollowersCount: 0,
		FollowingCount: 0,
		PostsCount:     0,
		IsSelf:         isSelf,
	}
}
