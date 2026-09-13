package auth

import (
	"context"
	"log/slog"
	"strings"

	"ArifulProtik/TownHall/ent"
	"ArifulProtik/TownHall/ent/user"
	"ArifulProtik/TownHall/internal/config"
	"ArifulProtik/TownHall/pkg/apperror"
	"ArifulProtik/TownHall/pkg/logger"

	"golang.org/x/crypto/bcrypt"
)

// Service implements auth use cases over the Ent client.
type Service struct {
	config *config.Config
	db     *ent.Client
	log    *slog.Logger
}

// NewService builds an Service around cfg and db.
func NewService(config *config.Config, db *ent.Client, log *slog.Logger) *Service {
	return &Service{config: config, db: db, log: log}
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

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
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
