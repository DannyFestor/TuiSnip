package savegate_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/savegate"
)

const concurrentSaves = 16

var errDiskFull = errors.New("disk full")

func TestGate_Save(t *testing.T) {
	t.Parallel()

	t.Run("runs saves that arrive in the order they were started", func(t *testing.T) {
		t.Parallel()

		var gate savegate.Gate

		first, second := gate.Ticket(), gate.Ticket()
		saved := saveAll(t, &gate, first, second)

		assert.Equal(t, []savegate.Ticket{first, second}, saved)
	})

	t.Run("skips a save started before one that already ran", func(t *testing.T) {
		t.Parallel()

		var gate savegate.Gate

		first, second := gate.Ticket(), gate.Ticket()
		saved := saveAll(t, &gate, second, first)

		assert.Equal(t, []savegate.Ticket{second}, saved)
	})

	t.Run("reports the save's error", func(t *testing.T) {
		t.Parallel()

		var gate savegate.Gate

		err := gate.Save(gate.Ticket(), func() error { return errDiskFull })

		require.ErrorIs(t, err, errDiskFull)
		assert.ErrorContains(t, err, "savegate.Gate.Save: ")
	})

	t.Run("ends on the last started save when saves race", func(t *testing.T) {
		t.Parallel()

		var (
			gate  savegate.Gate
			last  savegate.Ticket
			group sync.WaitGroup
		)

		tickets := make([]savegate.Ticket, 0, concurrentSaves)
		for range concurrentSaves {
			tickets = append(tickets, gate.Ticket())
		}

		for _, ticket := range tickets {
			group.Go(func() {
				assert.NoError(t, gate.Save(ticket, func() error {
					last = ticket

					return nil
				}))
			})
		}

		group.Wait()

		assert.Equal(t, tickets[len(tickets)-1], last)
	})
}

func saveAll(t *testing.T, gate *savegate.Gate, tickets ...savegate.Ticket) []savegate.Ticket {
	t.Helper()

	var saved []savegate.Ticket

	for _, ticket := range tickets {
		require.NoError(t, gate.Save(ticket, func() error {
			saved = append(saved, ticket)

			return nil
		}))
	}

	return saved
}
