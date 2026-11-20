//go:build integration

package xiaohongshu

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoginStateLocal(t *testing.T) {
	b := newSearchFixtureBrowser(t)
	for _, tc := range []struct {
		name  string
		state string
		want  bool
	}{
		{"logged_in_value", `window.__INITIAL_STATE__ = {user:{loggedIn:{value:true}}}`, true},
		{"logged_in_private_value", `window.__INITIAL_STATE__ = {user:{loggedIn:{_value:true}}}`, true},
		{"user_info_value", `window.__INITIAL_STATE__ = {user:{userInfo:{value:{userId:"fixture",nickname:"Fixture"}}}}`, true},
		{"user_info_private_value", `window.__INITIAL_STATE__ = {user:{userInfo:{_value:{user_id:"fixture",nickname:"Fixture"}}}}`, true},
		{"current_user", `window.__INITIAL_STATE__ = {user:{currentUser:{userId:"fixture"}}}`, true},
		{"me", `window.__INITIAL_STATE__ = {user:{me:{nickname:"Fixture"}}}`, true},
		{"guest", `window.__INITIAL_STATE__ = {user:{userInfo:{value:{userId:"guest",guest:true}}}}`, false},
		{"no_state", `delete window.__INITIAL_STATE__`, false},
		{"rednote_marker", `document.body.innerHTML = '<div id="global" data-logged="1"></div>'`, true},
		{"sidebar", `document.body.innerHTML = '<div class="main-container"><div class="user"><div class="link-wrapper"><div class="channel"></div></div></div></div>'`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			page := b.MustPage("data:text/html,<html><body></body></html>")
			defer page.Close()
			page.MustEval("() => {" + tc.state + "}")
			loggedIn, err := isLoggedIn(page)
			require.NoError(t, err)
			require.Equal(t, tc.want, loggedIn)
			if tc.name == "user_info_value" || tc.name == "user_info_private_value" {
				user, err := NewLogin(page).CurrentUser(context.Background())
				require.NoError(t, err)
				require.Equal(t, "fixture", user.UserID)
				require.Equal(t, "Fixture", user.Nickname)
			}
		})
	}
}
