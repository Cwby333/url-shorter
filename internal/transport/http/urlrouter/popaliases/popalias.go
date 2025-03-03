package popaliases

import (
	"sync"
	"time"
)

type PopAlias struct {
	mu       *sync.Mutex
	storage  map[string]int
	TimeSend time.Duration
}

func New(timeSend time.Duration) PopAlias {
	return PopAlias{
		mu:       &sync.Mutex{},
		storage:  make(map[string]int),
		TimeSend: timeSend,
	}
}

func (p PopAlias) Inc(alias string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	_, ok := p.storage[alias]

	if !ok {
		p.storage[alias] = 1
		return
	}

	p.storage[alias]++
	return
}

func (p PopAlias) GetMostPopularAlias() (string, int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	max := 0
	out := ""
	for alias, count := range p.storage {
		if count > max {
			out = alias
			max = count
		}
	}

	if max == 0 {
		return "", 0
	}

	clear(p.storage)
	return out, max
}
