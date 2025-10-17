package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/stopwatch"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	webhook "github.com/jimtang2/grok-webhook"
)

type Model struct {
	table table.Model
	watch stopwatch.Model
	rows  []table.Row
	last  time.Time
}

func NewModel() *Model {
	m := &Model{
		rows:  []table.Row{},
		watch: stopwatch.NewWithInterval(time.Second),
		last:  time.Now(),
	}
	styles := table.DefaultStyles()
	// styles.Header = styles.Header.Copy()
	styles.Header = styles.Header.Copy().Padding(0, 0)
	styles.Cell = styles.Cell.Copy().Padding(0, 0).Foreground(lipgloss.Color("42"))
	styles.Selected = styles.Cell
	m.table = table.New(
		table.WithColumns([]table.Column{
			{Title: "Project/Branch", Width: 20},
			{Title: "Files", Width: 80},
		}),
		table.WithRows(m.rows),
		table.WithStyles(styles),
	)
	return m
}

func (m *Model) Init() tea.Cmd {
	return m.watch.Init()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.table.SetWidth(msg.Width)
	case *webhook.Message:
		if time.Now().Sub(m.last) > 5*time.Second {
			m.last = time.Now()
			m.rows = []table.Row{}
		}
		m.rows = append(m.rows, table.Row{
			fmt.Sprintf("%v - %v", msg.Project, msg.Branch),
			msg.File,
		})
		m.table.SetRows(m.rows)
	}
	m.watch, cmd = m.watch.Update(msg)
	return m, cmd
}

func (m *Model) View() string {
	d := time.Since(m.last)
	return strings.Join([]string{
		m.table.View(),
		fmt.Sprintf("Idle for %vm%vs",
			int(d.Minutes()),
			int(d.Seconds())%60,
		),
	}, "\n")
}

func (m *Model) run() error {
	_, err := tea.NewProgram(m).Run()
	return err
}
