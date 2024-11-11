package dockerShell

import (
	"context"
	"docker-tui-go/models"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/docker/docker/api/types/container"
	dockerClient "github.com/docker/docker/client"
	"golang.org/x/term"
)

func Dockershell(cli *dockerClient.Client, selectedContainer models.Items) tea.Cmd {
	return func() tea.Msg {

		// create file
		file, _ := os.Create("file.txt")
		defer file.Close()


			_, err := file.WriteString(fmt.Sprintf("Iniciating terminal process") + "\n")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Setup signal handling
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigChan
			cancel()
		}()

		// Create exec configuration
		execConfig := container.ExecOptions{
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Tty:          true,
			Cmd:          []string{"/bin/bash"},
		}

		// Create exec instance
		execID, err := cli.ContainerExecCreate(ctx, selectedContainer.Id, execConfig)
		if err != nil {

			_, err := file.WriteString(fmt.Sprintf("Error creating exec: %v", err) + "\n")
			return models.ShellFetchMsg{
				Error:    fmt.Sprintf("Error creating exec: %v", err),
				Finished: true,
			}
		}

		// Attach to the exec instance
		resp, err := cli.ContainerExecAttach(ctx, execID.ID, container.ExecAttachOptions{
			Tty: true,
		})
		if err != nil {

			_, err := file.WriteString(fmt.Sprintf("Error attaching to exec: %v", err) + "\n")
			return models.ShellFetchMsg{
				Error:    fmt.Sprintf("Error attaching to exec: %v", err),
				Finished: true,
			}
		}
		defer resp.Close()

		// Get the terminal settings
		fd := int(os.Stdin.Fd())
		oldState, err := term.GetState(fd)
		if err != nil {

			_, err := file.WriteString(fmt.Sprintf("Error getting terminal state:", err) + "\n")
			return models.ShellFetchMsg{
				Error:    fmt.Sprintf("Error attaching to exec: %v", err),
				Finished: true,
			}
		}
		defer term.Restore(fd, oldState)

		// Put terminal into raw mode
		rawState, err := term.MakeRaw(fd)
		if err != nil {

			_, err := file.WriteString(fmt.Sprintf("Error setting raw terminal:", err) + "\n")
			return models.ShellFetchMsg{
				Error:    fmt.Sprintf("Error setting raw terminal: %v", err),
				Finished: true,
			}
		}
		defer term.Restore(fd, rawState)

		// Channel to signal when copying is done
		doneChan := make(chan struct{}, 2)
		// Create a wait group to manage goroutines
		var wg sync.WaitGroup



		file.WriteString(fmt.Sprintf("Goroutine for copying container output to stdout \n"))
		// Goroutine for copying container output to stdout
		wg.Add(1)
		go func() {
			defer func() { doneChan <- struct{}{} }()
			io.Copy(os.Stdout, resp.Reader)
      defer wg.Done()
		}()


		file.WriteString(fmt.Sprintf("Goroutine for copying stdin to container \n"))
		// Goroutine for copying stdin to container
		wg.Add(1)
		go func() {
			defer func() { doneChan <- struct{}{} }()
			io.Copy(resp.Conn, os.Stdin)
      defer wg.Done()
		}()


		file.WriteString(fmt.Sprintf("Wait for either context cancellation or copying to finish \n"))
		// Wait for either context cancellation or copying to finish
		go func() {
			// Wait for both copies to finish or context to be cancelled
			select {
			case <-ctx.Done():
				resp.Close()
			case <-doneChan:
				// One copy finished, close the connection
				//fmt.Println("Finalizado")
				resp.Close()
			}
		}()

		// Wait for both goroutines to finish
		wg.Wait()
		return models.ShellFetchMsg{
			Finished: true,
		}
	}
}
