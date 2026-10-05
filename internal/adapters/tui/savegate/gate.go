package savegate

import (
	"fmt"
	"sync"
)

type Ticket uint64

type Gate struct {
	mutex  sync.Mutex
	issued Ticket
	latest Ticket
}

func (g *Gate) Ticket() Ticket {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	g.issued++

	return g.issued
}

func (g *Gate) Save(ticket Ticket, save func() error) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if ticket < g.latest {
		return nil
	}

	g.latest = ticket

	err := save()
	if err != nil {
		return fmt.Errorf("savegate.Gate.Save: %w", err)
	}

	return nil
}
