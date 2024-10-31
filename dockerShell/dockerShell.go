package dockerShell

import (
	"context"
	"docker-tui-go/models"
	"fmt"
	"io"
	"os"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/docker/docker/api/types/container"
	dockerClient "github.com/docker/docker/client"
	"golang.org/x/term"
)

func Dockershell(cli *dockerClient.Client, selectedContainer models.Items) tea.Cmd {
	return func() tea.Msg {
		// Create exec configuration
		execConfig := container.ExecOptions{
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Tty:          true,
			Cmd:          []string{"/bin/bash"},
		}

		// Create exec instance
		execID, err := cli.ContainerExecCreate(context.Background(), selectedContainer.Id, execConfig)
		if err != nil {
			return models.ShellFetchMsg{
				Error:    fmt.Sprintf("Error creating exec: %v", err),
				Finished: true,
			}
		}

		// Attach to the exec instance
		resp, err := cli.ContainerExecAttach(context.Background(), execID.ID, container.ExecAttachOptions{
			Tty: true,
		})
		if err != nil {
			return models.ShellFetchMsg{
				Error:    fmt.Sprintf("Error attaching to exec: %v", err),
				Finished: true,
			}
		}
		defer resp.Close()

		// Save terminal state
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			return models.ShellFetchMsg{
				Error:    fmt.Sprintf("Error setting raw terminal: %v", err),
				Finished: true,
			}
		}
		defer term.Restore(int(os.Stdin.Fd()), oldState)

		// Create context with cancellation
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Create WaitGroup for goroutines
		var wg sync.WaitGroup

		// Create error channel
		errChan := make(chan error, 2)

		// Start copying from container to stdout
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := io.Copy(os.Stdout, resp.Reader)
			if err != nil {
				errChan <- err
			}
		}()

		// Start copying from stdin to container
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := io.Copy(resp.Conn, os.Stdin)
			if err != nil {
				errChan <- err
			}
			// Signal cancellation when stdin is closed (e.g., after 'exit' command)
			cancel()
		}()

		// Monitor exec status
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					inspect, err := cli.ContainerExecInspect(context.Background(), execID.ID)
					if err != nil {
						errChan <- err
						return
					}
					if !inspect.Running {
						cancel()
						return
					}
				}
			}
		}()

		// Wait for completion
		go func() {
			wg.Wait()
			close(errChan)
		}()

		// Wait for either error or completion
		select {
		case err := <-errChan:
			if err != nil && err != io.EOF {
				return models.ShellFetchMsg{
					Error:    fmt.Sprintf("Shell error: %v", err),
					Finished: true,
				}
			}
		case <-ctx.Done():
			// Normal termination
		}

		// Ensure terminal is restored and return success
		term.Restore(int(os.Stdin.Fd()), oldState)
		return models.ShellFetchMsg{
			Finished: true,
		}
	}
}
