package main

import (
	"os"
	"runtime"
	"time"

	"github.com/MainframeHQ/swarmer/admin"
	"github.com/MainframeHQ/swarmer/util"
	"github.com/docker/docker/client"

	log "github.com/camronlevanger/logrus"
	"github.com/go-errors/errors"
	"gopkg.in/urfave/cli.v1"

	"github.com/MainframeHQ/swarmer/cmd"
	models "github.com/MainframeHQ/swarmer/models"
)

// DATAPATH is a best effort at finding the src directory of this project in order
// to make use of the Dockerfile and docker-compose.yml files etc..
var DATAPATH = os.Getenv("GOPATH") + "/src/github.com/MainframeHQ/swarmer/"

// APPNAME is aptly named.
const APPNAME = "swarmer"

func init() {

	log.SetFormatter(
		&log.TextFormatter{
			DisableColors: true,
			FullTimestamp: false,
		},
	)

	if runtime.GOOS == "darwin" {
		log.Debug("Mac OS detected")
	} else if runtime.GOOS == "linux" {
		log.Debug("Linux detected")
	} else {
		log.Fatal(APPNAME+" does not support the %s operating system at this time.", runtime.GOOS)
	}

}

func main() {
	path := DATAPATH + "docker"
	config := models.Config{}
	config.Path = path

	var start *cmd.StartCommand
	var stop *cmd.StopCommand
	var status *cmd.StatusCommand

	dockerClient, err := client.NewClientWithOpts(client.WithVersion("1.38"))
	if err != nil {
		panic("Must have Docker compatible with API 1.38: " + err.Error())
	}

	adminClient := admin.GetClient()
	lookup := util.GetLookup()
	parser := util.GetConfigParser()

	app := cli.NewApp()
	app.Name = APPNAME
	app.Version = "0.1"
	app.Usage = "Run a local, containerized Swarm cluster comprised of (N) peered nodes."
	app.Compiled = time.Now()
	app.Authors = []cli.Author{
		{
			Name:  "Camron G. Levanger",
			Email: "camron@mainframe.com",
		},
	}
	app.Copyright = "(c) 2018 Mainframe"
	app.EnableBashCompletion = true

	app.Commands = []cli.Command{
		{
			Name:    "start",
			Aliases: []string{"s"},
			Usage:   "Start the Swarm cluster",
			Action: func(c *cli.Context) error {
				start = cmd.GetStartCommand(config, dockerClient, adminClient, lookup, parser)
				err := start.Start(c)

				return err
			},
		},
		{
			Name:    "stop",
			Aliases: []string{"t"},
			Usage:   "Stop the Swarm cluster",
			Action: func(c *cli.Context) error {
				stop = cmd.GetStopCommand(config, dockerClient)
				err := stop.Stop(c)

				return err
			},
		},
		{
			Name:    "status",
			Aliases: []string{"a"},
			Usage:   "Get a list of running nodes",
			Action: func(c *cli.Context) error {
				status = cmd.GetStatusCommand(config, dockerClient, adminClient)
				err := status.Status(c)

				return err
			},
		},
	}
	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:        "nodes, n",
			Value:       1,
			Usage:       "how many swarm nodes to start",
			EnvVar:      "DEVCLUSTER_NODES",
			Destination: &config.Nodes,
		},
		cli.StringFlag{
			Name:        "config, C",
			Value:       "",
			Usage:       "load a YAML or TOML config file rather than supplying args and flags",
			EnvVar:      "DEVCLUSTER_CONFIG",
			Destination: &config.Config,
		},
		cli.StringFlag{
			Name:        "repo, r",
			Value:       "",
			Usage:       "URL to Git repository containing Swarm source to be built",
			EnvVar:      "DEVCLUSTER_REPO",
			Destination: &config.Repo,
		},
		cli.StringFlag{
			Name:        "srcdir, d",
			Value:       "",
			Usage:       "build source from given directory rather than from Git repo",
			EnvVar:      "DEVCLUSTER_SRC",
			Destination: &config.LocalSrc,
		},
		cli.StringFlag{
			Name:        "checkout, c",
			Value:       "",
			Usage:       "branch, tag, or hash to checkout from the Git repo",
			EnvVar:      "DEVCLUSTER_CHECKOUT",
			Destination: &config.Checkout,
		},
		cli.StringFlag{
			Name:        "ens-api, e",
			Value:       "",
			Usage:       "this value is passed directly to Swarm ens-api flag",
			EnvVar:      "DEVCLUSTER_ENS",
			Destination: &config.ENS,
		},
		cli.BoolFlag{
			Name:        "geth, g",
			Usage:       "run Geth as well as swarm",
			EnvVar:      "DEVCLUSTER_GETH",
			Destination: &config.Geth,
		},
		cli.StringFlag{
			Name:        "docker_log, b",
			Value:       "docker_log",
			Usage:       "local logfile for Docker build logs",
			EnvVar:      "DEVCLUSTER_DOCKER_LOG",
			Destination: &config.DockerLog,
		},
		cli.StringFlag{
			Name:        "swarm_log, s",
			Value:       "swarm_log",
			Usage:       "local logfile for Swarm logs",
			EnvVar:      "DEVCLUSTER_SWARM_LOG",
			Destination: &config.SwarmLog,
		},
		cli.StringFlag{
			Name:        "add, a",
			Value:       "",
			Usage:       "add directory to the swarmer containers",
			EnvVar:      "DEVCLUSTER_ADD",
			Destination: &config.Add,
		},
		cli.BoolFlag{
			Name:        "follow, f",
			Usage:       "remain attached and display Swarm logs",
			EnvVar:      "DEVCLUSTER_FOLLOW",
			Destination: &config.Follow,
		},
	}

	app.Action = func(c *cli.Context) error {
		// this uses the start command as default if no command given
		start = cmd.GetStartCommand(config, dockerClient, adminClient, lookup, parser)
		err := start.Start(c)

		return errors.Wrap(err, 1)
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatalf("Error starting swarm nodes: %+v", err)
	}
}
