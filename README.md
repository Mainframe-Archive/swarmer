# Swarmer

Swarmer treats Swarm similar to a dev dependency. Add a `swarmer.yml` file to your directory and then simply run `swarmer start`.

Swarmer automatically spins up the number of Swarm nodes you specify, peers the nodes, and then returns the details of each node in JSON as output to be used by build systems, etc..

## Installation

`go get github.com/MainframeHQ/swarmer` or clone and build manually.

Other means of installation may be provided in the future for those without a Golang environment already setup.

## Usage

Swarmer can be invoked either with a Yaml file describing the Swarm nodes, or by using command line flags.

To get started quickly, see the `swarmer.yml` file in this repo as an example to get started.

For command line usage have a look at the output from `swarmer help`.

### Commands and Flags

COMMANDS

 * start, s   Start the Swarm cluster
 * stop, t    Stop the Swarm cluster
 * status, a  Get a list of running nodes
 * help, h    Shows a list of commands or help for one command

GLOBAL OPTIONS
   
   * --nodes value, -n value       how many swarm nodes to start (default: 1) [$DEVCLUSTER_NODES]
   * --config value, -C value      load a YAML or TOML config file rather than supplying args and flags [$DEVCLUSTER_CONFIG]
   * --repo value, -r value        URL to Git repository containing Swarm source to be built [$DEVCLUSTER_REPO]
   * --srcdir value, -d value      build source from given directory rather than from Git repo [$DEVCLUSTER_SRC]
   * --checkout value, -c value    branch, tag, or hash to checkout from the Git repo [$DEVCLUSTER_CHECKOUT]
   * --ens-api value, -e value     this value is passed directly to Swarm ens-api flag [$DEVCLUSTER_ENS]
   * --geth, -g                    run Geth as well as swarm [$DEVCLUSTER_GETH]
   * --docker_log value, -b value  local logfile for Docker build logs (default: "docker_log") [$DEVCLUSTER_DOCKER_LOG]
   * --swarm_log value, -s value   local logfile for Swarm logs (default: "swarm_log") [$DEVCLUSTER_SWARM_LOG]
   * --follow, -f                  remain attached and display Swarm logs [$DEVCLUSTER_FOLLOW]
   * --help, -h                    show help
   * --version, -v                 print the version
   
#### Example

`swarmer --nodes 3 --repo https://github.com/ethereum/go-ethereum --checkout v1.8.17 --ens-api https://mainnet.infura.io/v3/YOUR-INFURA-KEY --geth start`

### Using swarmer.yml

For convenience, Swarmer automatically looks for a `swarmer.yml` in the current directory. If your project requires Swarm, you will find it rather convenient to simply include a `swarmer.yml` file in your source repository. This way when a developer, or a build system checks out your repo to work with it, all that has to be done is to run `swarmer`. 

Swarmer spins up the required number of nodes and peers them together. Additionally it gives you confidence that developers are working with the same version of Swarm, as you can pin Swarm to a specific version in the Yaml file.

Currently, `swarmer.yml` is flat, and accepts the same arguments as supported by command line flags listed above.