package social

import (
	"context"
	"encoding/base64"
	"strings"

	"ArifulProtik/TownHall/ent"
	"ArifulProtik/TownHall/ent/follow"
	"ArifulProtik/TownHall/ent/user"
	"ArifulProtik/TownHall/pkg/apperror"
)

const (
	defaultListLimit = 20
	maxListLimit     = 100
)

// Service owns follow edges. Friendship is derived (mutual edges), never stored.
type Service struct {
	db *ent.Client
}

func NewService(db *ent.Client) *Service {
	return &Service{db: db}
}

// resolveTarget maps handle -> user: trim, "me" -> viewer (401 when anonymous),
// username (lowercased) first, id second, else 404.
func (s *Service) resolveTarget(ctx context.Context, viewerID, target string) (*ent.User, error) {
	clean := strings.TrimSpace(target)
	if clean == "" {
		return nil, apperror.BadRequest("handle is required")
	}
	if strings.EqualFold(clean, "me") {
		if viewerID == "" {
			return nil, apperror.Unauthorized("unauthorized")
		}
		clean = viewerID
	}
	u, err := s.db.User.Query().Where(user.UsernameEQ(strings.ToLower(clean))).Only(ctx)
	if err == nil {
		return u, nil
	}
	if !ent.IsNotFound(err) {
		return nil, apperror.Internal()
	}
	u, err = s.db.User.Get(ctx, clean)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperror.NotFound("user not found")
		}
		return nil, apperror.Internal()
	}
	return u, nil
}

func (s *Service) edgeExists(ctx context.Context, followerID, followingID string) bool {
	n, err := s.db.Follow.Query().
		Where(follow.FollowerID(followerID), follow.FollowingID(followingID)).
		Count(ctx)
	return err == nil && n > 0
}

func (s *Service) counts(ctx context.Context, userID string) (followers, following int) {
	followers, _ = s.db.Follow.Query().Where(follow.FollowingID(userID)).Count(ctx)
	following, _ = s.db.Follow.Query().Where(follow.FollowerID(userID)).Count(ctx)
	return followers, following
}

func (s *Service) statusFor(ctx context.Context, viewerID string, target *ent.User) *StatusResponse {
	isSelf := viewerID != "" && viewerID == target.ID
	var isFollowing, isFollowedBy bool
	if viewerID != "" && !isSelf {
		isFollowing = s.edgeExists(ctx, viewerID, target.ID)
		isFollowedBy = s.edgeExists(ctx, target.ID, viewerID)
	}
	followers, following := s.counts(ctx, target.ID)
	return &StatusResponse{
		IsFollowing:    isFollowing,
		IsFollowedBy:   isFollowedBy,
		IsFriend:       isFollowing && isFollowedBy,
		IsSelf:         isSelf,
		FollowersCount: followers,
		FollowingCount: following,
	}
}

// Follow creates follower actorID -> target. Idempotent: duplicate follow is success.
func (s *Service) Follow(ctx context.Context, actorID, target string) (*StatusResponse, error) {
	if strings.TrimSpace(actorID) == "" {
		return nil, apperror.Unauthorized("unauthorized")
	}
	t, err := s.resolveTarget(ctx, actorID, target)
	if err != nil {
		return nil, err
	}
	if t.ID == actorID {
		return nil, apperror.BadRequest("cannot follow yourself")
	}
	_, err = s.db.Follow.Create().
		SetFollowerID(actorID).
		SetFollowingID(t.ID).
		Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return s.statusFor(ctx, actorID, t), nil
		}
		return nil, apperror.Internal()
	}
	return s.statusFor(ctx, actorID, t), nil
}

