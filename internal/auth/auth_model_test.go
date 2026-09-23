package auth

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ArifulProtik/TownHall/ent"
)

func TestToUserResponse_NilUsernameWhenEmpty(t *testing.T) {
	u := &ent.User{
		ID:       "user-1",
		Name:     "Test User",
		Email:    "test@example.com",
		Username: "",
	}
	resp := ToUserResponse(u)
	assert.Nil(t, resp.Username)

	data, err := json.Marshal(resp)
	require.NoError(t, err)
	assert.NotContains(t, string(data), `"username":""`)

	u2 := &ent.User{
		ID:       "user-2",
		Name:     "Test User 2",
		Email:    "test2@example.com",
		Username: "valid_user",
	}
	resp2 := ToUserResponse(u2)
	require.NotNil(t, resp2.Username)
	assert.Equal(t, "valid_user", *resp2.Username)
}
