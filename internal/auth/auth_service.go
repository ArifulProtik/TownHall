package auth

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"ArifulProtik/TownHall/ent"
	"ArifulProtik/TownHall/ent/refreshtoken"
	"ArifulProtik/TownHall/ent/user"
	"ArifulProtik/TownHall/internal/config"
	"ArifulProtik/TownHall/pkg/apperror"
	"ArifulProtik/TownHall/pkg/logger"

	"golang.org/x/crypto/bcrypt"
)

// Service implements auth use cases over the Ent client.
type Service struct {
	config  *config.Config
	db      *ent.Client
	log     *slog.Logger
	limiter Limiter
}

// NewService builds an Service around cfg and db.
func NewService(config *config.Config, db *ent.Client, log *slog.Logger) *Service {
	return &Service{config: config, db: db, log: log, limiter: NewMemoryLimiter()}
}

// GetAppEnv returns the configured application environment.
func (s *Service) GetAppEnv() string {
	return s.config.AppEnv
}

// SignupEmail hashes the password and persists a new email user.
// A duplicate email returns a 409 AppError.
func (s *Service) SignupEmail(ctx context.Context, in SignupEmail) (*ent.User, error) {
	log := logger.WithContext(ctx, s.log)
	name := strings.TrimSpace(in.Name)
	email := strings.ToLower(strings.TrimSpace(in.Email))

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), BcryptCost)
	if err != nil {
		log.Error("signup: password hash failed", slog.Any("error", err))
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
			log.Warn("signup: duplicate email", slog.String("email", email))
			return nil, apperror.Conflict("email already registered")
		}
		log.Error("signup: user create failed", slog.Any("error", err))
		return nil, apperror.Internal()
	}
	return u, nil
}

// AllowAttempt reports whether an auth attempt for email+ip passes rate limits.
func (s *Service) AllowAttempt(email, ip string) (bool, time.Duration) {
	return s.limiter.Allow(strings.ToLower(strings.TrimSpace(email)), ip)
}

// Login verifies credentials and issues an access + refresh token pair.
func (s *Service) Login(ctx context.Context, in LoginRequest) (*TokenPair, error) {
	log := logger.WithContext(ctx, s.log)
	email := strings.ToLower(strings.TrimSpace(in.Email))

	u, err := s.db.User.Query().Where(user.EmailEQ(email)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			// Timing guard: match the cost of a real compare before failing.
			_ = bcrypt.CompareHashAndPassword([]byte(dummyPasswordHash), []byte(in.Password))
			log.Warn("login: unknown email", slog.String("email", email))
			return nil, apperror.Unauthorized("invalid credentials")
		}
		log.Error("login: user lookup failed", slog.Any("error", err))
		return nil, apperror.Internal()
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(in.Password)); err != nil {
		log.Warn("login: bad password", slog.String("email", email))
		return nil, apperror.Unauthorized("invalid credentials")
	}
	return s.issuePair(ctx, u.ID)
}

// issuePair mints an access token and persists a new refresh token row.
func (s *Service) issuePair(ctx context.Context, userID string) (*TokenPair, error) {
	log := logger.WithContext(ctx, s.log)
	now := time.Now()
	access, err := MintAccessToken(s.config.JWTSecret, userID, s.config.JWTAccessTTL)
	if err != nil {
		log.Error("login: access token mint failed", slog.Any("error", err))
		return nil, apperror.Internal()
	}
	raw, hash, err := NewRefreshToken()
	if err != nil {
		log.Error("login: refresh token generation failed", slog.Any("error", err))
		return nil, apperror.Internal()
	}
	exp := now.Add(s.config.JWTRefreshTTL)
	if _, err := s.db.RefreshToken.Create().
		SetTokenHash(hash).
		SetExpiresAt(exp).
		SetUserID(userID).
		Save(ctx); err != nil {
		log.Error("login: refresh token store failed", slog.Any("error", err))
		return nil, apperror.Internal()
	}
	return &TokenPair{AccessToken: access, AccessExp: now.Add(s.config.JWTAccessTTL), RefreshRaw: raw, RefreshExp: exp}, nil
}

