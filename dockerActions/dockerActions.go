package dockerActions

import (
	"bytes"
	"docker-tui-go/models"
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

func CommandItems(container models.Items, command string) tea.Cmd {

	return func() tea.Msg {

		cmd := exec.Command("docker", command, container.Id)

		var out bytes.Buffer
		cmd.Stdout = &out

		err := cmd.Run()
		if err != nil {
			fmt.Printf("Error docker %s \n", command)
			return models.Action{Error: "Error running Command"}
		}

		return models.Action{Finished: true}
	}
}
