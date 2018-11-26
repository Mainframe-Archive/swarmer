package models

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
)

// ContainerInfo is the data structure for the container information we output in JSON format
// to be used from build systems.
type ContainerInfo struct {
	Containers []types.ContainerJSON `json:"containers" yaml:"containers"`
}

// DockerContainers is intended to be a lighter weight struct for types.ContainerJSON.
type DockerContainers struct {
	ID         string                `json:"id" yaml:"id"`
	Created    string                `json:"created" yaml:"created"`
	Args       []string              `json:"args" yaml:"args"`
	State      *types.ContainerState `json:"state" yaml:"state"`
	Image      string                `json:"image" yaml:"image"`
	HostsPath  string                `json:"hosts_path" yaml:"hosts_path"`
	LogPath    string                `json:"log_path" yaml:"log_path"`
	Node       *types.ContainerNode  `json:"node" yaml:"node"`
	Name       string                `json:"name" yaml:"name"`
	HostConfig *container.HostConfig `json:"host_config" yaml:"host_config"`
}

// NodeInfo holds data we need for peering.
type NodeInfo struct {
	CommPort      string `json:"comm_port" yaml:"comm_port"`
	GatewayPort   string `json:"gateway_port" yaml:"gateway_port"`
	WebsocketPort string `json:"websocket_port" yaml:"websocket_port"`
	Enode         string `json:"enode" yaml:"enode"`
	Enr           string `json:"enr" yaml:"enr"`
	ID            string `json:"id" yaml:"id"`
	Name          string `json:"name" yaml:"name"`
	ContainerID   string `json:"container_id" yaml:"container_id"`
}

// Ports maps go-ethereum ports section of NodeInfo.
type Ports struct {
	Discovery int
	Listener  int
}
