package main

import (
	"bytes"
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

type Items struct {
	id   string
	name string
}

type model struct {
	items         []Items // items on the to-do list
	cursor        int     // which to-do list item our cursor is pointing at
	item_selected Items
	action        string
	loading       bool
	logs          string

	// lipgloss styles and dimention
	width  int
	height int
	styles *Styles
}
type Styles struct {
	BorderColor lipgloss.Color
}

func DefaultStyles() *Styles {
	s := new(Styles)
	s.BorderColor = lipgloss.Color("36")

	return s
}

func getMenuItems() []Items {

	menu := []Items{
		{id: "shell", name: "Shell"},
		{id: "logs", name: "Logs"},
		{id: "stop", name: "Stop"},
		{id: "list", name: "List"},
	}

	return menu
}

func initialModel(items []Items) model {
	styles := DefaultStyles()
	return model{items: items, styles: styles}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
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

			//Send user to menu lista
		case "m":
			m.cursor = 0
			m.action = ""
			m.item_selected = Items{}
			m.items = getMenuItems()

		// The "enter" key and the spacebar toggle
		// the container selected
		case "enter", " ":

			switch m.items[m.cursor].id {
			case "shell", "logs", "stop", "list":
				m.action = m.items[m.cursor].id
				m.items = getRunningItems()
        m.cursor = 0
			default:
				m.item_selected = m.items[m.cursor]
			}
		}

		if m.action == "logs" && m.item_selected != (Items{}) {
			m.logs = "" // Reset logs
			m.loading = true
			// Fetch the logs asynchronously
			go func() {
				m.logs = fetchLogs(m.item_selected)
				m.loading = false
			}()
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
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
	content += fmt.Sprintf("Items selected %s \n\n", m.item_selected.name)

	if m.action == "logs" && m.item_selected != (Items{}) {
		content += m.logs
	} else {

		// Iterate over our choices
		for i, choice := range m.items {

			// Is the cursor pointing at this choice?
			cursor := " " // no cursor
			if m.cursor == i {
				cursor = ">" // cursor!
			}

			// Render the row
			content += fmt.Sprintf("%s %s\n", cursor, choice.name)
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

func getRunningItems() []Items {

	cmd := exec.Command("docker", "ps", "--format", "{{.ID}} {{.Names}}")

	var containers []Items
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
			containers = append(containers, Items{
				id:   itemFormated[0],
				name: itemFormated[1],
			})
		}
	}

	return containers

}

func commandItems(container Items, command string) {

	fmt.Printf("running docker %s \n", command)

	fmt.Println(container)

	cmd := exec.Command("docker", command, container.id)

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error docker %s \n", command)
		return
	}

	fmt.Println(out.String())
}

func fetchLogs(container Items) string {
	cmd := exec.Command("docker", "logs", container.id)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		return fmt.Sprintf("Error fetching logs: %v", err)
	}

	return out.String()
}

func shellItems(container Items) {
	cmd := exec.Command("docker", "exec", "-it", container.id, "/bin/sh")

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
