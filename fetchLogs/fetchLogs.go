package fetchLogs

import (
	"bytes"
	"context"
	"docker-tui-go/models"
	"fmt"
	"io"
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	dockerContainer "github.com/docker/docker/api/types/container"
	dockerClient "github.com/docker/docker/client"
)

// fetchLogsCmd will return a command to fetch logs for the given container
func FetchLogsCmd(cli *dockerClient.Client, container models.Items) tea.Cmd {
	return func() tea.Msg {
		logs := containerLogs(cli, container)
		return models.LogsFetchedMsg{Logs: logs}
	}
}

func containerLogs(cli *dockerClient.Client, containerSelected models.Items) string {

	logs, err := cli.ContainerLogs(context.Background(), containerSelected.Id, dockerContainer.LogsOptions{ShowStdout: true, ShowStderr: true, Timestamps: false, Details: true})
	if err != nil {
		log.Fatalf("Error retrieving logs: %v", err)
	}

	var out bytes.Buffer
	_, err = io.Copy(&out, logs)
	if err != nil && err != io.EOF {
		fmt.Sprintf("%v", err)
	}

	return out.String()

}

func SplitIntoPages(logs string, height int) []string {
	lines := strings.Split(logs, "\n")
	pageSize := height - 15 //  number of lines per page ( based on terminal height)
	var pages []string

	for i := 0; i < len(lines); i += pageSize {
		end := i + pageSize
		if end > len(lines) {
			end = len(lines)
		}
    line := strings.Join(lines[i:end], " \n")
    pages = append(pages, line)
	}

	return pages
}
