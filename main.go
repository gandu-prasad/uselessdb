package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"key_value_db/store"
)

const maxLines = 15

type model struct {
	db     *store.Store
	input  textinput.Model
	output []string
}

func initialModel(db *store.Store) model {
	ti := textinput.New()
	ti.Placeholder = "put KEY VALUE | get KEY | delete KEY | compact"
	ti.CharLimit = 4096
	ti.Width = 60
	ti.Focus()
	return model{db: db, input: ti}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			line := strings.TrimSpace(m.input.Value())
			m.input.SetValue("")
			if line == "" {
				return m, nil
			}
			m.output = append(m.output, "> "+line, m.run(line))
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) run(line string) string {
	parts := strings.SplitN(line, " ", 3)
	switch parts[0] {
	case "put":
		if len(parts) != 3 {
			return "usage: put KEY VALUE"
		}
		if err := m.db.Put([]byte(parts[1]), []byte(parts[2])); err != nil {
			return "error: " + err.Error()
		}
		return "ok"

	case "get":
		if len(parts) != 2 {
			return "usage: get KEY"
		}
		v, err := m.db.Get([]byte(parts[1]))
		if err != nil {
			return "error: " + err.Error()
		}
		return string(v)

	case "delete":
		if len(parts) != 2 {
			return "usage: delete KEY"
		}
		if err := m.db.Delete([]byte(parts[1])); err != nil {
			return "error: " + err.Error()
		}
		return "ok"

	case "compact":
		if err := m.db.Compact(); err != nil {
			return "error: " + err.Error()
		}
		return "compacted"

	default:
		return fmt.Sprintf("unknown command %q", parts[0])
	}
}

func (m model) View() string {
	var b strings.Builder

	start := 0
	if len(m.output) > maxLines {
		start = len(m.output) - maxLines
	}
	for _, line := range m.output[start:] {
		b.WriteString(line + "\n")
	}

	b.WriteString("\n" + m.input.View() + "\n\n")
	b.WriteString("(esc to quit)\n")
	return b.String()
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run keeps the defer on db.Close in effect, since os.Exit skips defers.
func run() error {
	db, err := store.Open("store.db")
	if err != nil {
		return err
	}
	defer db.Close()

	p := tea.NewProgram(initialModel(db), tea.WithAltScreen())
	_, err = p.Run()
	return err
}
