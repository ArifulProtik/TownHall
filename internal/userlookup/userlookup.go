// Package userlookup resolves a handle (username, id, or "me") to a user.
// It exists so social, profile, and notification share one implementation
// instead of triplicating resolveTarget.
package userlookup

import (
	"context"
	"strings"

	"ArifulProtik/TownHall/ent"
	"ArifulProtik/TownHall/ent/user"
	"ArifulProtik/TownHall/pkg/apperror"
)

func Resolve(ctx context.Context, db *ent.Client, viewerID, target string) (*ent.User, error) {
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
	u, err := db.User.Query().Where(user.UsernameEQ(strings.ToLower(clean))).Only(ctx)
	if err == nil {
		return u, nil
	}
	if !ent.IsNotFound(err) {
		return nil, apperror.Internal()
	}
	u, err = db.User.Get(ctx, clean)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperror.NotFound("user not found")
		}
		return nil, apperror.Internal()
	}
	return u, nil
}
