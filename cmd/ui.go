package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/pterm/pterm"
	"github.com/sunil-saini/astat/internal/refresh"
)

// UI provides an abstraction for user interface output
// It supports both interactive (pterm) and non-interactive (plain text) modes
type UI interface {
	Header(text string)
	Info(text string)
	Success(text string)
	Warning(text string)
	Error(text string)
	Spinner(text string) Spinner
	Section(title string)
	BulletList(items []string)
	Println()
	GetRefreshTracker(serviceName string) refresh.Tracker
	IsInteractive() bool
	StartRefresh()
	StopRefresh()
}

// Spinner represents a progress indicator
type Spinner interface {
	Update(text string)
	Success(text string)
	Fail(text string)
}

// NewUI creates a UI instance based on whether we're in an interactive terminal
func NewUI() UI {
	if isInteractive() {
		return &interactiveUI{}
	}
	return &plainUI{}
}

// interactiveUI uses pterm for rich terminal output
type interactiveUI struct {
	multi *pterm.MultiPrinter
}

func (u *interactiveUI) Header(text string) {
	pterm.DefaultHeader.WithFullWidth().Println(text)
}

func (u *interactiveUI) Info(text string) {
	pterm.Info.Println(text)
}

func (u *interactiveUI) Success(text string) {
	pterm.Success.Println(text)
}

func (u *interactiveUI) Warning(text string) {
	pterm.Warning.Println(text)
}

func (u *interactiveUI) Error(text string) {
	pterm.Error.Println(text)
}

func (u *interactiveUI) Spinner(text string) Spinner {
	s, _ := pterm.DefaultSpinner.Start(text)
	return &ptermSpinner{spinner: s}
}

func (u *interactiveUI) Section(title string) {
	pterm.DefaultSection.Println(title)
}

func (u *interactiveUI) BulletList(items []string) {
	bulletItems := make([]pterm.BulletListItem, len(items))
	for i, item := range items {
		bulletItems[i] = pterm.BulletListItem{Text: item, Bullet: "→"}
	}
	pterm.BulletListPrinter{Items: bulletItems}.Render()
}

func (u *interactiveUI) Println() {
	fmt.Println()
}

func (u *interactiveUI) GetRefreshTracker(serviceName string) refresh.Tracker {
	var writer io.Writer
	if u.multi != nil {
		writer = u.multi.NewWriter()
	} else {
		writer = os.Stdout
	}

	s, _ := pterm.DefaultSpinner.
		WithSequence("⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏").
		WithStyle(pterm.NewStyle(pterm.FgLightCyan)).
		WithMessageStyle(pterm.NewStyle(pterm.FgLightCyan)).
		WithWriter(writer).
		WithRemoveWhenDone(false).
		Start(pterm.LightCyan(fmt.Sprintf("%s pending...", serviceName)))

	s.SuccessPrinter = pterm.Success.WithPrefix(pterm.Prefix{Text: " ✓ ", Style: pterm.NewStyle(pterm.FgLightGreen)}).
		WithMessageStyle(pterm.NewStyle(pterm.FgLightGreen))
	s.FailPrinter = pterm.Error.WithPrefix(pterm.Prefix{Text: " ✗ ", Style: pterm.NewStyle(pterm.FgLightRed)}).
		WithMessageStyle(pterm.NewStyle(pterm.FgLightRed))

	return &ptermSpinner{spinner: s}
}

func (u *interactiveUI) IsInteractive() bool {
	return true
}

func (u *interactiveUI) StartRefresh() {
	multi := pterm.DefaultMultiPrinter
	u.multi = &multi
	u.multi.Start()
}

func (u *interactiveUI) StopRefresh() {
	if u.multi != nil {
		u.multi.Stop()
	}
}

// ptermSpinner wraps pterm.SpinnerPrinter
type ptermSpinner struct {
	spinner *pterm.SpinnerPrinter
}

func (s *ptermSpinner) Update(text string) {
	s.spinner.UpdateText(text)
}

func (s *ptermSpinner) Success(text string) {
	s.spinner.Success(text)
}

func (s *ptermSpinner) Fail(text string) {
	s.spinner.Fail(text)
}

func (s *ptermSpinner) Error(text string) {
	s.spinner.Fail(text)
}

// plainUI uses simple fmt output for non-interactive terminals
type plainUI struct{}

func (u *plainUI) Header(text string) {
	fmt.Printf("==> %s\n", text)
}

func (u *plainUI) Info(text string) {
	fmt.Printf("ℹ %s\n", text)
}

func (u *plainUI) Success(text string) {
	fmt.Printf("✓ %s\n", text)
}

func (u *plainUI) Warning(text string) {
	fmt.Printf("⚠ %s\n", text)
}

func (u *plainUI) Error(text string) {
	fmt.Printf("✗ %s\n", text)
}

func (u *plainUI) Spinner(text string) Spinner {
	fmt.Printf("→ %s\n", text)
	return &plainSpinner{}
}

func (u *plainUI) Section(title string) {
	fmt.Printf("\n%s:\n", title)
}

func (u *plainUI) BulletList(items []string) {
	for _, item := range items {
		fmt.Printf("  → %s\n", item)
	}
}

func (u *plainUI) Println() {
	fmt.Println()
}

func (u *plainUI) GetRefreshTracker(serviceName string) refresh.Tracker {
	return newPlainTracker(serviceName)
}

func (u *plainUI) IsInteractive() bool {
	return false
}

func (u *plainUI) StartRefresh() {
	// No-op for plain UI
}

func (u *plainUI) StopRefresh() {
	// No-op for plain UI
}

// plainSpinner is a no-op spinner for plain output
type plainSpinner struct{}

func (s *plainSpinner) Update(text string) {
	// Updates are silent in plain mode
	_ = text
}

func (s *plainSpinner) Success(text string) {
	fmt.Printf("✓ %s\n", text)
}

func (s *plainSpinner) Fail(text string) {
	fmt.Printf("✗ %s\n", text)
}

// MultiSpinnerUI provides multi-spinner support for refresh operations
type MultiSpinnerUI interface {
	NewWriter() io.Writer
	Start()
	Stop()
}

// NewMultiSpinnerUI creates a multi-spinner UI based on terminal type
func NewMultiSpinnerUI() MultiSpinnerUI {
	if isInteractive() {
		return &ptermMultiSpinner{multi: pterm.DefaultMultiPrinter}
	}
	return &plainMultiSpinner{}
}

// ptermMultiSpinner wraps pterm.MultiPrinter
type ptermMultiSpinner struct {
	multi pterm.MultiPrinter
}

func (m *ptermMultiSpinner) NewWriter() io.Writer {
	return m.multi.NewWriter()
}

func (m *ptermMultiSpinner) Start() {
	m.multi.Start()
}

func (m *ptermMultiSpinner) Stop() {
	m.multi.Stop()
}

// plainMultiSpinner is a simple implementation for non-interactive mode
type plainMultiSpinner struct{}

func (m *plainMultiSpinner) NewWriter() io.Writer {
	return io.Discard // In plain mode, we don't use multi-spinner
}

func (m *plainMultiSpinner) Start() {
	// No-op
}

func (m *plainMultiSpinner) Stop() {
	// No-op
}
