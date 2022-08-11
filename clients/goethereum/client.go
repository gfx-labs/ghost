// supposedly this should get updated as the dependency does automaticaly
// if they do breaking changes tho i will be sad :(
package goethereum

import (
	"context"

	"gfx.cafe/open/ghost"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// mimics dial but returns a ghost.Client interface
func Dial(uri string) (ghost.Client, error) {
	cl, err := ethclient.Dial(uri)
	return cl, err
}

// same as dial, except with a context!
func DialContext(ctx context.Context, uri string) (ghost.Client, error) {
	cl, err := ethclient.Dial(uri)
	return cl, err
}

// mimics NewClient but returns a ghost.Client interface
func NewClient(c *rpc.Client) ghost.Client {
	return ethclient.NewClient(c)
}
