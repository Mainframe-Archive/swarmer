package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/MainframeHQ/swarmer/admin"
	"github.com/MainframeHQ/swarmer/util"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/pkg/stdcopy"
	"golang.org/x/net/context"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	log "github.com/camronlevanger/logrus"
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
	adminClient  admin.IAdminClient
	lookup       util.ILookup
	parser       util.IConfigParser
}

// GetStartCommand returns a pointer to a new instance of this implementation of IStartCommand.
func GetStartCommand(
	c models.Config,
	d *client.Client,
	a admin.IAdminClient,
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

	err := os.Chdir(path)

	if s.config.Repo == "" && s.config.Checkout == "" && s.config.Config == "" {
		s.config, err = s.parser.ParseYamlConfig("../swarmer.yml")
		if err != nil {
			log.Fatal(err)
			return err
		}
	}

	if s.config.Config != "" {
		s.config, err = s.parser.ParseYamlConfig("../" + s.config.Config)
		if err != nil {
			log.Fatal(err)
			return err
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
	cmd.Env = append(cmd.Env, "ENS="+s.config.ENS)
	cmd.Env = append(cmd.Env, "GETH="+strconv.FormatBool(s.config.Geth))

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Error("std io pipes broken")
	}

	scanner := bufio.NewScanner(stdout)

	go func() {
		if s.config.DockerLog == "" {
			s.config.DockerLog = "docker_log"
		}
		if s.config.SwarmLog == "" {
			s.config.DockerLog = "swarm_log"
		}
		f, err := os.Create(s.config.DockerLog)
		if err != nil {
			panic(err)
		}
		f, err = os.OpenFile(s.config.DockerLog, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}

		defer f.Close()
		for scanner.Scan() {
			line := scanner.Text()
			if _, err = f.WriteString(line + "\n"); err != nil {
				panic(err)
			}
			if strings.Contains(line, "started") {
				fmt.Println(line)
				break
			}
		}
	}()

	err = cmd.Start()
	if err != nil {
		log.Error(err.Error())
		return err
	}

	err = cmd.Wait()
	if err != nil {
		log.Error(err.Error())
	}

	cmd.Process.Release()

	var options types.ContainerListOptions

	options.All = true
	options.Filters = filters.NewArgs()
	options.Filters.Add("status", "running")
	options.Filters.Add("label", "org.mfhq.domain=swarm")

	containers, err := s.dockerClient.ContainerList(context.Background(), options)
	if err != nil {
		panic(err)
	}

	logsOptions := types.ContainerLogsOptions{
		ShowStderr: true,
		ShowStdout: true,
		Follow:     true,
	}

	var containerInfo types.ContainerJSON
	var info models.ContainerInfo
	var data []types.ContainerJSON
	for _, container := range containers {
		containerInfo, err = s.dockerClient.ContainerInspect(context.Background(), container.ID)
		if err != nil {
			log.Error("Unable to get container data: " + err.Error())
		}
		data = append(data, containerInfo)
		stream, err := s.dockerClient.ContainerLogs(context.Background(), container.ID, logsOptions)
		if err != nil {
			return err
		}

		swarmScanner := bufio.NewScanner(stream)

		f, err := os.Create(s.config.SwarmLog)
		if err != nil {
			panic(err)
		}
		f, err = os.OpenFile(s.config.SwarmLog, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}

		defer f.Close()
		for swarmScanner.Scan() {
			line := swarmScanner.Text()
			if _, err = f.WriteString(line + "\n"); err != nil {
				panic(err)
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

	var peerResult bool

	//dockerHost, err := s.lookup.GetIP("host.docker.internal")
	//if err != nil {
	//	return err
	//}

	// peering
	if len(nodeResults) > 1 {
		for _, nodeResult := range nodeResults {
			conn, err := s.adminClient.GetConnection("http://localhost:" + nodeResult.GatewayPort)
			if err != nil {
				return err
			}

			splitEnode := strings.Split(nodeResult.Enode, "@")
			enode := splitEnode[0] + "192.168.65.1:" + nodeResult.AdminPort

			err = conn.Call(&peerResult, "admin_addPeer", enode)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

	jsonData, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		log.Error(err.Error())
	}

	fmt.Println(string(jsonData))

	return nil
}
