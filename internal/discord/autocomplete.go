package discord

import (
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jakeale/tonberry/internal/lodestone"
)

const maxAutocompleteChoices = 25 // Discord's hard limit on choices per response

// worldAutocompleteChoices returns world names from the current snapshot whose name
// contains the user's in-progress input (case-insensitive), sorted and capped at
// Discord's per-response choice limit.
func worldAutocompleteChoices(servers lodestone.Servers, input string) []*discordgo.ApplicationCommandOptionChoice {
	input = strings.ToLower(input)

	matches := make([]string, 0, len(servers))
	for world := range servers {
		if strings.Contains(strings.ToLower(world), input) {
			matches = append(matches, world)
		}
	}
	sort.Strings(matches)

	if len(matches) > maxAutocompleteChoices {
		matches = matches[:maxAutocompleteChoices]
	}

	choices := make([]*discordgo.ApplicationCommandOptionChoice, len(matches))
	for i, world := range matches {
		choices[i] = &discordgo.ApplicationCommandOptionChoice{Name: world, Value: world}
	}

	return choices
}

// isKnownWorld reports whether world is present in the current snapshot, used to
// validate submitted (non-autocomplete) values before hitting the store or Lodestone.
func isKnownWorld(servers lodestone.Servers, world string) bool {
	_, found := servers[world]
	return found
}
