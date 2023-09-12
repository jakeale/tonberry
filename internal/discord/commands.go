package discord

import "github.com/bwmarrin/discordgo"

const (
	commandSubscribe     = "subscribe"
	commandUnsubscribe   = "unsubscribe"
	commandSubscriptions = "subscriptions"
	commandStatus        = "status"
	commandCharacter     = "character"
	commandFreeCompany   = "freecompany"

	optionWorld = "world"
	optionName  = "name"
)

// commandDefinitions is bulk-registered on bot startup via ApplicationCommandBulkOverwrite,
// which both creates new commands and removes any stale ones from a previous version.
var commandDefinitions = []*discordgo.ApplicationCommand{
	{
		Name:        commandSubscribe,
		Description: "Subscribe this channel to a world's status changes",
		Options: []*discordgo.ApplicationCommandOption{
			worldOption("The world to subscribe to", true),
		},
	},
	{
		Name:        commandUnsubscribe,
		Description: "Unsubscribe this channel from a world's status changes",
		Options: []*discordgo.ApplicationCommandOption{
			worldOption("The world to unsubscribe from", true),
		},
	},
	{
		Name:        commandSubscriptions,
		Description: "List this server's world status subscriptions",
	},
	{
		Name:        commandStatus,
		Description: "Show the current status of a world, or a summary of all worlds",
		Options: []*discordgo.ApplicationCommandOption{
			worldOption("The world to check (omit for a summary of all worlds)", false),
		},
	},
	{
		Name:        commandCharacter,
		Description: "Look up a character on the Lodestone",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        optionName,
				Description: "The character's exact name",
				Required:    true,
			},
			worldOption("The character's world", true),
		},
	},
	{
		Name:        commandFreeCompany,
		Description: "Look up a Free Company on the Lodestone",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        optionName,
				Description: "The Free Company's exact name",
				Required:    true,
			},
			worldOption("The Free Company's world", true),
		},
	},
}

func worldOption(description string, required bool) *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:         discordgo.ApplicationCommandOptionString,
		Name:         optionWorld,
		Description:  description,
		Required:     required,
		Autocomplete: true,
	}
}
