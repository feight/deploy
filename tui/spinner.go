package tui

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type errMsg error

type spinnerModel struct {
	spinner  spinner.Model
	quitting bool
	err      error
	fn       func()
	message  string
}

type Complete struct{}

func initialModel() *spinnerModel {

	s := spinner.New()
	s.Spinner = spinner.MiniDot
	return &spinnerModel{spinner: s}
}

func (m *spinnerModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			m.fn()
			return Complete{}
		})
}

func (m *spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		default:
			return m, nil
		}

	case Complete:
		return m, tea.Quit

	case errMsg:
		m.err = msg
		return m, nil

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m *spinnerModel) View() string {
	if m.err != nil {
		return m.err.Error()
	}
	str := fmt.Sprintf("%s %s", m.spinner.View(), m.message)
	if m.quitting {
		return ""
	}
	return str
}

type Cmd struct {
	message string
	name    string
	arg     []string

	Stdout io.Writer
	Stderr io.Writer
	Env    []string
	Dir    string
}

func (s *Cmd) Run() error {

	var fnErr error

	fn := func() {

		cmd := exec.Command(s.name, s.arg...)

		cmd.Dir = s.Dir
		cmd.Env = s.Env
		cmd.Stdout = s.Stdout
		cmd.Stderr = s.Stderr
		cmd.Stdin = os.Stdin

		fnErr = cmd.Run()
	}

	model := initialModel()
	model.fn = fn
	model.message = s.message

	p := tea.NewProgram(model)

	if _, err := p.Run(); err != nil {
		panic(err)
	}

	if model.quitting {
		os.Exit(0)
	}

	if fnErr != nil {
		return fnErr
	}

	return nil
}

func Command(message string, name string, arg ...string) *Cmd {
	return &Cmd{
		arg:     arg,
		name:    name,
		message: message,
	}
}
