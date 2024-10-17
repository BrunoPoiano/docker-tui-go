package main

import (
	"bytes"
	"docker-tui-go/dockerActions"
	"docker-tui-go/fetchLogs"
	"docker-tui-go/models"
	"fmt"

	//	"io"
	"os"
	"os/exec"
	"strings"

	//	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"os/signal"
	"syscall"
)

type model struct {
	items         []models.Items // items on the to-do list
	cursor        int            // which to-do list item our cursor is pointing at
	item_selected models.Items
	action        string
	loading       bool
	logs          Logs
	debug         string

	// lipgloss styles and dimention
	width  int
	height int
	styles *Styles
}

type Logs struct {
	logs        string
	logsPages   []string
	currentPage int
}

type Styles struct {
	BorderColor lipgloss.Color
}

func DefaultStyles() *Styles {
	s := new(Styles)
	s.BorderColor = lipgloss.Color("36")

	return s
}

func getMenuItems() []models.Items {

	menu := []models.Items{
		{Id: "shell", Name: "Shell"},
		{Id: "logs", Name: "Logs"},
		{Id: "stop", Name: "Stop"},
		{Id: "restart", Name: "Restart"},
		{Id: "list", Name: "List"},
	}

	return menu
}

func initialModel(items []models.Items) model {
	styles := DefaultStyles()
	return model{items: items, styles: styles}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func resetModel(m model) tea.Cmd {
	return func() tea.Msg {
		m.cursor = 0
		m.action = ""
		m.logs.logs = ""
		m.logs.currentPage = 0
		m.item_selected = models.Items{}

		return models.Action{Finished: true}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	// get key pressed
	case tea.KeyMsg:

		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}

		// Move to the next log page
		case "l":
			if m.logs.currentPage < len(m.logs.logsPages)-1 {
				m.logs.currentPage++

				return m, cmd
			}

		// Move to the previous log page
		case "h":
			if m.logs.currentPage > 0 {
				m.logs.currentPage--

				return m, cmd
			}

			//Send user to menu lista
		case "m":
			m.cursor = 0
			m.action = ""
			m.logs.logs = ""
			m.logs.currentPage = 0
			m.item_selected = models.Items{}
			m.items = getMenuItems()

		// The "enter" key and the spacebar toggle
		// the container selected
		case "enter", " ":

			switch m.items[m.cursor].Id {
			case "shell", "logs", "stop", "list", "restart":
				m.action = m.items[m.cursor].Id
				m.items = getRunningItems()
				m.cursor = 0

			default:
				m.item_selected = m.items[m.cursor]
			}
		}

		if m.action == "logs" && m.item_selected != (models.Items{}) {
			m.logs.logs = "" // Reset logs
			m.loading = true
			cmd = fetchLogs.FetchLogsCmd(m.item_selected)
		}

		switch m.action {
		case "stop", "restart":
			if m.item_selected != (models.Items{}) {
				m.loading = true
				cmd = dockerActions.CommandItems(m.item_selected, m.action)
			}
		}

	case models.Action:
		m.cursor = 0
		m.action = ""
		m.logs.logs = ""
		m.logs.currentPage = 0
		m.item_selected = models.Items{}
		m.items = getMenuItems()
		m.loading = !msg.Finished

	case models.LogsFetchedMsg:
		// Once logs are fetched, update the model with the logs
		m.logs.logs = msg.Logs
		m.logs.logsPages = fetchLogs.SplitIntoPages(m.logs.logs, m.height)
		m.loading = false

		if len(m.logs.logsPages) > 0 {
			m.logs.currentPage = len(m.logs.logsPages) - 1
		}
	}

	return m, cmd
}

func (m model) View() string {

	// The header
	header := "Docker-TUI"
	padding := (m.width - len(header)) / 2

	if padding < 0 {
		padding = 0
	}

	content := fmt.Sprintf("%s%s\n\n", strings.Repeat(" ", padding), header)

	content += fmt.Sprintf("action selected %s \n\n", m.action)
	content += fmt.Sprintf("Items selected %s \n\n", m.item_selected.Name)
	content += fmt.Sprintf("debug %s \n\n", m.debug)

	if m.loading {
		content += "Loading ... \n"
	} else if m.action == "logs" && m.item_selected != (models.Items{}) {

		if len(m.logs.logsPages) > 0 {
			content += m.logs.logsPages[m.logs.currentPage]
			content += fmt.Sprintf("\n\nPage %d/%d", m.logs.currentPage+1, len(m.logs.logsPages))
		} else {
			content += "No available \n"
		}

	} else {

		// Iterate over our choices
		for i, choice := range m.items {

			// Is the cursor pointing at this choice?
			cursor := " " // no cursor
			if m.cursor == i {
				cursor = ">" // cursor!
			}

			// Render the row
			content += fmt.Sprintf("%s %s\n", cursor, choice.Name)
		}

		// The footer
		content += "\nPress q to quit.\n"

	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true, true, true, true).
		BorderForeground(lipgloss.Color("32")).
		Padding(2).
		Margin(2).
		Width(m.width - 6).
		Height(m.height - 12).
		Render(content)
}

func main() {

	//containers := getRunningItemss()
	menu := getMenuItems()

	p := tea.NewProgram(initialModel(menu), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func getRunningItems() []models.Items {

	cmd := exec.Command("docker", "ps", "--format", "{{.ID}} {{.Names}}")

	var containers []models.Items
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		fmt.Println("Error:", err)
		return containers
	}

	cmdReturn := strings.Split(out.String(), "\n")
	for _, item := range cmdReturn {
		itemFormated := strings.Split(item, " ")

		if len(itemFormated) == 2 {
			containers = append(containers, models.Items{
				Id:   itemFormated[0],
				Name: itemFormated[1],
			})
		}
	}

	return containers

}

func shellItems(container models.Items) {
	cmd := exec.Command("docker", "exec", "-it", container.Id, "/bin/sh")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signals
		if err := cmd.Process.Signal(syscall.SIGINT); err != nil {
			fmt.Println("Error sending signal to Docker process:", err)
		}
	}()

	fmt.Println("Running Command")

	if err := cmd.Start(); err != nil {
		fmt.Printf("Error starting Docker: %s\n", err)
		return
	}

	if err := cmd.Wait(); err != nil {
		fmt.Printf("Error waiting for Docker: %s\n", err)
	}

}