// Refresh rotates a refresh token inside a single transaction.
func (s *Service) Refresh(ctx context.Context, raw string) (*TokenPair, error) {
	log := logger.WithContext(ctx, s.log)
	now := time.Now()

	tx, err := s.db.Tx(ctx)
	if err != nil {
		log.Error("refresh: begin tx failed", slog.Any("error", err))
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
			log.Warn("refresh: unknown token")
			return nil, apperror.Unauthorized("invalid refresh token")
		}
		log.Error("refresh: token lookup failed", slog.Any("error", err))
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
		// Reuse detected: assume compromise, revoke everything.
		if err := revokeAll(); err != nil {
			log.Error("refresh: revoke-all failed", slog.Any("error", err))
			return nil, apperror.Internal()
		}
		if err := tx.Commit(); err != nil {
			log.Error("refresh: commit failed", slog.Any("error", err))
			return nil, apperror.Internal()
		}
		committed = true
		log.Warn("refresh: reuse detected, revoked all user tokens")
		return nil, apperror.Unauthorized("invalid refresh token")
	}
	if !rt.ExpiresAt.After(now) {
		if _, err := tx.RefreshToken.UpdateOne(rt).SetRevokedAt(now).Save(ctx); err != nil {
			log.Error("refresh: revoke expired failed", slog.Any("error", err))
			return nil, apperror.Internal()
		}
		if err := tx.Commit(); err != nil {
			log.Error("refresh: commit failed", slog.Any("error", err))
			return nil, apperror.Internal()
		}
		committed = true
		log.Warn("refresh: expired token")
		return nil, apperror.Unauthorized("invalid refresh token")
	}

	raw2, hash2, err := NewRefreshToken()
	if err != nil {
		log.Error("refresh: token generation failed", slog.Any("error", err))
		return nil, apperror.Internal()
	}
	exp2 := now.Add(s.config.JWTRefreshTTL)
	next, err := tx.RefreshToken.Create().
		SetTokenHash(hash2).
		SetExpiresAt(exp2).
		SetUserID(uid).
		Save(ctx)
	if err != nil {
		log.Error("refresh: successor store failed", slog.Any("error", err))
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
		log.Error("refresh: revoke current failed", slog.Any("error", err))
		return nil, apperror.Internal()
	}
	if n != 1 {
		log.Warn("refresh: concurrent rotation lost")
		return nil, apperror.Unauthorized("invalid refresh token")
	}
	access, err := MintAccessToken(s.config.JWTSecret, uid, s.config.JWTAccessTTL)
	if err != nil {
		log.Error("refresh: access token mint failed", slog.Any("error", err))
		return nil, apperror.Internal()
	}
	if err := tx.Commit(); err != nil {
		log.Error("refresh: commit failed", slog.Any("error", err))
		return nil, apperror.Internal()
	}
	committed = true
	return &TokenPair{AccessToken: access, AccessExp: now.Add(s.config.JWTAccessTTL), RefreshRaw: raw2, RefreshExp: exp2}, nil
}

// Logout revokes a single refresh token; unknown tokens are still success.
func (s *Service) Logout(ctx context.Context, raw string) error {
	log := logger.WithContext(ctx, s.log)
	if raw == "" {
		return nil
	}
	if _, err := s.db.RefreshToken.Update().
		Where(refreshtoken.TokenHashEQ(HashRefreshToken(raw))).
		SetRevokedAt(time.Now()).
		Save(ctx); err != nil {
		log.Error("logout: revoke failed", slog.Any("error", err))
		return apperror.Internal()
	}
	return nil
}

// LogoutAll revokes every refresh token for userID.
func (s *Service) LogoutAll(ctx context.Context, userID string) error {
	log := logger.WithContext(ctx, s.log)
	if _, err := s.db.RefreshToken.Update().
		Where(refreshtoken.HasUserWith(user.IDEQ(userID))).
		SetRevokedAt(time.Now()).
		Save(ctx); err != nil {
		log.Error("logout-all: revoke failed", slog.Any("error", err))
		return apperror.Internal()
	}
	return nil
}

// GetUser retrieves an existing user by ID.
func (s *Service) GetUser(ctx context.Context, id string) (*ent.User, error) {
	u, err := s.db.User.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperror.New(http.StatusNotFound, "user not found", "not_found")
		}
		return nil, err
	}
	return u, nil
}

// CheckUsername checks whether the given username is available.
// If the caller already owns the username, it is considered available.
func (s *Service) CheckUsername(ctx context.Context, currentUserID, rawUsername string) (bool, string, error) {
	log := logger.WithContext(ctx, s.log)
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
		log.Error("check username: query failed", slog.Any("error", err))
		return false, "", apperror.Internal()
	}

	if exists {
		return false, "already_taken", nil
	}
	return true, "", nil
}

// SetupUsername sets or updates the authenticated user's username.
func (s *Service) SetupUsername(ctx context.Context, userID, rawUsername string) (*ent.User, error) {
	log := logger.WithContext(ctx, s.log)
	username := strings.ToLower(strings.TrimSpace(rawUsername))

	u, err := s.db.User.UpdateOneID(userID).
		SetUsername(username).
		Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			log.Warn("setup username: duplicate username", slog.String("username", username))
			return nil, apperror.Conflict("username is already taken")
		}
		log.Error("setup username: update failed", slog.Any("error", err))
		return nil, apperror.Internal()
	}
	return u, nil
}

// ChangePassword verifies the current password and updates it to the new password.
func (s *Service) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	log := logger.WithContext(ctx, s.log)
	u, err := s.db.User.Get(ctx, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return apperror.NotFound("user not found")
		}
		log.Error("change password: get user failed", slog.Any("error", err))
		return apperror.Internal()
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(currentPassword)); err != nil {
		log.Warn("change password: invalid current password", slog.String("user_id", userID))
		return apperror.New(http.StatusBadRequest, "invalid_password", "current password does not match")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), BcryptCost)
	if err != nil {
		log.Error("change password: hash failed", slog.Any("error", err))
		return apperror.Internal()
	}

	if err := s.db.User.UpdateOneID(userID).SetPassword(string(hash)).Exec(ctx); err != nil {
		log.Error("change password: update failed", slog.Any("error", err))
		return apperror.Internal()
	}

	// A password change must invalidate existing sessions: revoke every
	// refresh token, including tokens minted before the change.
	if err := s.LogoutAll(ctx, userID); err != nil {
		log.Error("change password: revoke sessions failed", slog.Any("error", err))
		return apperror.Internal()
	}
	return nil
}
