package profile

import (
	"context"
	"strconv"
	"strings"
	"unicode/utf8"

	"ArifulProtik/TownHall/ent"
	"ArifulProtik/TownHall/ent/follow"
	"ArifulProtik/TownHall/internal/filestore"
	"ArifulProtik/TownHall/internal/userlookup"
	"ArifulProtik/TownHall/pkg/apperror"
)

// Uploader is what profiles need. *filestore.Storage satisfies it.
type Uploader interface {
	Upload(ctx context.Context, filename string, data []byte) (*filestore.Result, error)
}

type Service struct {
	db    *ent.Client
	store Uploader
}

func NewService(db *ent.Client, store Uploader) *Service {
	return &Service{db: db, store: store}
}

func (s *Service) UploadFile(ctx context.Context, filename string, data []byte) (*filestore.Result, error) {
	return s.store.Upload(ctx, filename, data)
}

func (s *Service) GetProfile(ctx context.Context, viewerID, handle string) (*Response, error) {
	u, err := userlookup.Resolve(ctx, s.db, viewerID, handle)
	if err != nil {
		return nil, err
	}

	resp := ToResponse(u, viewerID)
	resp.FollowersCount, resp.FollowingCount = s.FollowCounts(ctx, u.ID)
	return &resp, nil
}

// FollowCounts returns (followers, following) for userID, querying the Follow
// edge table directly (no domain-to-domain import). Stats never fail callers:
// errors degrade to zero so profile reads stay available.
func (s *Service) FollowCounts(ctx context.Context, userID string) (int, int) {
	followers, _ := s.db.Follow.Query().Where(follow.FollowingID(userID)).Count(ctx)
	following, _ := s.db.Follow.Query().Where(follow.FollowerID(userID)).Count(ctx)
	return followers, following
}

// textField is one editable profile field: its input, its length limits in
// characters, and how it lands on the update.
type textField struct {
	value *string
	min   int
	max   int
	name  string
	set   func(*ent.UserUpdateOne, string) *ent.UserUpdateOne
}

func (s *Service) UpdateProfile(ctx context.Context, userID string, req UpdateRequest) (*ent.User, error) {
	fields := []textField{
		{req.Name, 2, 100, "name", (*ent.UserUpdateOne).SetName},
		{req.Bio, 0, 280, "bio", (*ent.UserUpdateOne).SetBio},
		{req.AvatarURL, 0, 1000, "avatar url", (*ent.UserUpdateOne).SetAvatarURL},
		{req.BannerURL, 0, 1000, "banner url", (*ent.UserUpdateOne).SetBannerURL},
		{req.Location, 0, 100, "location", (*ent.UserUpdateOne).SetLocation},
		{req.Website, 0, 200, "website", (*ent.UserUpdateOne).SetWebsite},
	}

	update := s.db.User.UpdateOneID(userID)
	for _, f := range fields {
		if f.value == nil {
			continue
		}
		value := strings.TrimSpace(*f.value)
		n := utf8.RuneCountInString(value)
		if n < f.min || n > f.max {
			if f.min > 0 {
				return nil, apperror.BadRequest("name must be between 2 and 100 characters")
			}
			return nil, apperror.BadRequest(f.name + " must not exceed " + strconv.Itoa(f.max) + " characters")
		}
		update = f.set(update, value)
	}

	u, err := update.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperror.NotFound("user not found")
		}
		return nil, apperror.Internal()
	}
	return u, nil
}
