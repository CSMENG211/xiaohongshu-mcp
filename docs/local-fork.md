# Run our Rednote fork

Use `CSMENG211/xiaohongshu-mcp` as `origin` on branch `main` and `xpzouying/xiaohongshu-mcp` as `upstream`. The Go module path stays unchanged so internal imports continue to work.

This fork includes the overseas-session changes from upstream PR #915 (commit `6f99629a55494f4bb72821235c09aef22123cc04`) and the login behavior recovered from our previous `v2.7.4-rednote` binary:

- Select the main site and creator center from the session cookie domain, with `XHS_SITE` as an override.
- Recognize Rednote search responses and creator-center sessions.
- Use `https://www.rednote.com/login` for Rednote QR login.
- Detect login from page state, including `value` and `_value` refs, and the Rednote login marker. Keep the existing sidebar check as a fallback.
- Read the current user from either site's page state.

The page-state checks need JavaScript because that state is held in the site's JavaScript objects. Navigation, polling, and QR image access use go-rod.

## Build on Linux or macOS

Install Go 1.27.3 or newer. On an Apple Silicon Mac, use `brew install go` in a native ARM64 terminal.

```bash
git clone --branch main git@github.com:CSMENG211/xiaohongshu-mcp.git
cd xiaohongshu-mcp
git remote add upstream https://github.com/xpzouying/xiaohongshu-mcp.git
./scripts/build-local.sh
```

The script builds `bin/xiaohongshu-mcp` and `bin/xiaohongshu-login` for the current machine. Rebuild after pulling source changes. Build output and cookies are ignored by Git.

```bash
COOKIES_PATH="/absolute/path/to/assistant/mcp/cookies.json" XHS_SITE=rednote DBUS_SESSION_BUS_ADDRESS=/dev/null ./bin/xiaohongshu-mcp -port 127.0.0.1:18060
```

The assistant keeps the existing cookie file, fingerprint seed, and MCP URL (`http://localhost:18060/mcp`). `assistant/start.sh` starts this checkout's binary. The assistant's `docs/modules/macos_migration.md` covers Mac startup and LaunchAgents.

For a fresh login, run the assistant's `scripts/rednote_login.py` while MCP is running, then scan the generated QR code. On a desktop you can also run:

```bash
COOKIES_PATH="/absolute/path/to/assistant/mcp/cookies.json" XHS_SITE=rednote ./bin/xiaohongshu-login
```

## Verify

```bash
go test -failfast ./...
go vet ./...
XHS_TEST_BROWSER=/path/to/chrome go test -failfast -tags integration ./xiaohongshu -run '^TestLoginStateLocal$' -count=1
curl -fsS http://127.0.0.1:18060/health
curl -fsS http://127.0.0.1:18060/api/v1/login/status
```

The browser fixture tests use local pages and do not need account cookies. Live verification should check login, MCP tool discovery, search, and reading a note. Publishing, comments, likes, and favorites change the account and require their own user request.
