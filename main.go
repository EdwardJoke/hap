package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// Style definitions
var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFD700")).
			Background(lipgloss.Color("#2C3E50")).
			Padding(0, 2).
			Width(40).
			Align(lipgloss.Center)

	contentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ECF0F1")).
			Padding(1, 2).
			Width(40).
			Align(lipgloss.Center)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3498DB")).
			Padding(0, 2).
			Width(40).
			Align(lipgloss.Center)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#2ECC71")).
			Bold(true).
			Padding(0, 2).
			Width(40).
			Align(lipgloss.Center)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E74C3C")).
			Bold(true).
			Padding(0, 2).
			Width(40).
			Align(lipgloss.Center)

	countdownStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F39C12")).
			Bold(true).
			Width(40).
			Align(lipgloss.Center)

	codeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9B59B6")).
			Bold(true).
			Width(40).
			Align(lipgloss.Center)

	instructionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#1ABC9C")).
				Padding(1, 2).
				Width(40).
				Align(lipgloss.Center)

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#3498DB")).
			Padding(1, 2).
			Width(44)
)

// Account represents a 2FA account
type Account struct {
	Name        string
	Issuer      string
	Secret      string
	CurrentCode string
}

// Model represents the application state
type model struct {
	accounts         []Account
	selectedIdx      int
	qrCode           string
	newAccountName   string
	newAccountIssuer string
	inputCode        string
	validCode        string
	status           string
	countdown        int
	step             int // 0: account list, 1: add account name, 2: add account issuer, 3: show QR, 4: verify code
}

// Msg types
type tickMsg time.Time
type verifyMsg bool

// Init initializes the application
func (m model) Init() tea.Cmd {
	// Initialize with empty accounts slice if needed
	if m.accounts == nil {
		m.accounts = []Account{}
	}
	return tea.Batch(tickCmd(), tea.EnterAltScreen)
}

// Update handles messages and updates the model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			if m.step > 0 {
				// Go back to previous step or account list
				m.step = 0
				m.status = ""
				return m, nil
			}
			return m, tea.Quit
		case "enter":
			return handleEnter(m)
		case "up":
			if m.step == 0 && len(m.accounts) > 0 {
				if m.selectedIdx > 0 {
					m.selectedIdx--
				}
			}
		case "down", "j":
			if m.step == 0 && len(m.accounts) > 0 {
				if m.selectedIdx < len(m.accounts)-1 {
					m.selectedIdx++
				}
			}
		case "a", "A":
			if m.step == 0 {
				// Start adding new account
				m.step = 1
				m.newAccountName = ""
				m.newAccountIssuer = ""
				m.status = "Enter account name (e.g., user@example.com)"
			} else {
				// Allow 'a' to be typed in edit mode
				handleTyping(&m, msg.String())
			}
		case "k":
			if m.step == 0 {
				// Use 'k' for moving up only in account list
				if len(m.accounts) > 0 && m.selectedIdx > 0 {
					m.selectedIdx--
				}
			} else {
				// Allow 'k' to be typed in edit mode
				handleTyping(&m, msg.String())
			}
		case "backspace":
			handleBackspace(&m)
		default:
			handleTyping(&m, msg.String())
		}

	case tickMsg:
		// Update countdown and all account codes
		m.countdown = 30 - (int(time.Now().Unix()) % 30)
		for i := range m.accounts {
			code, err := totp.GenerateCode(m.accounts[i].Secret, time.Now())
			if err == nil {
				m.accounts[i].CurrentCode = code
			}
		}
		if m.step == 3 || m.step == 4 {
			code, err := totp.GenerateCode(m.accounts[m.selectedIdx].Secret, time.Now())
			if err == nil {
				m.validCode = code
			}
		}
		return m, tickCmd()

	case verifyMsg:
		if bool(msg) {
			m.status = "✅ Code verified successfully!"
		} else {
			m.status = "❌ Invalid code, please try again"
		}
		m.inputCode = ""
	}

	return m, nil
}

// handleEnter handles enter key presses based on current step
func handleEnter(m model) (tea.Model, tea.Cmd) {
	switch m.step {
	case 0:
		// Select account from list
		if len(m.accounts) > 0 {
			m.step = 4
			m.status = "Enter the 6-digit code from your authenticator app"
			// Update valid code for selected account
			code, err := totp.GenerateCode(m.accounts[m.selectedIdx].Secret, time.Now())
			if err == nil {
				m.validCode = code
			}
		}
	case 1:
		// Move to entering issuer after name
		if m.newAccountName != "" {
			m.step = 2
			m.status = "Enter issuer name (e.g., Google, GitHub)"
		}
	case 2:
		// Generate secret and QR code in one step
		if m.newAccountIssuer != "" {
			// Generate TOTP key with all required parameters
			key, err := totp.Generate(totp.GenerateOpts{
				Issuer:      m.newAccountIssuer,
				AccountName: m.newAccountName,
				SecretSize:  16,
				Algorithm:   otp.AlgorithmSHA1,
				Digits:      otp.DigitsSix,
				Period:      30,
			})
			if err != nil {
				m.status = "Error generating secret: " + err.Error()
				return m, nil
			}
			// Create new account
			newAcc := Account{
				Name:   m.newAccountName,
				Issuer: m.newAccountIssuer,
				Secret: key.Secret(),
			}
			// Generate QR code URL directly from the key
			m.qrCode = key.URL()
			// Add to accounts list
			m.accounts = append(m.accounts, newAcc)
			m.selectedIdx = len(m.accounts) - 1
			m.step = 3
			m.status = "Scan the QR code with your authenticator app"
			// Generate initial code
			code, _ := totp.GenerateCode(key.Secret(), time.Now())
			m.validCode = code
			m.accounts[m.selectedIdx].CurrentCode = code
		}
	case 3:
		// Move to verification step after QR
		m.step = 4
		m.status = "Enter the 6-digit code from your authenticator app"
	case 4:
		// Verify the code for selected account
		return m, verifyCmd(m.accounts[m.selectedIdx].Secret, m.inputCode)
	}
	return m, nil
}

