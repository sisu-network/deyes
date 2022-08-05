package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBlockTimeTracker(t *testing.T) {
	timeTracker := NewBlockTimeTracker(10000)

	timeTracker.HitBlock()
	require.Equal(t, 9500, timeTracker.GetSleepTime())
	timeTracker.HitBlock()
	require.Equal(t, 9025, timeTracker.GetSleepTime())
	timeTracker.HitBlock()
	require.Equal(t, 5415, timeTracker.GetSleepTime())
	require.Equal(t, 3, timeTracker.consecutiveHit)

	timeTracker.HitBlockWithMinorDelay()
	require.Equal(t, 5550, timeTracker.GetSleepTime())
	require.Equal(t, 0, timeTracker.consecutiveHit)
	timeTracker.MissBlock()
	require.Equal(t, 6105, timeTracker.GetSleepTime())
	require.Equal(t, 0, timeTracker.consecutiveHit)
}
