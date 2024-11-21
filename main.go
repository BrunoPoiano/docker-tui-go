package main

import (
	"docker-tui-go/appActions"
	"docker-tui-go/dockerShell"
	"docker-tui-go/fetchLogs"
	"docker-tui-go/models"
	"fmt"

	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/docker/docker/client"

	dockerClient "github.com/docker/docker/client"
)

type Model struct {
	Items          []models.Items
	Cursor         int
	ItemSelected   models.Items
	Action         string
	Loading        bool
	LoadingMessage string
	Logs           models.Logs
	Debug          string
	cli            *dockerClient.Client
	Width          int
	Height         int
	Styles         *models.Styles
}

func initialModel(items []models.Items, cli *dockerClient.Client) Model {
	styles := appActions.DefaultStyles()
	return Model{Items: items, Styles: styles, cli: cli}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func menuActions(m Model, action string) Model {
	m.Action = action
	m.Cursor = 0

	switch action {
	case "start":
		m.Items = appActions.GetStoppedItems()
	case "list":
		m.Items = appActions.GetAllContainers(m.cli, m.Width)
	default:
		m.Items = appActions.GetRunningItems(m.cli)
	}

	return m
}

func resetMenu(m Model) Model {
	m.Cursor = 0
	m.Action = ""
	m.Logs.Logs = ""
	m.Logs.CurrentPage = 0
	m.ItemSelected = models.Items{}
	m.Items = appActions.GetMenuItems()

	return m
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
		case "ctrl+c", "Q":
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
			m = menuActions(m, "restart")
			return m, cmd

		case "I":
			m = menuActions(m, "list")
			return m, cmd

			// Logsss
		case "A":
			m = menuActions(m, "start")
			return m, cmd

		case "S":
			m = menuActions(m, "shell")
			return m, cmd

		case "T":
			m = menuActions(m, "stop")
			return m, cmd

		case "L":
			m = menuActions(m, "logs")
			return m, cmd

			//Send user to menu lista
		case "M":
			m = resetMenu(m)

		// the container selected
		case "enter", " ":

			switch m.Items[m.Cursor].Id {
			case "shell", "logs", "stop", "restart":
				m.Action = m.Items[m.Cursor].Id
				m.Items = appActions.GetRunningItems(m.cli)
				m.Cursor = 0

			case "list":
				m.Action = m.Items[m.Cursor].Id
				m.Items = appActions.GetAllContainers(m.cli, m.Width)
				m.Cursor = 0

			case "start":
				m.Action = m.Items[m.Cursor].Id
				m.Items = appActions.GetStoppedItems()
				m.Cursor = 0

			default:
				m.ItemSelected = m.Items[m.Cursor]
			}
		}

		if m.Action == "logs" && m.ItemSelected != (models.Items{}) {
			m.Logs.Logs = "" // Reset logs
			m.Loading = true
			cmd = fetchLogs.FetchLogsCmd(m.cli, m.ItemSelected)
		}

		if m.Action == "shell" && m.ItemSelected != (models.Items{}) {
			m.Loading = true
			cmd = dockerShell.Dockershell(m.cli, m.ItemSelected)
		}
		switch m.Action {
		case "stop", "restart", "start":
			if m.ItemSelected != (models.Items{}) {
				m.Loading = true
				m.LoadingMessage = fmt.Sprintf("%sing %s", m.Action, m.ItemSelected.Name)
				cmd = appActions.CommandItem(m.ItemSelected, m.Action)

			}
		}

	case models.Action:
		m = resetMenu(m)
		m.Debug = msg.Error
		m.Loading = !msg.Finished

	case models.ShellFetchMsg:
		m = resetMenu(m)
		m.Debug = msg.Error
		m.Loading = false

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

func (m Model) View() string {

	tea.ClearScreen()
	if m.Action == "shell" && m.ItemSelected != (models.Items{}) {
		return ""
	}
	content := []string{}

	// The header
	header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("1")).Render("DOCKER-TUI \n\n")
	// Header
	content = append(content, header)

	menu := appActions.GetMenuItems()

	actions := "|"
	for _, item := range menu {

		menuItem := fmt.Sprintf(" %s: %s ", item.Name, item.Command)
		if m.Action == item.Id {
			actions += lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("1")).Foreground(lipgloss.Color("#FFFFFF")).Render(menuItem)
		} else if m.Action == "" && item.Id == "menu" {
			actions += lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("1")).Foreground(lipgloss.Color("#FFFFFF")).Render(menuItem)
		} else {
			actions += lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render(menuItem)
		}

		actions += "|"
	}
	actions += "\n\n"

	content = append(content, actions)
	//content = append(content, fmt.Sprintf("action selected %s \n\n", m.Action))
	//content = append(content, fmt.Sprintf("Items selected %s \n\n", m.ItemSelected.Name))
	//content = append(content, fmt.Sprintf("debug %s \n\n", m.Debug))
	//content = append(content, fmt.Sprintf("width %d \n\n", m.Width))

	// Loading message
	if m.Loading {
		content = append(content, "Loading ... \n\n")

		loadingMessage := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("1")).Render(m.LoadingMessage)
		content = append(content, loadingMessage+" \n\n")
	} else if m.Action == "logs" && m.ItemSelected != (models.Items{}) {
		// Logs view
		if len(m.Logs.LogsPages) > 0 {
			content = append(content, m.Logs.LogsPages[m.Logs.CurrentPage])
			content = append(content, fmt.Sprintf("\n\nPage %d/%d", m.Logs.CurrentPage+1, len(m.Logs.LogsPages)))
		} else {
			content = append(content, "No available logs \n")
		}
	} else {
		for i, choice := range m.Items {
			Cursor := "" // no cursor
			if m.Cursor == i && m.Action != "list" {
				Cursor = ">" // cursor at this choice!
			}
			// Render the row with the Cursor
			content = append(content, fmt.Sprintf("%s %d: %s\n", Cursor, i, choice.Name))
		}
	}

	// Footer
	width := 10

	footerItems := []models.FooterItem{
		{Label: "Quit", Action: "Q"},
		{Label: "Select", Action: "enter"},
		{Label: "Up", Action: "k"},
		{Label: "Down", Action: "j"},
		{Label: "Left/PgDn", Action: "h"},
		{Label: "Right/PgUp", Action: "l"},
	}

	footerString := "\n\n"
	for index, item := range footerItems {
		footerString += fmt.Sprintf("%-*s %-*s", width, item.Label, width, item.Action)

		if index%2 != 0 {
			footerString += "\n"
		}
	}

	footer := lipgloss.NewStyle().Foreground(lipgloss.Color("#9e9e9e")).Render(footerString)
	content = append(content, footer)

	// Combine content into a single string
	finalContent := strings.Join(content, "")

	padding := 2
	// Render the styled content
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true, true, true, true).
		BorderForeground(lipgloss.Color("1")).
		Padding(padding).
		Width(m.Width - padding).
		Render(finalContent)
}

func main() {

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	//containers := getRunningItemss()
	menu := appActions.GetMenuItems()

	p := tea.NewProgram(initialModel(menu, cli), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
