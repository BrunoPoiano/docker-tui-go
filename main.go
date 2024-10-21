package main

import (
	"docker-tui-go/appActions"
	"docker-tui-go/fetchLogs"
	"docker-tui-go/models"
	"fmt"

	//	"io"
	"os"
	"strings"

	//	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	Items        []models.Items // items on the to-do list
	Cursor       int            // which to-do list item our cursor is pointing at
	ItemSelected models.Items
	Action       string
	Loading      bool
	Logs         models.Logs
	Debug        string

	// lipgloss styles and dimention
	Width  int
	Height int
	Styles *models.Styles
}

func initialModel(items []models.Items) Model {
	styles := appActions.DefaultStyles()
	return Model{Items: items, Styles: styles}
}

func (m Model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

	// get key pressed
	case tea.KeyMsg:

		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "up" and "k" keys move the Cursor up
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}

		// The "down" and "j" keys move the Cursor down
		case "down", "j":
			if m.Cursor < len(m.Items)-1 {
				m.Cursor++
			}

		// Move to the next log page
		case "l":
			if m.Logs.CurrentPage < len(m.Logs.LogsPages)-1 {
				m.Logs.CurrentPage++

				return m, cmd
			}

		// Move to the previous log page
		case "h":
			if m.Logs.CurrentPage > 0 {
				m.Logs.CurrentPage--

				return m, cmd
			}

		case "R":
			m.Action = "restart"
			m.Items = appActions.GetRunningItems()
			m.Cursor = 0
			return m, cmd

		case "I":
			m.Action = "I"
			m.Items = appActions.GetRunningItems()
			m.Cursor = 0
			return m, cmd

			// Logsss
		case "S":
			m.Action = "shell"
			m.Items = appActions.GetRunningItems()
			m.Cursor = 0
			return m, cmd

		case "T":
			m.Action = "stop"
			m.Items = appActions.GetRunningItems()
			m.Cursor = 0
			return m, cmd

		case "L":
			m.Action = "logs"
			m.Items = appActions.GetRunningItems()
			m.Cursor = 0
			return m, cmd

			//Send user to menu lista
		case "M":
			m.Cursor = 0
			m.Action = ""
			m.Logs.Logs = ""
			m.Logs.CurrentPage = 0
			m.ItemSelected = models.Items{}
			m.Items = appActions.GetMenuItems()

		// The "enter" key and the spacebar toggle
		// the container selected
		case "enter", " ":

			switch m.Items[m.Cursor].Id {
			case "shell", "logs", "stop", "list", "restart":
				m.Action = m.Items[m.Cursor].Id
				m.Items = appActions.GetRunningItems()
				m.Cursor = 0

			default:
				m.ItemSelected = m.Items[m.Cursor]
			}
		}

		if m.Action == "logs" && m.ItemSelected != (models.Items{}) {
			m.Logs.Logs = "" // Reset logs
			m.Loading = true
			cmd = fetchLogs.FetchLogsCmd(m.ItemSelected)
		}

		switch m.Action {
		case "stop", "restart":
			if m.ItemSelected != (models.Items{}) {
				m.Loading = true
				cmd = appActions.CommandItems(m.ItemSelected, m.Action)
			}
		}

	case models.Action:
		m.Cursor = 0
		m.Action = ""
		m.Logs.Logs = ""
		m.Logs.CurrentPage = 0
		m.ItemSelected = models.Items{}
		m.Items = appActions.GetMenuItems()
		m.Loading = !msg.Finished

	case models.LogsFetchedMsg:
		// Once logs are fetched, update the model with the logs
		m.Logs.Logs = msg.Logs
		m.Logs.LogsPages = fetchLogs.SplitIntoPages(m.Logs.Logs, m.Height)
		m.Loading = false

		if len(m.Logs.LogsPages) > 0 {
			m.Logs.CurrentPage = len(m.Logs.LogsPages) - 1
		}
	}

	return m, cmd
}

// lipgloss color cheat sheet
var colors = []lipgloss.Color{
	lipgloss.Color("1"), // Red
	lipgloss.Color("2"), // Green
	lipgloss.Color("3"), // Yellow
	lipgloss.Color("4"), // Blue
	lipgloss.Color("5"), // Magenta
	lipgloss.Color("6"), // Cyan
	lipgloss.Color("7"), // White
}

func (m Model) View() string {

	content := []string{}

	// The header
	header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("1")).Render("DOCKER-TUI \n\n")

	// Header
	content = append(content, header)

	// Action, selected item, and debug
	content = append(content, fmt.Sprintf("action selected %s \n\n", m.Action))
	content = append(content, fmt.Sprintf("Items selected %s \n\n", m.ItemSelected.Name))
	content = append(content, fmt.Sprintf("debug %s \n\n", m.Debug))

	// Loading message
	if m.Loading {
		content = append(content, "Loading ... \n")
	} else if m.Action == "logs" && m.ItemSelected != (models.Items{}) {
		// Logs view
		if len(m.Logs.LogsPages) > 0 {
			content = append(content, m.Logs.LogsPages[m.Logs.CurrentPage])
			content = append(content, fmt.Sprintf("\n\nPage %d/%d", m.Logs.CurrentPage+1, len(m.Logs.LogsPages)))
		} else {
			content = append(content, "No available logs \n")
		}
	} else {
		// Iterating over choices
		for i, choice := range m.Items {
			Cursor := " " // no cursor
			if m.Cursor == i {
				Cursor = ">" // cursor at this choice!
			}
			// Render the row with the Cursor
			content = append(content, fmt.Sprintf("%s %s\n", Cursor, choice.Name))
		}
	}

	actions := lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Render("\n\n\n Menu M | Shell: S | Logs L |Stop T | Restart R | List I")
	// Footer
	footer := lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Render("\n Quit:  q | Up: j | Down: k | Left: h | Right: l \n")
	content = append(content, actions)
	content = append(content, footer)

	// Combine content into a single string
	finalContent := strings.Join(content, "")

	// Render the styled content
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true, true, true, true).
		BorderForeground(lipgloss.Color("32")).
		Padding(2).
		Margin(2).
		Width(m.Width).
		Height(m.Height - 5).
		Render(finalContent)
}

func main() {

	//containers := getRunningItemss()
	menu := appActions.GetMenuItems()

	p := tea.NewProgram(initialModel(menu), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
