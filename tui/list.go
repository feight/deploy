package tui

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
)

var (
	listHeight        = 16
	titleStyle        = lipgloss.NewStyle().MarginLeft(0)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("244")) // XTerm colors
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(2)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(2).PaddingBottom(1)
	historyFilename   = path.Join(os.TempDir(), "newsteam.deploy.json")
)

type Option interface {
	SetKey(string)
	Text() string
}

type History struct {
	Value string
}

type Item struct {
	Key  string
	Text string
}

func (i Item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {

	i, ok := listItem.(Item)
	if !ok {
		return
	}

	str := i.Text

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type model struct {
	list     list.Model
	choice   string
	quitting bool
}

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(Item)
			if ok {
				m.choice = i.Key
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *model) View() string {

	if m.choice != "" {
		return ""
	}
	if m.quitting {
		return "\n"
	}
	return "\n" + m.list.View()
}

func RenderList[T Option](opts map[string]T, key string, message string) T {

	var (
		defaultWidth = 20
		sorted       = []Item{}
		items        = []list.Item{}
		hr           = strings.Repeat("-", len(message))
	)

	for key, opt := range opts {
		if !reflect.ValueOf(opt).IsNil() {
			sorted = append(sorted, Item{Key: key, Text: opt.Text()})
		}
	}

	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Key < sorted[j].Key })

	for _, v := range sorted {
		items = append(items, Item{Key: v.Key, Text: v.Text})
	}

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.DisableQuitKeybindings()
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Title = color.HiGreenString(fmt.Sprintf("%s\n%s\n%s", hr, message, hr))
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m := model{list: l}

	if history, ok := getHistory()[key]; ok {
		for i, item := range items {
			if item.(Item).Key == history.Value {
				m.list.Select(i)
				break
			}
		}
	}

	if _, err := tea.NewProgram(&m).Run(); err != nil {
		log.Fatal("error running program:", err)
	}

	if m.choice == "" {
		os.Exit(0)
	}

	writeHistory(key, m.choice)

	var val T
	for key, opt := range opts {
		if key == m.choice {
			val = opt
			val.SetKey(key)
		}
	}

	return val
}

func getHistory() map[string]*History {

	dst := map[string]*History{}
	if bin, err := os.ReadFile(historyFilename); err == nil {
		json.Unmarshal(bin, &dst)
	}

	return dst
}

func writeHistory(key, value string) {

	dst := getHistory()
	dst[key] = &History{Value: value}
	bin, _ := json.Marshal(dst)
	os.WriteFile(historyFilename, bin, os.ModePerm)
}