// Unfollow deletes the edge if present. Idempotent.
func (s *Service) Unfollow(ctx context.Context, actorID, target string) (*StatusResponse, error) {
	if strings.TrimSpace(actorID) == "" {
		return nil, apperror.Unauthorized("unauthorized")
	}
	t, err := s.resolveTarget(ctx, actorID, target)
	if err != nil {
		return nil, err
	}
	if t.ID == actorID {
		return nil, apperror.BadRequest("cannot unfollow yourself")
	}
	id, err := s.db.Follow.Query().
		Where(follow.FollowerID(actorID), follow.FollowingID(t.ID)).
		OnlyID(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return s.statusFor(ctx, actorID, t), nil
		}
		return nil, apperror.Internal()
	}
	if err := s.db.Follow.DeleteOneID(id).Exec(ctx); err != nil {
		return nil, apperror.Internal()
	}
	return s.statusFor(ctx, actorID, t), nil
}

// GetStatus is public: anonymous viewers get is_*=false with real counts.
func (s *Service) GetStatus(ctx context.Context, viewerID, target string) (*StatusResponse, error) {
	t, err := s.resolveTarget(ctx, viewerID, target)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(viewerID) == "" {
		followers, following := s.counts(ctx, t.ID)
		return &StatusResponse{FollowersCount: followers, FollowingCount: following}, nil
	}
	return s.statusFor(ctx, viewerID, t), nil
}

// FollowersCount returns how many users follow userID. Counts never fail callers.
func (s *Service) FollowersCount(ctx context.Context, userID string) int {
	n, _ := s.db.Follow.Query().Where(follow.FollowingID(userID)).Count(ctx)
	return n
}

// FollowingCount returns how many users userID follows.
func (s *Service) FollowingCount(ctx context.Context, userID string) int {
	n, _ := s.db.Follow.Query().Where(follow.FollowerID(userID)).Count(ctx)
	return n
}

func clampLimit(limit int) int {
	if limit <= 0 {
		return defaultListLimit
	}
	if limit > maxListLimit {
		return maxListLimit
	}
	return limit
}

func encodeCursor(id string) string {
	return base64.URLEncoding.EncodeToString([]byte(id))
}

func decodeCursor(cursor string) (string, bool) {
	if cursor == "" {
		return "", false
	}
	raw, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil || len(raw) == 0 {
		return "", false
	}
	return string(raw), true
}

// listEdges pages follow edges ordered by edge ID. Edge IDs are UUIDv7
// (see BaseMixin), so lexicographic order is creation order: new edges sort
// last and pagination stays stable under inserts. ID-only cursors also dodge
// timestamp serialization mismatches between drivers (sqlite compares
// datetimes as strings, so a UTC cursor never matches local-zone rows).
func (s *Service) listEdges(ctx context.Context, preds []func(*ent.FollowQuery) *ent.FollowQuery, limit int, cursor string) ([]*ent.Follow, bool, error) {
	limit = clampLimit(limit)
	q := s.db.Follow.Query()
	for _, p := range preds {
		q = p(q)
	}
	if cursorID, ok := decodeCursor(cursor); ok {
		q = q.Where(follow.IDGT(cursorID))
	}
	edges, err := q.Order(ent.Asc(follow.FieldID)).Limit(limit + 1).All(ctx)
	if err != nil {
		return nil, false, apperror.Internal()
	}
	hasMore := len(edges) > limit
	if hasMore {
		edges = edges[:limit]
	}
	return edges, hasMore, nil
}

func (s *Service) usersByIDs(ctx context.Context, ids []string) (map[string]*ent.User, error) {
	users, err := s.db.User.Query().Where(user.IDIn(ids...)).All(ctx)
	if err != nil {
		return nil, apperror.Internal()
	}
	out := make(map[string]*ent.User, len(users))
	for _, u := range users {
		out[u.ID] = u
	}
	return out, nil
}

func edgesNextCursor(edges []*ent.Follow) string {
	if len(edges) == 0 {
		return ""
	}
	last := edges[len(edges)-1]
	return encodeCursor(last.ID)
}

