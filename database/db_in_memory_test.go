package database

import (
	"testing"
)

func TestInMemory_SetVaults(t *testing.T) {
	testSetVaults(t, true)
}

func TestInMemory_TokenPrice(t *testing.T) {
	testTokenPrice(t, true)
}
