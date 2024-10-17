package fetchLogs

import (
	"bytes"
	"docker-tui-go/models"
	"fmt"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// fetchLogsCmd will return a command to fetch logs for the given container
func FetchLogsCmd(container models.Items) tea.Cmd {
	return func() tea.Msg {

		logsChannel := make(chan string)

		go func() {
			logs := fetchLogs(container)
			logsChannel <- logs
		}()

		logs := <-logsChannel
 
	 	return models.LogsFetchedMsg{Logs: logs}
	}
}

func fetchLogs(container models.Items) string {
	cmd := exec.Command("docker", "logs", container.Id)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		return fmt.Sprintf("Error fetching logs: %v", err)
	}

	// Split logs into pages based on terminal height
	return out.String()
}

func SplitIntoPages(logs string) []string {
	lines := strings.Split(logs, "\n")
	pageSize := 20 // Define the number of lines per page (adjust this based on terminal height)
	var pages []string

	for i := 0; i < len(lines); i += pageSize {
		end := i + pageSize
		if end > len(lines) {
			end = len(lines)
		}
		pages = append(pages, strings.Join(lines[i:end], "\n"))
	}

	return pages
}
