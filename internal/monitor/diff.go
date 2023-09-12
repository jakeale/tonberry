package monitor

import (
	"sort"

	"github.com/jakeale/tonberry/internal/lodestone"
)

// StatusChange describes a single field that changed for a world between two scrapes.
type StatusChange struct {
	World    string
	Field    string // "status" | "characterCreationStatus"
	OldValue string
	NewValue string
}

const fieldStatus = "status"
const fieldCharacterCreationStatus = "characterCreationStatus"

// diffServers compares two Servers snapshots and returns the status changes between them.
//
// Worlds that appear or disappear between snapshots are intentionally skipped - Lodestone
// adding or removing a world entirely is not a meaningful "status changed" event for
// subscribers, and would otherwise spam channels during unrelated Lodestone maintenance.
// Category changes are not reported for the same reason: they are not actionable enough
// to warrant a notification.
//
// The result is sorted by world name for deterministic output.
func diffServers(prev, next lodestone.Servers) []StatusChange {
	changes := make([]StatusChange, 0)

	for world, nextStatus := range next {
		prevStatus, existed := prev[world]
		if !existed {
			continue
		}

		if prevStatus.Status != nextStatus.Status {
			changes = append(changes, StatusChange{
				World:    world,
				Field:    fieldStatus,
				OldValue: prevStatus.Status,
				NewValue: nextStatus.Status,
			})
		}

		if prevStatus.CharacterCreationStatus != nextStatus.CharacterCreationStatus {
			changes = append(changes, StatusChange{
				World:    world,
				Field:    fieldCharacterCreationStatus,
				OldValue: prevStatus.CharacterCreationStatus,
				NewValue: nextStatus.CharacterCreationStatus,
			})
		}
	}

	sort.Slice(changes, func(i, j int) bool {
		if changes[i].World != changes[j].World {
			return changes[i].World < changes[j].World
		}
		return changes[i].Field < changes[j].Field
	})

	return changes
}