// ListFollowers pages users following target.
func (s *Service) ListFollowers(ctx context.Context, target string, limit int, cursor string) (*ListResponse, error) {
	t, err := s.resolveTarget(ctx, "", target)
	if err != nil {
		return nil, err
	}
	edges, hasMore, err := s.listEdges(ctx, []func(*ent.FollowQuery) *ent.FollowQuery{
		func(q *ent.FollowQuery) *ent.FollowQuery { return q.Where(follow.FollowingID(t.ID)) },
	}, limit, cursor)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(edges))
	for _, e := range edges {
		ids = append(ids, e.FollowerID)
	}
	byID, err := s.usersByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	users := make([]ListUser, 0, len(edges))
	for _, id := range ids {
		if u, ok := byID[id]; ok {
			users = append(users, toListUser(u))
		}
	}
	resp := &ListResponse{Users: users, HasMore: hasMore}
	if hasMore {
		resp.NextCursor = edgesNextCursor(edges)
	}
	return resp, nil
}

// ListFollowing pages users target follows.
func (s *Service) ListFollowing(ctx context.Context, target string, limit int, cursor string) (*ListResponse, error) {
	t, err := s.resolveTarget(ctx, "", target)
	if err != nil {
		return nil, err
	}
	edges, hasMore, err := s.listEdges(ctx, []func(*ent.FollowQuery) *ent.FollowQuery{
		func(q *ent.FollowQuery) *ent.FollowQuery { return q.Where(follow.FollowerID(t.ID)) },
	}, limit, cursor)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(edges))
	for _, e := range edges {
		ids = append(ids, e.FollowingID)
	}
	byID, err := s.usersByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	users := make([]ListUser, 0, len(edges))
	for _, id := range ids {
		if u, ok := byID[id]; ok {
			users = append(users, toListUser(u))
		}
	}
	resp := &ListResponse{Users: users, HasMore: hasMore}
	if hasMore {
		resp.NextCursor = edgesNextCursor(edges)
	}
	return resp, nil
}

// maxFriendScanPages bounds one ListFriends call: each page costs one edge
// query plus one batched mutual-set query, never per-candidate round-trips.
const maxFriendScanPages = 5

// ListFriends pages mutual follows of target: users both following and followed by target.
func (s *Service) ListFriends(ctx context.Context, target string, limit int, cursor string) (*ListResponse, error) {
	t, err := s.resolveTarget(ctx, "", target)
	if err != nil {
		return nil, err
	}
	limit = clampLimit(limit)
	// Scan following-edge pages (bounded), keeping mutuals in edge order.
	// hasMore means "more edges remain unexamined", never a promise that more
	// mutuals exist: filtered keyset pagination can legitimately end with a
	// sparse page.
	var mutualIDs []string
	next := cursor
	moreEdges := false
	for i := 0; i < maxFriendScanPages && len(mutualIDs) < limit; i++ {
		edges, more, err := s.listEdges(ctx, []func(*ent.FollowQuery) *ent.FollowQuery{
			func(q *ent.FollowQuery) *ent.FollowQuery { return q.Where(follow.FollowerID(t.ID)) },
		}, limit, next)
		if err != nil {
			return nil, err
		}
		if len(edges) > 0 {
			next = edgesNextCursor(edges)
		}
		moreEdges = more
		candidates := make([]string, 0, len(edges))
		for _, e := range edges {
			candidates = append(candidates, e.FollowingID)
		}
		mutualSet := map[string]bool{}
		if len(candidates) > 0 {
			rows, err := s.db.Follow.Query().
				Where(follow.FollowerIDIn(candidates...), follow.FollowingID(t.ID)).
				Select(follow.FieldFollowerID).
				Strings(ctx)
			if err != nil {
				return nil, apperror.Internal()
			}
			for _, id := range rows {
				mutualSet[id] = true
			}
		}
		for _, id := range candidates {
			if mutualSet[id] {
				mutualIDs = append(mutualIDs, id)
				if len(mutualIDs) == limit {
					break
				}
			}
		}
		if !more {
			break
		}
	}
	byID, err := s.usersByIDs(ctx, mutualIDs)
	if err != nil {
		return nil, err
	}
	users := make([]ListUser, 0, len(mutualIDs))
	for _, id := range mutualIDs {
		if u, ok := byID[id]; ok {
			users = append(users, toListUser(u))
		}
	}
	resp := &ListResponse{Users: users, HasMore: moreEdges}
	if moreEdges {
		resp.NextCursor = next
	}
	return resp, nil
}
