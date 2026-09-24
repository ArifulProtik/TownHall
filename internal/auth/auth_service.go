package auth

import (
	"context"
	"net/http"
	"strings"
	"time"

	"ArifulProtik/TownHall/ent"
	"ArifulProtik/TownHall/ent/refreshtoken"
	"ArifulProtik/TownHall/ent/user"
	"ArifulProtik/TownHall/pkg/apperror"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	db         *ent.Client
	limiter    Limiter
	secret     string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewService(db *ent.Client, secret string, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{db: db, secret: secret, accessTTL: accessTTL, refreshTTL: refreshTTL, limiter: NewMemoryLimiter()}
}

func (s *Service) SignupEmail(ctx context.Context, in SignupEmail) (*ent.User, error) {
	name := strings.TrimSpace(in.Name)
	email := strings.ToLower(strings.TrimSpace(in.Email))

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), BcryptCost)
	if err != nil {
		return nil, apperror.Internal()
	}

	u, err := s.db.User.
		Create().
		SetName(name).
		SetEmail(email).
		SetPassword(string(hash)).
		SetProvider(user.ProviderEmail).
		Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, apperror.Conflict("email already registered")
		}
		return nil, apperror.Internal()
	}
	return u, nil
}

func (s *Service) AllowAttempt(email, ip string) (bool, time.Duration) {
	return s.limiter.Allow(strings.ToLower(strings.TrimSpace(email)), ip)
}

func (s *Service) Login(ctx context.Context, in LoginRequest) (*TokenPair, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))

	u, err := s.db.User.Query().Where(user.EmailEQ(email)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			// Match the cost of a real compare so unknown emails
			// take as long as wrong passwords.
			_ = bcrypt.CompareHashAndPassword([]byte(dummyPasswordHash), []byte(in.Password))
			return nil, apperror.Unauthorized("invalid credentials")
		}
		return nil, apperror.Internal()
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(in.Password)); err != nil {
		return nil, apperror.Unauthorized("invalid credentials")
	}
	return s.issuePair(ctx, u.ID)
}

func (s *Service) issuePair(ctx context.Context, userID string) (*TokenPair, error) {
	now := time.Now()
	access, err := MintAccessToken(s.secret, userID, s.accessTTL)
	if err != nil {
		return nil, apperror.Internal()
	}
	raw, hash, err := NewRefreshToken()
	if err != nil {
		return nil, apperror.Internal()
	}
	exp := now.Add(s.refreshTTL)
	if _, err := s.db.RefreshToken.Create().
		SetTokenHash(hash).
		SetExpiresAt(exp).
		SetUserID(userID).
		Save(ctx); err != nil {
		return nil, apperror.Internal()
	}
	return &TokenPair{AccessToken: access, AccessExp: now.Add(s.accessTTL), RefreshRaw: raw, RefreshExp: exp}, nil
}

