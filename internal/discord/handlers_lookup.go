package discord

import (
	"context"
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jakeale/tonberry/internal/godestone"
)

func (bot *Bot) handleCharacter(
	ctx context.Context,
	session *discordgo.Session,
	interaction *discordgo.InteractionCreate,
	data discordgo.ApplicationCommandInteractionData,
) {
	name := data.GetOption(optionName).StringValue()
	world := data.GetOption(optionWorld).StringValue()

	if !bot.deferOrLog(session, interaction) {
		return
	}

	character, err := bot.godestoneClient.FindCharacter(ctx, name, world)
	if err != nil {
		bot.editWithContent(session, interaction, lookupErrorMessage("character", name, world, err))
		return
	}

	bot.editWithEmbed(session, interaction, characterEmbed(character))
}

func (bot *Bot) handleFreeCompany(
	ctx context.Context,
	session *discordgo.Session,
	interaction *discordgo.InteractionCreate,
	data discordgo.ApplicationCommandInteractionData,
) {
	name := data.GetOption(optionName).StringValue()
	world := data.GetOption(optionWorld).StringValue()

	if !bot.deferOrLog(session, interaction) {
		return
	}

	freeCompany, err := bot.godestoneClient.FindFreeCompany(ctx, name, world)
	if err != nil {
		bot.editWithContent(session, interaction, lookupErrorMessage("Free Company", name, world, err))
		return
	}

	bot.editWithEmbed(session, interaction, freeCompanyEmbed(freeCompany))
}

func lookupErrorMessage(kind, name, world string, err error) string {
	if errors.Is(err, godestone.ErrNotFound) {
		return fmt.Sprintf("No %s named %q was found on %s.", kind, name, world)
	}
	return fmt.Sprintf("Something went wrong looking up that %s. Please try again.", kind)
}