// handleBackspace handles backspace key presses
func handleBackspace(m *model) {
	switch m.step {
	case 1:
		if len(m.newAccountName) > 0 {
			m.newAccountName = m.newAccountName[:len(m.newAccountName)-1]
		}
	case 2:
		if len(m.newAccountIssuer) > 0 {
			m.newAccountIssuer = m.newAccountIssuer[:len(m.newAccountIssuer)-1]
		}
	case 4:
		if len(m.inputCode) > 0 {
			m.inputCode = m.inputCode[:len(m.inputCode)-1]
		}
	}
}

// handleTyping handles text input based on current step
func handleTyping(m *model, key string) {
	if len(key) != 1 {
		return
	}
	switch m.step {
	case 1:
		m.newAccountName += key
	case 2:
		m.newAccountIssuer += key
	case 4:
		if len(m.inputCode) < 6 {
			m.inputCode += key
		}
	}
}

// View renders the UI
func (m model) View() string {
	header := headerStyle.Render("CLI 2FA Tool")

	var content string
	switch m.step {
	case 0:
		// Account list view
		if len(m.accounts) == 0 {
			content = instructionStyle.Render("No accounts yet. Press 'a' to add a new account.")
		} else {
			accountList := ""
			for i, acc := range m.accounts {
				prefix := "  "
				if i == m.selectedIdx {
					prefix = "▶ "
					accountList += codeStyle.Render(fmt.Sprintf("%s%s (%s): %s", prefix, acc.Name, acc.Issuer, acc.CurrentCode))
				} else {
					accountList += contentStyle.Render(fmt.Sprintf("%s%s (%s): %s", prefix, acc.Name, acc.Issuer, acc.CurrentCode))
				}
				accountList += "\n"
			}
			content = fmt.Sprintf("%s\n\n%s", accountList, instructionStyle.Render("Press 'a' to add account, arrow keys to navigate, Enter to verify"))
		}
	case 1:
		// Add account name
		content = fmt.Sprintf("%s\n\n%s: %s",
			statusStyle.Render(m.status),
			contentStyle.Render("Account Name"),
			codeStyle.Render(m.newAccountName))
	case 2:
		// Add account issuer
		content = fmt.Sprintf("%s\n\n%s: %s",
			statusStyle.Render(m.status),
			contentStyle.Render("Issuer"),
			codeStyle.Render(m.newAccountIssuer))
	case 3:
		// Show QR code
		qrLine := contentStyle.Render(m.qrCode)
		codeLine := codeStyle.Render(m.validCode)
		instruction := instructionStyle.Render("Press Enter to verify code")
		content = fmt.Sprintf("%s\n\n%s\n\n%s", qrLine, codeLine, instruction)
	case 4:
		// Verify code
		countdownLine := countdownStyle.Render(fmt.Sprintf("Countdown: %ds", m.countdown))
		inputLine := contentStyle.Render(fmt.Sprintf("Enter code: %s", m.inputCode))
		validLine := codeStyle.Render(m.validCode)
		content = fmt.Sprintf("%s\n\n%s\n\n%s", countdownLine, inputLine, validLine)
	}

	// Render status with appropriate style
	var status string
	if m.step == 0 || m.step == 1 || m.step == 2 {
		// Don't render status again if it's already in the content
		status = ""
	} else {
		if strings.Contains(m.status, "✅") {
			status = successStyle.Render(m.status)
		} else if strings.Contains(m.status, "❌") {
			status = errorStyle.Render(m.status)
		} else {
			status = statusStyle.Render(m.status)
		}
	}

	quitText := statusStyle.Render("Press q/esc to quit")

	// Combine all parts with border
	combined := fmt.Sprintf(
		"%s\n\n%s",
		header,
		content,
	)

	if status != "" {
		combined += fmt.Sprintf("\n\n%s", status)
	}

	combined += fmt.Sprintf("\n\n%s", quitText)

	return borderStyle.Render(combined)
}

// tickCmd sends a tick message every second
func tickCmd() tea.Cmd {
	return tea.Every(1*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// verifyCmd verifies the TOTP code
func verifyCmd(secret, code string) tea.Cmd {
	return func() tea.Msg {
		valid := totp.Validate(code, secret)
		return verifyMsg(valid)
	}
}

// generateTOTPSecret generates a new TOTP secret
func generateTOTPSecret() (string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		SecretSize: 16,
		Algorithm:  otp.AlgorithmSHA1,
		Digits:     otp.DigitsSix,
		Period:     30,
	})
	if err != nil {
		return "", err
	}
	return key.Secret(), nil
}

// generateQRCodeForAccount generates a QR code URL for an account
func generateQRCodeForAccount(acc Account) (string, error) {
	// Generate key from account details
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      acc.Issuer,
		AccountName: acc.Name,
		SecretSize:  16,
		Secret:      []byte(acc.Secret),
		Algorithm:   otp.AlgorithmSHA1,
		Digits:      otp.DigitsSix,
		Period:      30,
	})
	if err != nil {
		return "", err
	}
	return key.URL(), nil
}

// centerText centers text within a given width
func centerText(s string, width int) string {
	lines := []string{}
	for _, line := range strings.Split(s, "\n") {
		padding := (width - len(line)) / 2
		if padding > 0 {
			line = strings.Repeat(" ", padding) + line + strings.Repeat(" ", padding)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func main() {
	p := tea.NewProgram(model{})
	if err := p.Start(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
