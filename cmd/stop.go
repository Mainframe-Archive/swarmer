package cmd

import (
	"fmt"
	"golang.org/x/net/context"

	"github.com/MainframeHQ/swarmer/models"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"gopkg.in/urfave/cli.v1"
)

// IStopCommand is the interface to implement for the stop command.
type IStopCommand interface {
	Stop(c *cli.Context) error
}

// StopCommand is the struct for this implementation of IStopCommand.
type StopCommand struct {
	config       models.Config
	dockerClient *client.Client
}

// GetStopCommand returns a pointer to a new instance of this implementation of IStopCommand.
func GetStopCommand(c models.Config, d *client.Client) *StopCommand {
	var s = StopCommand{
		config:       c,
		dockerClient: d,
	}

	return &s
}

// Stop is the command that stops the Swarm nodes.
func (s *StopCommand) Stop(c *cli.Context) error {

	var options types.ContainerListOptions

	options.All = true
	options.Filters = filters.NewArgs()
	options.Filters.Add("status", "running")
	options.Filters.Add("label", "org.mfhq.domain=swarm")

	containers, err := s.dockerClient.ContainerList(context.Background(), options)
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		fmt.Printf("Stopping container %s...\n", container.ID)
		if err := s.dockerClient.ContainerStop(context.Background(), container.ID, nil); err != nil {
			panic(err)
		}
		fmt.Printf("Container %s stopped successfully.\n", container.ID)
	}

	return nil

}
