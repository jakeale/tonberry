package discord

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jakeale/tonberry/internal/lodestone"
	"github.com/jakeale/tonberry/internal/monitor"
	upstream "github.com/xivapi/godestone/v2"
)

const colorOnline = 0x57F287  // Discord "green"
const colorOffline = 0xED4245 // Discord "red"
const colorInfo = 0x5865F2    // Discord "blurple"

// characterEmbed renders a godestone Character as a Discord embed.
func characterEmbed(character *upstream.Character) *discordgo.MessageEmbed {
	fields := []*discordgo.MessageEmbedField{
		{Name: "World", Value: fmt.Sprintf("%s (%s)", character.World, character.DC), Inline: true},
	}

	if character.ActiveClassJob != nil {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Class/Job",
			Value:  fmt.Sprintf("%s (Lv. %d)", character.ActiveClassJob.Name, character.ActiveClassJob.Level),
			Inline: true,
		})
	}

	if character.FreeCompanyName != "" {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Free Company",
			Value:  character.FreeCompanyName,
			Inline: true,
		})
	}

	title := character.Name
	if character.Title != nil && character.Title.Name != "" {
		title = fmt.Sprintf("%s <%s>", character.Name, character.Title.Name)
	}

	return &discordgo.MessageEmbed{
		Title:     title,
		Color:     colorInfo,
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: character.Avatar},
		Image:     &discordgo.MessageEmbedImage{URL: character.Portrait},
		Fields:    fields,
	}
}

// freeCompanyEmbed renders a godestone FreeCompany as a Discord embed.
func freeCompanyEmbed(freeCompany *upstream.FreeCompany) *discordgo.MessageEmbed {
	fields := []*discordgo.MessageEmbedField{
		{Name: "World", Value: fmt.Sprintf("%s (%s)", freeCompany.World, freeCompany.DC), Inline: true},
		{Name: "Active Members", Value: fmt.Sprintf("%d", freeCompany.ActiveMemberCount), Inline: true},
		{Name: "Rank", Value: fmt.Sprintf("%d", freeCompany.Rank), Inline: true},
	}

	if freeCompany.Slogan != "" {
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Slogan", Value: freeCompany.Slogan})
	}

	title := freeCompany.Name
	if freeCompany.Tag != "" {
		title = fmt.Sprintf("%s <%s>", freeCompany.Name, freeCompany.Tag)
	}

	return &discordgo.MessageEmbed{
		Title:  title,
		Color:  colorInfo,
		Fields: fields,
	}
}

// statusEmbed renders a single world's current status.
func statusEmbed(world string, status lodestone.ServerStatus) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: world,
		Color: statusColor(status.Status),
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Status", Value: status.Status, Inline: true},
			{Name: "Category", Value: status.Category, Inline: true},
			{Name: "Character Creation", Value: status.CharacterCreationStatus, Inline: true},
		},
	}
}

// statusSummaryEmbed renders every world's status as a single compact embed, grouping
// worlds by whether they are online or offline to keep the message scannable and under
// Discord's field-count limits.
func statusSummaryEmbed(servers lodestone.Servers) *discordgo.MessageEmbed {
	worldNames := make([]string, 0, len(servers))
	for world := range servers {
		worldNames = append(worldNames, world)
	}
	sort.Strings(worldNames)

	var offline []string
	for _, world := range worldNames {
		if servers[world].Status != "Online" {
			offline = append(offline, world)
		}
	}

	offlineValue := "None"
	if len(offline) > 0 {
		offlineValue = strings.Join(offline, ", ")
	}

	return &discordgo.MessageEmbed{
		Title: "World Status Summary",
		Color: colorInfo,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Worlds Tracked", Value: fmt.Sprintf("%d", len(servers)), Inline: true},
			{Name: "Offline", Value: offlineValue},
		},
	}
}

// statusChangeEmbed renders a single detected status change for a notification.
func statusChangeEmbed(change monitor.StatusChange) *discordgo.MessageEmbed {
	fieldLabel := "Status"
	if change.Field == "characterCreationStatus" {
		fieldLabel = "Character Creation"
	}

	return &discordgo.MessageEmbed{
		Title: fmt.Sprintf("%s status changed", change.World),
		Color: statusColor(change.NewValue),
		Fields: []*discordgo.MessageEmbedField{
			{Name: fieldLabel, Value: fmt.Sprintf("%s -> %s", change.OldValue, change.NewValue)},
		},
	}
}

func statusColor(status string) int {
	if status == "Online" || strings.Contains(status, "Available") {
		return colorOnline
	}
	return colorOffline
}
