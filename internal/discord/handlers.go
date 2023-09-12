package discord

import (
	"context"
	"time"

	"github.com/bwmarrin/discordgo"
)

const interactionTimeout = 20 * time.Second

// handleInteraction routes an incoming interaction to the appropriate command handler.
func (bot *Bot) handleInteraction(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	switch interaction.Type {
	case discordgo.InteractionApplicationCommand:
		bot.handleCommand(session, interaction)
	case discordgo.InteractionApplicationCommandAutocomplete:
		bot.handleAutocomplete(session, interaction)
	}
}

func (bot *Bot) handleCommand(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	ctx, cancel := context.WithTimeout(context.Background(), interactionTimeout)
	defer cancel()

	data := interaction.ApplicationCommandData()

	switch data.Name {
	case commandSubscribe:
		bot.handleSubscribe(ctx, session, interaction, data)
	case commandUnsubscribe:
		bot.handleUnsubscribe(ctx, session, interaction, data)
	case commandSubscriptions:
		bot.handleSubscriptions(ctx, session, interaction, data)
	case commandStatus:
		bot.handleStatus(session, interaction, data)
	case commandCharacter:
		bot.handleCharacter(ctx, session, interaction, data)
	case commandFreeCompany:
		bot.handleFreeCompany(ctx, session, interaction, data)
	}
}

func (bot *Bot) handleAutocomplete(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	data := interaction.ApplicationCommandData()

	focused := focusedOption(data.Options)
	if focused == nil || focused.Name != optionWorld {
		return
	}

	choices := worldAutocompleteChoices(bot.monitor.Snapshot(), focused.StringValue())

	err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{Choices: choices},
	})
	if err != nil {
		bot.logger.Error("respond to autocomplete failed", "error", err)
	}
}

func focusedOption(options []*discordgo.ApplicationCommandInteractionDataOption) *discordgo.ApplicationCommandInteractionDataOption {
	for _, option := range options {
		if option.Focused {
			return option
		}
	}
	return nil
}

// respondEphemeral sends a plain-text reply visible only to the invoking user, used for
// validation errors and short confirmations.
func (bot *Bot) respondEphemeral(session *discordgo.Session, interaction *discordgo.InteractionCreate, content string) {
	err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		bot.logger.Error("respond to interaction failed", "error", err)
	}
}

// respondEmbed sends a public reply containing a single embed.
func (bot *Bot) respondEmbed(session *discordgo.Session, interaction *discordgo.InteractionCreate, embed *discordgo.MessageEmbed) {
	err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Embeds: []*discordgo.MessageEmbed{embed}},
	})
	if err != nil {
		bot.logger.Error("respond to interaction failed", "error", err)
	}
}

// deferResponse acknowledges the interaction immediately, buying time for a slow
// lookup. Discord requires an ack within 3 seconds; the real result is sent later
// via InteractionResponseEdit.
func (bot *Bot) deferResponse(session *discordgo.Session, interaction *discordgo.InteractionCreate) error {
	return session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
}

// deferOrLog defers the interaction and logs on failure, reporting whether the caller
// should continue with its (slow) lookup.
func (bot *Bot) deferOrLog(session *discordgo.Session, interaction *discordgo.InteractionCreate) bool {
	if err := bot.deferResponse(session, interaction); err != nil {
		bot.logger.Error("defer interaction response failed", "error", err)
		return false
	}
	return true
}

func (bot *Bot) editWithEmbed(session *discordgo.Session, interaction *discordgo.InteractionCreate, embed *discordgo.MessageEmbed) {
	_, err := session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})
	if err != nil {
		bot.logger.Error("edit interaction response failed", "error", err)
	}
}

func (bot *Bot) editWithContent(session *discordgo.Session, interaction *discordgo.InteractionCreate, content string) {
	_, err := session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
	if err != nil {
		bot.logger.Error("edit interaction response failed", "error", err)
	}
}
