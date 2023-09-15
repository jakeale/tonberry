package discord

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func (bot *Bot) handleSubscribe(
	ctx context.Context,
	session *discordgo.Session,
	interaction *discordgo.InteractionCreate,
	data discordgo.ApplicationCommandInteractionData,
) {
	world := optionByName(data.Options, optionWorld).StringValue()

	if !isKnownWorld(bot.monitor.Snapshot(), world) {
		bot.respondEphemeral(session, interaction, fmt.Sprintf("Unknown world %q.", world))
		return
	}

	created, err := bot.store.AddSubscription(ctx, interaction.GuildID, interaction.ChannelID, world)
	if err != nil {
		bot.logger.Error("add subscription failed", "world", world, "error", err)
		bot.respondEphemeral(session, interaction, "Something went wrong subscribing this channel. Please try again.")
		return
	}

	if !created {
		bot.respondEphemeral(session, interaction, fmt.Sprintf("This channel is already subscribed to %s.", world))
		return
	}

	bot.respondEphemeral(session, interaction, fmt.Sprintf("Subscribed this channel to status changes for %s.", world))
}

func (bot *Bot) handleUnsubscribe(
	ctx context.Context,
	session *discordgo.Session,
	interaction *discordgo.InteractionCreate,
	data discordgo.ApplicationCommandInteractionData,
) {
	world := optionByName(data.Options, optionWorld).StringValue()

	existed, err := bot.store.RemoveSubscription(ctx, interaction.GuildID, interaction.ChannelID, world)
	if err != nil {
		bot.logger.Error("remove subscription failed", "world", world, "error", err)
		bot.respondEphemeral(session, interaction, "Something went wrong unsubscribing this channel. Please try again.")
		return
	}

	if !existed {
		bot.respondEphemeral(session, interaction, fmt.Sprintf("This channel was not subscribed to %s.", world))
		return
	}

	bot.respondEphemeral(session, interaction, fmt.Sprintf("Unsubscribed this channel from status changes for %s.", world))
}

func (bot *Bot) handleSubscriptions(
	ctx context.Context,
	session *discordgo.Session,
	interaction *discordgo.InteractionCreate,
	_ discordgo.ApplicationCommandInteractionData,
) {
	subscriptions, err := bot.store.ListSubscriptionsByGuild(ctx, interaction.GuildID)
	if err != nil {
		bot.logger.Error("list subscriptions failed", "guild_id", interaction.GuildID, "error", err)
		bot.respondEphemeral(session, interaction, "Something went wrong listing subscriptions. Please try again.")
		return
	}

	if len(subscriptions) == 0 {
		bot.respondEphemeral(session, interaction, "This server has no world status subscriptions yet.")
		return
	}

	content := "**Subscriptions for this server:**\n"
	for _, subscription := range subscriptions {
		content += fmt.Sprintf("- %s in <#%s>\n", subscription.WorldName, subscription.ChannelID)
	}

	bot.respondEphemeral(session, interaction, content)
}
