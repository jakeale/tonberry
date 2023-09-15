package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func (bot *Bot) handleStatus(
	session *discordgo.Session,
	interaction *discordgo.InteractionCreate,
	data discordgo.ApplicationCommandInteractionData,
) {
	servers := bot.monitor.Snapshot()

	worldOpt := optionByName(data.Options, optionWorld)
	if worldOpt == nil {
		bot.respondEmbed(session, interaction, statusSummaryEmbed(servers))
		return
	}

	world := worldOpt.StringValue()

	status, found := servers[world]
	if !found {
		bot.respondEphemeral(session, interaction, fmt.Sprintf("Unknown world %q.", world))
		return
	}

	bot.respondEmbed(session, interaction, statusEmbed(world, status))
}
