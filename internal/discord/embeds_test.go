package discord

import (
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/jakeale/tonberry/internal/lodestone"
	"github.com/jakeale/tonberry/internal/monitor"
	upstream "github.com/xivapi/godestone/v2"
)

func TestCharacterEmbed(t *testing.T) {
	character := &upstream.Character{
		Name:            "Potato Chippy",
		World:           "Balmung",
		DC:              "Crystal",
		FreeCompanyName: "Chips Ahoy",
		ActiveClassJob:  &upstream.ClassJob{Name: "Paladin", Level: 90},
	}

	embed := characterEmbed(character)

	if embed.Title != "Potato Chippy" {
		t.Errorf("Title = %q, want %q", embed.Title, "Potato Chippy")
	}

	if !containsField(embed.Fields, "World", "Balmung (Crystal)") {
		t.Error("expected a World field with \"Balmung (Crystal)\"")
	}

	if !containsField(embed.Fields, "Class/Job", "Paladin (Lv. 90)") {
		t.Error("expected a Class/Job field with \"Paladin (Lv. 90)\"")
	}

	if !containsField(embed.Fields, "Free Company", "Chips Ahoy") {
		t.Error("expected a Free Company field with \"Chips Ahoy\"")
	}
}

func TestFreeCompanyEmbed(t *testing.T) {
	freeCompany := &upstream.FreeCompany{
		Name:              "Chips Ahoy",
		Tag:               "CHIP",
		World:             "Balmung",
		DC:                "Crystal",
		ActiveMemberCount: 42,
		Rank:              7,
		Slogan:            "Crunch time",
	}

	embed := freeCompanyEmbed(freeCompany)

	if embed.Title != "Chips Ahoy <CHIP>" {
		t.Errorf("Title = %q, want %q", embed.Title, "Chips Ahoy <CHIP>")
	}

	if !containsField(embed.Fields, "Active Members", "42") {
		t.Error("expected an Active Members field with \"42\"")
	}

	if !containsField(embed.Fields, "Slogan", "Crunch time") {
		t.Error("expected a Slogan field with \"Crunch time\"")
	}
}

func TestStatusSummaryEmbed(t *testing.T) {
	servers := lodestone.Servers{
		"Adamantoise": {Status: "Online"},
		"Zalera":      {Status: "Offline"},
	}

	embed := statusSummaryEmbed(servers)

	if !containsField(embed.Fields, "Worlds Tracked", "2") {
		t.Error("expected a Worlds Tracked field with \"2\"")
	}

	if !containsFieldContaining(embed.Fields, "Offline", "Zalera") {
		t.Error("expected the Offline field to mention Zalera")
	}
}

func TestStatusChangeEmbed_ColorReflectsNewValue(t *testing.T) {
	onlineChange := monitor.StatusChange{World: "Zalera", Field: "status", OldValue: "Offline", NewValue: "Online"}
	offlineChange := monitor.StatusChange{World: "Zalera", Field: "status", OldValue: "Online", NewValue: "Offline"}

	onlineEmbed := statusChangeEmbed(onlineChange)
	offlineEmbed := statusChangeEmbed(offlineChange)

	if onlineEmbed.Color != colorOnline {
		t.Errorf("online change color = %#x, want %#x", onlineEmbed.Color, colorOnline)
	}

	if offlineEmbed.Color != colorOffline {
		t.Errorf("offline change color = %#x, want %#x", offlineEmbed.Color, colorOffline)
	}
}

func containsField(fields []*discordgo.MessageEmbedField, name, value string) bool {
	for _, field := range fields {
		if field.Name == name && field.Value == value {
			return true
		}
	}
	return false
}

func containsFieldContaining(fields []*discordgo.MessageEmbedField, name, substring string) bool {
	for _, field := range fields {
		if field.Name == name && strings.Contains(field.Value, substring) {
			return true
		}
	}
	return false
}
