# Vern Roadmap

## Current Status (v0.3.0)

### ✅ Completed
- Core CLI commands: `install`, `list`, `remove`, `default`, `update`, `setup`
- Smart version resolution (latest, partial, exact)
- Multi-select remove with comma-separated input
- Language version switching via `.vern` files
- Shims system with PATH integration
- Source compilation support (Python compiles from source like pyenv)
- Pre-compiled binary support (Go, Node.js)
- `vern update` command (self-update + language list)
- Installer script (`curl | bash`)
- GitHub Actions for automated releases (4 platforms)
- Versioned language lists (independent from binary version)

### ⚠️ Known Issues
- **Ruby**: Compilation requires host Ruby (`basruby`). Need to either:
  - Install system Ruby first (`apt install ruby`)
  - Use pre-compiled Ruby binaries from another source
  - Bundle mini Ruby for bootstrapping
- **Python shared libraries**: Resolved by compiling from source
- **xz compression**: Requires `xz` command (not built into Go)

## Short Term (v0.4.0)

### Fix Ruby Installation
- [ ] Option 1: Add `--with-basruby=no` to skip Ruby compilation check
- [ ] Option 2: Use pre-compiled Ruby from `ruby-build` or similar
- [ ] Option 3: Install minimal host Ruby automatically

### Improve Version Resolution
- [ ] Cache scraped versions locally (reduce network calls)
- [ ] Add timeout for HTTP requests
- [ ] Handle rate limiting from GitHub API

### Enhance Shims
- [ ] Auto-create shims on `vern install`
- [ ] Add `vern shims` command to regenerate shims
- [ ] Support multiple binaries per language (e.g., `python3`, `pip3` for Python)

### Better Error Handling
- [ ] User-friendly error messages
- [ ] Suggest fixes (e.g., "Run: apt install gcc" when build fails)
- [ ] Check for required tools (`xz`, `gcc`, `make`) before installation

## Medium Term (v0.5.0)

### Plugin System
- [ ] Allow third-party language definitions
- [ ] Support fetching language configs from URLs
- [ ] Language config validation

### Shell Integration
- [ ] Auto-source shims in shell config detection
- [ ] Fish shell support (shims work, but setup could be better)
- [ ] Zsh/Fish completion scripts
- [ ] Prompt integration (show current version in prompt)

### Advanced Version Resolution
- [ ] Semantic versioning support (`~3.11`, `^3.11`)
- [ ] Version ranges (`>=3.11 <4.0`)
- [ ] Local version cache with TTL

## Long Term (v1.0.0)

### Full rbenv/pyenv Compatibility
- [ ] Support `.python-version` and `.ruby-version` files
- [ ] Support `.tool-versions` (asdf format)
- [ ] Migration tool from asdf/pyenv/rbenv

### IDE Integration
- [ ] VS Code extension
- [ ] Language server protocol support
- [ ] Version detection in project files (`package.json`, `requirements.txt`)

### Enterprise Features
- [ ] Offline mode (use cached versions)
- [ ] Private language registries
- [ ] Team settings (share `.vern` configs)
- [ ] Audit logging

## Future Ideas

### Performance
- [ ] Parallel downloads
- [ ] Incremental compilation
- [ ] Binary diff updates

### User Experience
- [ ] Interactive mode (`vern interactive`)
- [ ] Colored output
- [ ] Progress bars for downloads
- [ ] Desktop GUI (Electron/Tauri)

### Extensibility
- [ ] Plugin marketplace
- [ ] Webhook support (notify on install/remove)
- [ ] REST API for remote management

## Contributing

See `AGENTS.md` for development guidelines.

## Release Cadence

- **Patch releases** (0.3.x): Bug fixes, minor improvements
- **Minor releases** (0.x.0): New features, language support
- **Major releases** (x.0.0): Breaking changes, major milestones

Target: **v1.0.0** by end of 2026.
