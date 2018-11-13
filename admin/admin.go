package admin

import (
	"github.com/ethereum/go-ethereum/rpc"
)

// IClient is the interface for interacting with the Geth admin API.
type IClient interface {
	GetConnection(address string) (*rpc.Client, error)
}

// Client is the struct for this implementation of IClient.
type Client struct {
}

// GetClient returns a pointer to an instance of this implementation of IClient.
func GetClient() *Client {
	var s = Client{}

	return &s
}

// GetConnection returns a connection to the given Geth instance.
func (a *Client) GetConnection(address string) (*rpc.Client, error) {

	client, err := rpc.DialHTTP(address)
	if err != nil {
		return client, err
	}

	return client, nil
}
