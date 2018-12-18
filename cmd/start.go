package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/MainframeHQ/swarmer/admin"
	"github.com/MainframeHQ/swarmer/util"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/go-errors/errors"
	"golang.org/x/net/context"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	log "github.com/camronlevanger/logrus"
	"github.com/otiai10/copy"
	"gopkg.in/urfave/cli.v1"

	"github.com/MainframeHQ/swarmer/models"
)

// IStartCommand is the interface to implement for the start command.
type IStartCommand interface {
	Start(c *cli.Context) error
}

// StartCommand is the struct for this implementation of IStartCommand.
type StartCommand struct {
	config       models.Config
	dockerClient *client.Client
	adminClient  admin.IClient
	lookup       util.ILookup
	parser       util.IConfigParser
}

// GetStartCommand returns a pointer to a new instance of this implementation of IStartCommand.
func GetStartCommand(
	c models.Config,
	d *client.Client,
	a admin.IClient,
	l util.ILookup,
	p util.IConfigParser,
) *StartCommand {
	var s = StartCommand{
		config:       c,
		dockerClient: d,
		adminClient:  a,
		lookup:       l,
		parser:       p,
	}

	return &s
}

// Start is the command that starts the Swarm nodes.
func (s *StartCommand) Start(c *cli.Context) error {

	path := s.config.Path

	var err error

	if s.config.Repo == "" && s.config.Checkout == "" && s.config.Config == "" {
		s.config, err = s.parser.ParseYamlConfig("swarmer.yml")
		if err != nil {
			return errors.Errorf("Error parsing YAML config: %s", err.Error())
		}
	}

	if s.config.Config != "" {
		s.config, err = s.parser.ParseYamlConfig(s.config.Config)
		if err != nil {
			return errors.Errorf("Error parsing YAML config: %s", err.Error())
		}
	}

	err = os.Chdir(path)

	if s.config.Add != "" {
		err := copy.Copy(s.config.Add, "./addme")
		if err != nil {
			return errors.Errorf("Error copying directory from host to container: %s", err.Error())
		}
	}

	cmd := exec.Command("docker-compose", "up", "--build", "--force-recreate", "--detach")
	cmd.Dir = s.config.Path
	cmd.Env = os.Environ()
	cmd.Args = append(cmd.Args, "--scale")
	cmd.Args = append(cmd.Args, "swarm="+strconv.Itoa(s.config.Nodes))
	cmd.Env = append(cmd.Env, "REPO="+s.config.Repo)
	cmd.Env = append(cmd.Env, "CHECKOUT="+s.config.Checkout)
	cmd.Env = append(cmd.Env, "CONFIG="+s.config.Config)
	if s.config.ENS != "" {
		cmd.Env = append(cmd.Env, "ENS="+s.config.ENS)
	}
	cmd.Env = append(cmd.Env, "GETH="+strconv.FormatBool(s.config.Geth))

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Error("std io pipes broken")
	}

	scanner := bufio.NewScanner(stdout)

	go func() {
		if s.config.DockerLog == "" {
			s.config.DockerLog = "/var/log/docker_log"
		}
		if s.config.SwarmLog == "" {
			s.config.DockerLog = "/var/log/swarm_log"
		}
		f, err := os.Create(s.config.DockerLog)
		if err != nil {
			panic(errors.Wrap(err, 1))
		}
		f, err = os.OpenFile(s.config.DockerLog, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			panic(errors.Wrap(err, 1))
		}

		defer f.Close()
		for scanner.Scan() {
			line := scanner.Text()
			if _, err = f.WriteString(line + "\n"); err != nil {
				panic(errors.Wrap(err, 1))
			}
			if strings.Contains(line, "started") {
				break
			}
		}
	}()

	err = cmd.Start()
	if err != nil {
		return errors.Errorf("Error starting command: %s", err.Error())
	}

	err = cmd.Wait()
	if err != nil {
		return errors.Errorf("Error waiting for command: %s", err.Error())
	}

	cmd.Process.Release()

	var options types.ContainerListOptions

	options.All = true
	options.Filters = filters.NewArgs()
	options.Filters.Add("status", "running")
	options.Filters.Add("label", "org.mfhq.domain=swarm")

	containers, err := s.dockerClient.ContainerList(context.Background(), options)
	if err != nil {
		return errors.Wrap(err, 1)
	}

	logsOptions := types.ContainerLogsOptions{
		ShowStderr: true,
		ShowStdout: true,
		Follow:     true,
	}

	var containerInfo types.ContainerJSON
	var containerNames [][]string
	var splitNames []string
	var info models.ContainerInfo
	var data []types.ContainerJSON
	for _, container := range containers {
		containerInfo, err = s.dockerClient.ContainerInspect(context.Background(), container.ID)
		if err != nil {
			return errors.Errorf("Error inspecting container %s: %s", container.ID, err.Error())
		}

		data = append(data, containerInfo)

		for _, name := range container.Names {
			tokens := strings.Split(name, "/")
			name = tokens[1]
			splitNames = append(splitNames, name)
		}

		containerNames = append(containerNames, splitNames)

		stream, err := s.dockerClient.ContainerLogs(context.Background(), container.ID, logsOptions)
		if err != nil {
			return errors.Errorf("Error getting container log stream: %s", err.Error())
		}

		swarmScanner := bufio.NewScanner(stream)

		f, err := os.Create(s.config.SwarmLog)
		if err != nil {
			return errors.Errorf("Error creating swarm log file on host: %s", err.Error())
		}
		f, err = os.OpenFile(s.config.SwarmLog, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return errors.Errorf("Error opening swarm logs for writing on host: %s", err.Error())
		}

		defer f.Close()
		for swarmScanner.Scan() {
			line := swarmScanner.Text()
			if _, err = f.WriteString(line + "\n"); err != nil {
				return errors.Errorf("Error writing swarm logs to host machine: %s", err.Error())
			}
			if strings.Contains(line, "WebSocket endpoint opened") {
				break
			}
		}

		if s.config.Follow {
			stdcopy.StdCopy(os.Stdout, os.Stderr, stream)
			stream.Close()
		}
	}

	info.Containers = data
	var nodeCommPorts []string
	var nodeWebsocketPorts []string
	var nodeGatewayPorts []string
	var nodeAdminPorts []string
	var nodeResults []models.NodeInfo

	// get admin_nodeInfo data
	for i := range data {
		adminPort := info.Containers[i].NetworkSettings.Ports["8545/tcp"][0].HostPort
		nodeAdminPorts = append(nodeAdminPorts, adminPort)
		commPort := info.Containers[i].NetworkSettings.Ports["30303/tcp"][0].HostPort
		nodeCommPorts = append(nodeCommPorts, commPort)
		websocketPort := info.Containers[i].NetworkSettings.Ports["8546/tcp"][0].HostPort
		nodeWebsocketPorts = append(nodeWebsocketPorts, websocketPort)
		gatewayPort := info.Containers[i].NetworkSettings.Ports["8500/tcp"][0].HostPort
		nodeGatewayPorts = append(nodeGatewayPorts, gatewayPort)

		conn, err := s.adminClient.GetConnection("http://localhost:" + adminPort)
		if err != nil {
			return errors.Errorf("Error instantiating Geth admin connection over RPC: %s", err.Error())
		}

		var nodeInfoResult models.NodeInfo
		var args interface{}

		err = conn.Call(&nodeInfoResult, "admin_nodeInfo", args)
		if err != nil {
			return errors.Errorf("Unable to call nodeInfo function on geth node: %s", err.Error())
		}

		nodeInfoResult.ContainerID = info.Containers[i].ID
		nodeInfoResult.CommPort = commPort
		nodeInfoResult.GatewayPort = gatewayPort
		nodeInfoResult.WebsocketPort = websocketPort
		nodeInfoResult.AdminPort = adminPort
		nodeInfoResult.ContainerNames = containerNames[i]
		nodeInfoResult.IPAddress = info.Containers[i].NetworkSettings.Networks["docker_swarm_network"].IPAddress
		nodeResults = append(nodeResults, nodeInfoResult)

		conn.Close()
	}

	var peerResult bool

	// just a safety buffer to make sure the nodeInfo is setup (I know this sucks, I'll get to it...)
	time.Sleep(4 * time.Second)

	// peering
	if len(nodeResults) > 1 {
		for i, nodeResult := range nodeResults {
			conn, err := s.adminClient.GetConnection("http://localhost:" + nodeResult.AdminPort)
			if err != nil {
				return errors.Errorf("Unable to connect to geth on port %s", nodeResult.AdminPort)
			}

			var nextNode int
			if i < (len(nodeResults) - 1) {
				nextNode = i + 1
			} else {
				nextNode = 0
			}

			splitEnode := strings.Split(nodeResults[nextNode].Enode, "@")
			enode := splitEnode[0] + "@" + nodeResults[nextNode].IPAddress + ":" + nodeResults[nextNode].CommPort

			err = conn.Call(&peerResult, "admin_addPeer", enode)
			if err != nil {
				return errors.Errorf("Unable to call addPeer function on geth node %s with enode %s - %s", nodeResult.ContainerNames[0], enode, err.Error())
			}
		}
	}

	jsonData, err := json.MarshalIndent(nodeResults, "", "  ")
	if err != nil {
		return errors.Wrap(err, 1)
	}

	fmt.Println(string(jsonData))

	return nil
}
