package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type doneMsg struct{ err error }

func RunSpinner(message string, fn func() error) error {
	m := spinnerModel{
		spinner: spinner.New(),
		message: message,
	}
	m.spinner.Spinner = spinner.Dot
	m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("207"))

	p := tea.NewProgram(m, tea.WithoutSignalHandler(), tea.WithInput(nil))

	go func() {
		p.Send(doneMsg{err: fn()})
	}()

	final, err := p.Run()
	if err != nil {
		return err
	}
	if sm, ok := final.(spinnerModel); ok {
		return sm.err
	}
	return nil
}

type spinnerModel struct {
	spinner spinner.Model
	message string
	err     error
}

func (m spinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case doneMsg:
		m.err = msg.(doneMsg).err
		return m, tea.Quit
	case nil:
		return m, tea.Quit
	}
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m spinnerModel) View() string {
	return fmt.Sprintf("  %s %s", m.spinner.View(), m.message)
}

type ProgressFunc func(downloaded, total int64)

func PrintDownloadProgress(message string, download func(ProgressFunc) error) error {
	fmt.Printf("  📦 %s\n", message)

	barWidth := 30
	var downloaded int64

	onProgress := func(d, t int64) {
		downloaded = d
		if t <= 0 {
			return
		}
		pct := float64(d) / float64(t)
		if pct > 1 {
			pct = 1
		}
		filled := int(pct * float64(barWidth))
		bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
		fmt.Printf("\r  [%s] %s/%s  ", bar, formatBytes(d), formatBytes(t))
	}

	done := make(chan error, 1)
	go func() {
		done <- download(onProgress)
	}()

	err := <-done

	if downloaded > 0 {
		pct := float64(downloaded) / float64(downloaded)
		if pct > 1 {
			pct = 1
		}
		filled := int(pct * float64(barWidth))
		bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
		fmt.Printf("\r  [%s] %s          ", bar, formatBytes(downloaded))
	}
	fmt.Println()
	return err
}

func formatBytes(b int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case b >= GB:
		return fmt.Sprintf("%.1f GB", float64(b)/float64(GB))
	case b >= MB:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(MB))
	case b >= KB:
		return fmt.Sprintf("%.1f KB", float64(b)/float64(KB))
	default:
		return fmt.Sprintf("%d B", b)
	}
}
