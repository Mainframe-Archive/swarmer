package admin

import (
	"github.com/ethereum/go-ethereum/rpc"
)

// IAdminClient is the interface for interacting with the Geth admin API.
type IAdminClient interface {
	GetConnection(address string) (*rpc.Client, error)
}

// AdminClient is the struct for this implementation of IAdminClient.
type AdminClient struct {
}

func GetAdminClient() *AdminClient {
	var s = AdminClient{}

	return &s
}

// GetConnection returns a connection to the given Geth instance.
func (a *AdminClient) GetConnection(address string) (*rpc.Client, error) {

	client, err := rpc.DialHTTP(address)
	if err != nil {
		return client, err
	}

	return client, nil
}