func (s *Service) Refresh(ctx context.Context, raw string) (*TokenPair, error) {
	now := time.Now()

	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, apperror.Internal()
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	rt, err := tx.RefreshToken.Query().
		Where(refreshtoken.TokenHashEQ(HashRefreshToken(raw))).
		WithUser().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperror.Unauthorized("invalid refresh token")
		}
		return nil, apperror.Internal()
	}
	uid := rt.Edges.User.ID

	revokeAll := func() error {
		_, err := tx.RefreshToken.Update().
			Where(refreshtoken.HasUserWith(user.IDEQ(uid))).
			SetRevokedAt(now).
			Save(ctx)
		return err
	}

	if rt.RevokedAt != nil {
		// Reuse means the token was stolen: lock out everything,
		// even though the answer stays a plain 401.
		if err := revokeAll(); err != nil {
			return nil, apperror.Internal()
		}
		if err := tx.Commit(); err != nil {
			return nil, apperror.Internal()
		}
		committed = true
		return nil, apperror.Unauthorized("invalid refresh token")
	}
	if !rt.ExpiresAt.After(now) {
		if _, err := tx.RefreshToken.UpdateOne(rt).SetRevokedAt(now).Save(ctx); err != nil {
			return nil, apperror.Internal()
		}
		if err := tx.Commit(); err != nil {
			return nil, apperror.Internal()
		}
		committed = true
		return nil, apperror.Unauthorized("invalid refresh token")
	}

	raw2, hash2, err := NewRefreshToken()
	if err != nil {
		return nil, apperror.Internal()
	}
	exp2 := now.Add(s.refreshTTL)
	next, err := tx.RefreshToken.Create().
		SetTokenHash(hash2).
		SetExpiresAt(exp2).
		SetUserID(uid).
		Save(ctx)
	if err != nil {
		return nil, apperror.Internal()
	}
	// Conditional revoke: a concurrent refresh that committed first leaves
	// zero rows, and our uncommitted successor rolls back with the tx.
	n, err := tx.RefreshToken.Update().
		Where(refreshtoken.IDEQ(rt.ID), refreshtoken.RevokedAtIsNil()).
		SetRevokedAt(now).
		SetReplacedBy(next.ID).
		Save(ctx)
	if err != nil {
		return nil, apperror.Internal()
	}
	if n != 1 {
		return nil, apperror.Unauthorized("invalid refresh token")
	}
	access, err := MintAccessToken(s.secret, uid, s.accessTTL)
	if err != nil {
		return nil, apperror.Internal()
	}
	if err := tx.Commit(); err != nil {
		return nil, apperror.Internal()
	}
	committed = true
	return &TokenPair{AccessToken: access, AccessExp: now.Add(s.accessTTL), RefreshRaw: raw2, RefreshExp: exp2}, nil
}

func (s *Service) Logout(ctx context.Context, raw string) error {
	if raw == "" {
		return nil
	}
	if _, err := s.db.RefreshToken.Update().
		Where(refreshtoken.TokenHashEQ(HashRefreshToken(raw))).
		SetRevokedAt(time.Now()).
		Save(ctx); err != nil {
		return apperror.Internal()
	}
	return nil
}

func (s *Service) LogoutAll(ctx context.Context, userID string) error {
	if _, err := s.db.RefreshToken.Update().
		Where(refreshtoken.HasUserWith(user.IDEQ(userID))).
		SetRevokedAt(time.Now()).
		Save(ctx); err != nil {
		return apperror.Internal()
	}
	return nil
}

func (s *Service) GetUser(ctx context.Context, id string) (*ent.User, error) {
	u, err := s.db.User.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperror.New(http.StatusNotFound, "user not found", "not_found")
		}
		return nil, apperror.Internal()
	}
	return u, nil
}

// CheckUsername reports whether rawUsername is free. A caller keeping their
// own name counts as available.
func (s *Service) CheckUsername(ctx context.Context, currentUserID, rawUsername string) (bool, string, error) {
	username := strings.ToLower(strings.TrimSpace(rawUsername))

	if len(username) < 3 || len(username) > 30 {
		return false, "invalid_length", nil
	}

	exists, err := s.db.User.Query().
		Where(
			user.UsernameEQ(username),
			user.IDNEQ(currentUserID),
		).
		Exist(ctx)
	if err != nil {
		return false, "", apperror.Internal()
	}

	if exists {
		return false, "already_taken", nil
	}
	return true, "", nil
}

func (s *Service) SetupUsername(ctx context.Context, userID, rawUsername string) (*ent.User, error) {
	username := strings.ToLower(strings.TrimSpace(rawUsername))

	u, err := s.db.User.UpdateOneID(userID).
		SetUsername(username).
		Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, apperror.Conflict("username is already taken")
		}
		return nil, apperror.Internal()
	}
	return u, nil
}

func (s *Service) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	u, err := s.db.User.Get(ctx, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return apperror.NotFound("user not found")
		}
		return apperror.Internal()
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(currentPassword)); err != nil {
		return apperror.New(http.StatusBadRequest, "invalid_password", "current password does not match")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), BcryptCost)
	if err != nil {
		return apperror.Internal()
	}

	if err := s.db.User.UpdateOneID(userID).SetPassword(string(hash)).Exec(ctx); err != nil {
		return apperror.Internal()
	}

	// A new password ends every session, including ones from before the change.
	if err := s.LogoutAll(ctx, userID); err != nil {
		return apperror.Internal()
	}
	return nil
}
