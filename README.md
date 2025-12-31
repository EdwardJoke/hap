# Hap CLI 2FA Tool

A modern, interactive Command-Line Interface (CLI) for managing Two-Factor Authentication (2FA) accounts, built with Bubble Tea.

## Features

### 🔐 Core Functionality
- **Multiple Account Support**: Manage all your 2FA accounts in one place
- **TOTP Generation**: Secure Time-Based One-Time Passwords
- **QR Code Display**: Easy setup with authenticator apps via scannable QR codes
- **Real-time Countdown**: 30-second timers for all TOTP codes
- **Code Verification**: Validate 6-digit codes from authenticator apps

### 🎨 User Experience
- **Colorful, Modern UI**: Built with Bubble Tea and Lip Gloss for a professional look
- **Interactive Navigation**: Keyboard shortcuts for efficient use
- **Account List View**: See all accounts with their current codes at a glance
- **Step-by-Step Setup**: Guided process for adding new accounts

### ⌨️ Keyboard Shortcuts
- `a`: Add new account
- `k` / `↑`: Move up in account list
- `j` / `↓`: Move down in account list
- `Enter`: Select option or verify code
- `q` / `Esc`: Quit or go back

## Installation

### Prerequisites
- Go 1.25.5 or later

### Setup

1. **Clone or download the repository**
   ```bash
   cd /Users/edwardxie/Documents/hap
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

## Usage

### Running the Application

```bash
# Run directly
go run main.go

# Build and run
make build
./hap
```

### Adding a New Account

1. Press `a` to add a new account
2. Enter the **Account Name** (e.g., `user@example.com`)
3. Enter the **Issuer** (e.g., `Google`, `GitHub`, `Discord`)
4. Scan the displayed QR code with your authenticator app (Google Authenticator, Authy, etc.)
5. Press `Enter` to verify the code

### Managing Existing Accounts

- **Navigate**: Use arrow keys (`↑`/`↓`) or `j`/`k` to move between accounts
- **View Codes**: All accounts display their current TOTP codes in real-time
- **Verify Codes**: Press `Enter` on a selected account to verify its code

### Verifying Codes

1. Select an account from the list
2. Enter the 6-digit code from your authenticator app
3. Press `Enter` to validate
4. The app will show a success or error message

### Exiting the Application

- Press `q` or `Esc` to quit the application

## Project Structure

```
/Users/edwardxie/Documents/hap/
  - main.go            # Main application code
  - .gitignore         # Git ignore patterns
  - .goreleaser.yaml   # GoReleaser configuration
  - LICENSE            # MIT License
  - README.md          # This file
  - go.mod             # Go module definition
  - go.sum             # Go module checksums
```

## Dependencies

| Package | Purpose |
|---------|---------|
| [github.com/charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) | TUI framework |
| [github.com/charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) | Styling library |
| [github.com/pquerna/otp](https://github.com/pquerna/otp) | TOTP generation and validation |
| [github.com/pquerna/otp/totp](https://github.com/pquerna/otp) | TOTP-specific functionality |

## Build and Development

### Build for Production

```bash
go build -o hap
```

### Run Tests

```bash
go test ./...
```

### Lint and Format

```bash
# Lint
staticcheck ./...

# Format
go fmt ./...
```

## Bug Fixes

- Fixed issue where 'a' and 'k' keys couldn't be typed in edit mode
- Resolved "Issuer must be set" error during TOTP generation
- Improved error handling for TOTP key generation
- Fixed account selection and navigation issues

## License

Apache2.0 License - see [LICENSE](LICENSE) for details

## Thanks for:
- [Charm](https://charm.sh/): Bubbletea and Lip Gloss.
- [Pquerna](https://github.com/pquerna): OTP library.
- [Goreleaser](https://goreleaser.com/): Build Go application in cross-platform automatically.