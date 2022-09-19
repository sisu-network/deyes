package eth

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDispatcher_Shuffle(t *testing.T) {
	urls := []string{
		"example1.com",
		"example2.com",
		"example3.com",
		"example4.com",
		"example5.com",
	}
	d := NewEhtDispatcher("ganache1", urls).(*EthDispatcher)

	d.healthy[1] = true
	d.healthy[2] = true
	d.healthy[4] = true

	rand.Seed(100)
	_, healthy, rpcs := d.shuffle()

	require.Equal(t, []string{"example5.com", "example1.com", "example2.com", "example4.com", "example3.com"}, rpcs)
	require.Equal(t, []bool{true, false, true, false, true}, healthy)
}
