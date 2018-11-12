package models

// Config defines the values needed by the application at runtime.
type Config struct {
	LocalSrc  string `json:"local-src" yaml:"local-src"`
	Repo      string `json:"repo" yaml:"repo"`
	Checkout  string `json:"checkout" yaml:"checkout"`
	Nodes     int    `json:"nodes" yaml:"nodes"`
	ENS       string `json:"ens-api" yaml:"ens-api"`
	LogLevel  string `json:"loglevel" yaml:"loglevel"`
	Geth      bool   `json:"geth" yaml:"geth"`
	Config    string `json:"config" yaml:"config"`
	Path      string `json:"path" yaml:"path"`
	DockerLog string `json:"docker_log" yaml:"docker_log"`
	SwarmLog  string `json:"swarm_log" yaml:"swarm_log"`
	Follow    bool   `json:"follow" yaml:"follow"`
}
