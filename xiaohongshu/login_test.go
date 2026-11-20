package xiaohongshu

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoginPageURL(t *testing.T) {
	for _, tc := range []struct {
		site string
		want string
	}{
		{"rednote", "https://www.rednote.com/login"},
		{"xiaohongshu", "https://www.xiaohongshu.com/explore"},
	} {
		t.Run(tc.site, func(t *testing.T) {
			t.Setenv("XHS_SITE", tc.site)
			require.Equal(t, tc.want, loginPageURL())
		})
	}
}
