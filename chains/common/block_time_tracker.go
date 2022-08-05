package common

import (
	"sync"
)

const (
	MaxTrackSize = 5
)

type BlockTimeTracker struct {
	lock           *sync.RWMutex
	currentValue   int
	consecutiveHit int
}

func NewBlockTimeTracker(blockTime int) *BlockTimeTracker {
	return &BlockTimeTracker{
		lock:         &sync.RWMutex{},
		currentValue: blockTime,
	}
}

// HitBlock is called when a new block is retrieved within the the last block time.
func (t *BlockTimeTracker) HitBlock() {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.consecutiveHit++
	if t.consecutiveHit >= 3 {
		t.currentValue = t.currentValue * 6 / 10 // Drop block time by 40%
	} else {
		t.currentValue = t.currentValue * 950 / 1000 // Drop block time by 5%
	}

	if t.currentValue < 500 {
		t.currentValue = 500
	}
}

// HitBlockWithMinorDelay is called when new block time is slightly higher than the last one.
func (t *BlockTimeTracker) HitBlockWithMinorDelay() {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.currentValue = t.currentValue * 1025 / 1000 // Increase block time by 2.5%
	t.consecutiveHit = 0
}

// MissBlock is called when a block is missed.
func (t *BlockTimeTracker) MissBlock() {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.currentValue = t.currentValue * 11 / 10 // Increase block time by 10%
	t.consecutiveHit = 0
}

func (t *BlockTimeTracker) GetSleepTime() int {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return t.currentValue
}
