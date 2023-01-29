package database

import (
	"testing"
)

func TestInMemory_SetVaults(t *testing.T) {
	testSetVaults(t, true)
}
