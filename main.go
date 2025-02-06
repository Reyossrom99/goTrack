package main

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbles/textarea"
)

const timeout = time.Second * 5
const gap = "\n\n"
type model struct {
	timer    timer.Model
	keymap   keymap
	help     help.Model
	quitting bool
	task 	 string
	started  bool
	viewport viewport.Model
	textarea textarea.Model
}

type keymap struct {
	start key.Binding
	stop  key.Binding
	reset key.Binding
	quit  key.Binding
	newTimer key.Binding
	newTask key.Binding
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.timer.Init(), textarea.Blink)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)
	m.viewport, vpCmd = m.viewport.Update(msg)
	m.textarea, tiCmd = m.textarea.Update(msg)
	switch msg := msg.(type) {
	case timer.TickMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd

	case timer.StartStopMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		m.keymap.stop.SetEnabled(m.timer.Running())
		m.keymap.start.SetEnabled(!m.timer.Running())
		return m, cmd

	case timer.TimeoutMsg:
		m.quitting = true
		return m, tea.Quit

	case tea.KeyMsg:
		switch{
		//start timer
		case key.Matches(msg, m.keymap.newTimer): 
			m.timer = timer.NewWithInterval(timeout, time.Millisecond)
			m.started = true
			cmd:= m.timer.Start()
			return m, cmd
		case key.Matches(msg, m.keymap.quit):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, m.keymap.reset):
			m.timer.Timeout = timeout
		case key.Matches(msg, m.keymap.start, m.keymap.stop):
			return m, m.timer.Toggle()
		case key.Matches(msg, m.keymap.newTask): 
			m.viewport.SetContent(m.textarea.Value())
			m.textarea.Reset()
			
		}
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m model) helpView() string {
	return "\n" + m.help.ShortHelpView([]key.Binding{
		m.keymap.start,
		m.keymap.stop,
		m.keymap.reset,
		m.keymap.quit,
		m.keymap.newTimer,
		m.keymap.newTask,
	})
}

func (m model) View() string {
	// For a more detailed timer view you could read m.timer.Timeout to get
	// the remaining time as a time.Duration and skip calling m.timer.View()
	// entirely.
	if !m.started{
		return fmt.Sprintf(
		"%s%s%s",
		m.viewport.View(),
		gap,
		m.textarea.View(),
	)
	}else {
		s := m.timer.View()

		if m.timer.Timedout() {
			s = "All done!"
		}
		s += "\n"
		if !m.quitting {
			s = "Working on task: " + m.viewport.View() + "\n" + "Remaining time: " +s
			s += m.helpView()
		}
		return s
	}
}

func main() {
	ta := textarea.New()
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 280

	ta.SetWidth(30)
	ta.SetHeight(3)

	// Remove cursor line styling

	ta.ShowLineNumbers = false
	
	vp := viewport.New(30, 5)
	vp.SetContent("Enter a new task")
	m := model{
		keymap: keymap{
			start: key.NewBinding(
				key.WithKeys("s"),
				key.WithHelp("s", "start"),
			),
			stop: key.NewBinding(
				key.WithKeys("s"),
				key.WithHelp("s", "stop"),
			),
			reset: key.NewBinding(
				key.WithKeys("r"),
				key.WithHelp("r", "reset"),
			),
			quit: key.NewBinding(
				key.WithKeys("q", "ctrl+c"),
				key.WithHelp("q", "quit"),
			),
			newTimer: key.NewBinding(
				key.WithKeys("n"),
				key.WithHelp("n", "new"),

			),
			newTask: key.NewBinding(
				key.WithKeys("a"),
				key.WithHelp("a", "add new task"),

			),

		},
		help: help.New(),
		viewport: vp,
		textarea: ta,
	}
	m.keymap.start.SetEnabled(false)

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Uh oh, we encountered an error:", err)
		os.Exit(1)
	}
}
