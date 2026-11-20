package xiaohongshu

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-rod/rod"
	"github.com/pkg/errors"
	"github.com/xpzouying/xiaohongshu-mcp/configs"
)

const (
	loginStatusTimeout = 30 * time.Second
	loginStatusDelay   = time.Second
	loginPageDelay     = 2 * time.Second
	loginPollInterval  = 500 * time.Millisecond
	loginSelector      = ".main-container .user .link-wrapper .channel"
)

// loginPageURL 海外扫码使用独立登录页。
func loginPageURL() string {
	if configs.IsIntl() {
		return configs.WebURL("/login")
	}
	return configs.WebURL("/explore")
}

// isLoggedIn 兼容两站页面状态和旧侧栏标记。
func isLoggedIn(page *rod.Page) (bool, error) {
	result, err := page.Eval(`(selector) => {
		const unwrap = v => v?.value !== undefined ? v.value : (v?._value !== undefined ? v._value : v);
		const user = window.__INITIAL_STATE__?.user;
		const info = unwrap(user?.userInfo) || unwrap(user?.currentUser) || unwrap(user?.me);
		return unwrap(user?.loggedIn) === true ||
			!!(info && !info.guest && (info.userId || info.user_id || info.nickname)) ||
			document.querySelector("#global")?.getAttribute("data-logged") === "1" ||
			!!document.querySelector(selector);
	}`, loginSelector)
	if err != nil {
		return false, errors.Wrap(err, "read login state failed")
	}
	return result.Value.Bool(), nil
}

type LoginAction struct {
	page *rod.Page
}

func NewLogin(page *rod.Page) *LoginAction {
	return &LoginAction{page: page}
}

func (a *LoginAction) CheckLoginStatus(ctx context.Context) (bool, error) {
	// 加超时保护：只是查登录态的快速检查，不应无限挂（登录扫码的等待在 Login/WaitForLogin 里）
	pp := a.page.Context(ctx).Timeout(loginStatusTimeout)
	pp.MustNavigate(configs.WebURL("/explore")).MustWaitLoad()

	time.Sleep(loginStatusDelay)
	return isLoggedIn(pp)
}

// CurrentUser 当前登录用户的基础信息。
type CurrentUser struct {
	Nickname string `json:"nickname"`
	UserID   string `json:"userId"`
}

// CurrentUser 从当前页面的 __INITIAL_STATE__ 读取登录用户信息。
// 需在 CheckLoginStatus 之后调用：复用已加载的 explore 页，不做额外导航。
func (a *LoginAction) CurrentUser(ctx context.Context) (*CurrentUser, error) {
	pp := a.page.Context(ctx).Timeout(10 * time.Second)

	res, err := pp.Eval(`() => {
		const unwrap = v => v?.value !== undefined ? v.value : (v?._value !== undefined ? v._value : v);
		const user = window.__INITIAL_STATE__?.user;
		const info = unwrap(user?.userInfo) || unwrap(user?.currentUser) || unwrap(user?.me);
		if (!info || info.guest) return "";
		return JSON.stringify({nickname: info.nickname, userId: info.userId || info.user_id});
	}`)
	if err != nil {
		return nil, errors.Wrap(err, "read current user state failed")
	}

	raw := res.Value.String()
	if raw == "" {
		return nil, errors.New("current user not found in page state")
	}

	var user CurrentUser
	if err := json.Unmarshal([]byte(raw), &user); err != nil {
		return nil, errors.Wrap(err, "unmarshal current user failed")
	}

	return &user, nil
}

func (a *LoginAction) Login(ctx context.Context) error {
	pp := a.page.Context(ctx)

	pp.MustNavigate(loginPageURL()).MustWaitLoad()

	time.Sleep(loginPageDelay)

	if loggedIn, err := isLoggedIn(pp); err != nil || loggedIn {
		return err
	}

	if a.WaitForLogin(ctx) {
		return nil
	}
	return ctx.Err()
}

func (a *LoginAction) FetchQrcodeImage(ctx context.Context) (string, bool, error) {
	pp := a.page.Context(ctx)

	pp.MustNavigate(loginPageURL()).MustWaitLoad()

	time.Sleep(loginPageDelay)

	loggedIn, err := isLoggedIn(pp)
	if err != nil {
		return "", false, err
	}
	if loggedIn {
		return "", true, nil
	}

	src, err := pp.MustElement(".login-container .qrcode-img").Attribute("src")
	if err != nil {
		return "", false, errors.Wrap(err, "get qrcode src failed")
	}
	if src == nil || len(*src) == 0 {
		return "", false, errors.New("qrcode src is empty")
	}

	return *src, false, nil
}

func (a *LoginAction) WaitForLogin(ctx context.Context) bool {
	pp := a.page.Context(ctx)
	ticker := time.NewTicker(loginPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
			if loggedIn, err := isLoggedIn(pp); err == nil && loggedIn {
				return true
			}
		}
	}
}
