package abi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCallParser(t *testing.T) {
	sig, params := Call("add(uint256 1, uint256 2)").Decode()
	assert.EqualValues(t, "add(uint256,uint256)", sig)
	assert.EqualValues(t, []string{"1", "2"}, params)
}

func TestCallParserString(t *testing.T) {
	sig, params := Call(`add(uint256 1, string "wowza")`).Decode()
	assert.EqualValues(t, "add(uint256,string)", sig)
	assert.EqualValues(t, []string{"1", `"wowza"`}, params)
}
