package social

import "ArifulProtik/TownHall/ent"

type StatusResponse struct {
	IsFollowing    bool `json:"is_following"`
	IsFollowedBy   bool `json:"is_followed_by"`
	IsFriend       bool `json:"is_friend"`
	IsSelf         bool `json:"is_self"`
	FollowersCount int  `json:"followers_count"`
	FollowingCount int  `json:"following_count"`
}

type ListUser struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Username *string `json:"username,omitempty"`
}

type ListResponse struct {
	Users      []ListUser `json:"users"`
	NextCursor string     `json:"next_cursor,omitempty"`
	HasMore    bool       `json:"has_more"`
}

func toListUser(u *ent.User) ListUser {
	var username *string
	if u.Username != "" {
		username = &u.Username
	}
	return ListUser{ID: u.ID, Name: u.Name, Username: username}
}
