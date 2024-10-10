package main

import (
	"bytes"
	"fmt"

	//	"io"
	"os"
	"os/exec"
	"strings"

	//	"github.com/charmbracelet/bubbles/list"
	//tea "github.com/charmbracelet/bubbletea"
	//	"github.com/charmbracelet/lipgloss"

	"os/signal"
	"syscall"
)

type Container struct {
	id   string
	name string
}

func commandContainer(container Container, command string) {

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

func logsContainer(container Container) {
	cmd := exec.Command("docker", "logs", "--follow", container.id)

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

func shellContainer(container Container) {
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

func getRunningContainers() ([]Container){

  	cmd := exec.Command("docker", "ps", "--format", "{{.ID}} {{.Names}}")

	var containers []Container
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
			containers = append(containers, Container{
				id:   itemFormated[0],
				name: itemFormated[1],
			})
		}
	}

  return containers

}

func main() {

	args := os.Args

	var containers []Container = getRunningContainers()

	switch args[1] {
	case "-h":
		fmt.Printf("shell - to conect to a container \n")
		fmt.Printf("logs  - to check container logs \n")
		fmt.Printf("stop  - to stop a container \n")
		fmt.Printf("list  - to list running containers \n")

	case "stop":
		commandContainer(containers[0], "stop")

	case "restart":
		commandContainer(containers[0], "restart")

	case "shell":
		shellContainer(containers[0])

	case "logs":
		logsContainer(containers[0])

	case "list":
		for _, container := range containers {
			fmt.Printf("Container ID: %s, Name: %s\n", container.id, container.name)
		}

	default:
		fmt.Println("Unknown command:", args[1])
	}
}
