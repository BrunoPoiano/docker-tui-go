package appActions

import (
	"bytes"
	"context"
	"docker-tui-go/models"
	"fmt"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	//"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	dockerClient "github.com/docker/docker/client"
)

func CommandItem(container models.Items, command string) tea.Cmd {

	return func() tea.Msg {

		cmd := exec.Command("docker", command, container.Id)

		var out bytes.Buffer
		cmd.Stdout = &out

		err := cmd.Run()
		if err != nil {
			return models.Action{Error: fmt.Sprintf("%v", err), Finished: true}
		}

		return models.Action{Finished: true}
	}

}

func DefaultStyles() *models.Styles {
	s := new(models.Styles)
	s.BorderColor = lipgloss.Color("36")

	return s
}

func GetMenuItems() []models.Items {

	menu := []models.Items{
		{Id: "menu", Name: "Menu", Command: "M"},
		{Id: "shell", Name: "Shell", Command: "S"},
		{Id: "logs", Name: "Logs", Command: "L"},
		{Id: "start", Name: "Start", Command: "A"},
		{Id: "stop", Name: "Stop", Command: "T"},
		{Id: "restart", Name: "Restart", Command: "R"},
		{Id: "list", Name: "List", Command: "I"},
	}

	return menu
}

func GetStoppedItems() []models.Items {

	cmd := exec.Command(
		"docker",
		"ps",
		"-a",
		"--filter",
		"status=exited",
		"--filter",
		"status=created",
		"--format",
		"{{.ID}} {{.Names}}")

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

func GetRunningItems(cli *dockerClient.Client) []models.Items {

	var containersList []models.Items
	containers, err := cli.ContainerList(context.Background(), container.ListOptions{})

	if err != nil {
		panic(err)
	}

	for _, ctr := range containers {
		containersList = append(containersList, models.Items{
			Id:   ctr.ID,
			Name: ctr.Names[0],
		})
	}

	return containersList

}

func GetAllContainers(cli *dockerClient.Client, appWidth int) []models.Items {

	var containersList []models.Items
	containers, err := cli.ContainerList(context.Background(), container.ListOptions{All: true})

	width := 105
	if appWidth < 105 {
		// appWidth - appPadding / ( two columns, Id, name) - third column, status
		width = (appWidth-2)/2 - 10
	} else {
		width = (105-2)/2 - 10
	}

	if err != nil {
		panic(err)
	}

  headerString := fmt.Sprintf("%-*s | %-*s | Status", width, "Id", width, "Image")
	header := models.Items{Id: "-", Name: headerString}
	containersList = append(containersList, header)

	for _, ctr := range containers {

		status := "Running"
		if strings.Contains(ctr.Status, "Exited") {
			status = "Exited"
		}

		name := TruncateWithEllipsis(ctr.Names[0], width)
		image := TruncateWithEllipsis(ctr.Image, width)

		row := fmt.Sprintf("%-*s | %-*s | %s", width, name, width, image, status)
		containersList = append(containersList, models.Items{
			Id:   ctr.ID,
			Name: row,
		})
	}

	return containersList

}

func TruncateWithEllipsis(s string, width int) string {
	if len(s) > width {
		return s[:width-3] + "..." // Truncate and add "..."
	}
	return s
}
