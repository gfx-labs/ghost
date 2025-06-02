package ghost_test

import (
	"gfx.cafe/open/ghost"
	"github.com/ethereum/go-ethereum/ethclient"
)

var _ ghost.Client = (*ethclient.Client)(nil)
