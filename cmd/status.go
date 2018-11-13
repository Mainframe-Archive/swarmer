package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/MainframeHQ/swarmer/admin"
	"github.com/MainframeHQ/swarmer/models"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
	"gopkg.in/urfave/cli.v1"
)

// IStatusCommand is the interface to implement for the stop command.
type IStatusCommand interface {
	GetStatusCommand(c models.Config, d *client.Client, a admin.IClient) *StatusCommand
}

// StatusCommand is the struct for this implementation of IStatusCommand.
type StatusCommand struct {
	config       models.Config
	dockerClient *client.Client
	adminClient  admin.IClient
}

// GetStatusCommand returns a pointer to a new instance of this implementation of IStatusCommand.
func GetStatusCommand(c models.Config, d *client.Client, a admin.IClient) *StatusCommand {
	var s = StatusCommand{
		config:       c,
		dockerClient: d,
		adminClient:  a,
	}

	return &s
}

// Status shows the nodeInfo in JSON format of currently running nodes.
func (s *StatusCommand) Status(c *cli.Context) error {

	var options types.ContainerListOptions

	options.All = true
	options.Filters = filters.NewArgs()
	options.Filters.Add("status", "running")
	options.Filters.Add("label", "org.mfhq.domain=swarm")

	containers, err := s.dockerClient.ContainerList(context.Background(), options)
	if err != nil {
		panic(err)
	}

	var containerInfo types.ContainerJSON
	var info models.ContainerInfo
	var data []types.ContainerJSON
	for _, container := range containers {
		containerInfo, err = s.dockerClient.ContainerInspect(context.Background(), container.ID)
		if err != nil {
			return err
		}
		data = append(data, containerInfo)
	}

	info.Containers = data
	var nodeAdminPorts []string
	var nodeGatewayPorts []string
	var nodeResults []models.NodeInfo

	// get admin_nodeInfo data
	for i := range data {
		nodePort := info.Containers[i].NetworkSettings.Ports["30303/tcp"][0].HostPort
		nodeAdminPorts = append(nodeAdminPorts, nodePort)
		gatewayPort := info.Containers[i].NetworkSettings.Ports["8545/tcp"][0].HostPort
		nodeGatewayPorts = append(nodeGatewayPorts, gatewayPort)
		conn, err := s.adminClient.GetConnection("http://localhost:" + gatewayPort)
		if err != nil {
			return err
		}

		var nodeInfoResult models.NodeInfo
		var args interface{}

		err = conn.Call(&nodeInfoResult, "admin_nodeInfo", args)
		if err != nil {
			fmt.Println(err)
		} else {
			nodeInfoResult.AdminPort = nodePort
			nodeInfoResult.GatewayPort = gatewayPort
			nodeResults = append(nodeResults, nodeInfoResult)
		}

		conn.Close()
	}

	if len(nodeResults) > 0 {
		jsonData, err := json.MarshalIndent(nodeResults, "", "  ")
		if err != nil {
			return err
		}

		fmt.Println(string(jsonData))
		fmt.Println("There are " + strconv.Itoa(len(nodeResults)) + " active Swarm nodes.")
	} else {
		fmt.Println("There are no Swarm nodes running.")
	}

	return nil

}
