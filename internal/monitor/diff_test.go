package monitor

import (
	"reflect"
	"testing"

	"github.com/jakeale/tonberry/internal/lodestone"
)

func TestDiffServers(t *testing.T) {
	tests := []struct {
		name string
		prev lodestone.Servers
		next lodestone.Servers
		want []StatusChange
	}{
		{
			name: "no change",
			prev: lodestone.Servers{
				"Adamantoise": {Status: "Online", Category: "Standard", CharacterCreationStatus: "Creation of New Characters Available"},
			},
			next: lodestone.Servers{
				"Adamantoise": {Status: "Online", Category: "Standard", CharacterCreationStatus: "Creation of New Characters Available"},
			},
			want: []StatusChange{},
		},
		{
			name: "status change",
			prev: lodestone.Servers{
				"Adamantoise": {Status: "Online", CharacterCreationStatus: "Creation of New Characters Available"},
			},
			next: lodestone.Servers{
				"Adamantoise": {Status: "Offline", CharacterCreationStatus: "Creation of New Characters Available"},
			},
			want: []StatusChange{
				{World: "Adamantoise", Field: fieldStatus, OldValue: "Online", NewValue: "Offline"},
			},
		},
		{
			name: "character creation status change",
			prev: lodestone.Servers{
				"Adamantoise": {Status: "Online", CharacterCreationStatus: "Creation of New Characters Available"},
			},
			next: lodestone.Servers{
				"Adamantoise": {Status: "Online", CharacterCreationStatus: "Creation of New Characters Unavailable"},
			},
			want: []StatusChange{
				{World: "Adamantoise", Field: fieldCharacterCreationStatus, OldValue: "Creation of New Characters Available", NewValue: "Creation of New Characters Unavailable"},
			},
		},
		{
			name: "category change is not reported",
			prev: lodestone.Servers{
				"Adamantoise": {Status: "Online", Category: "Standard"},
			},
			next: lodestone.Servers{
				"Adamantoise": {Status: "Online", Category: "Preferred"},
			},
			want: []StatusChange{},
		},
		{
			name: "world added is not reported",
			prev: lodestone.Servers{},
			next: lodestone.Servers{
				"Adamantoise": {Status: "Online"},
			},
			want: []StatusChange{},
		},
		{
			name: "world removed is not reported",
			prev: lodestone.Servers{
				"Adamantoise": {Status: "Online"},
			},
			next: lodestone.Servers{},
			want: []StatusChange{},
		},
		{
			name: "multiple worlds change, sorted by world then field",
			prev: lodestone.Servers{
				"Zalera":      {Status: "Online", CharacterCreationStatus: "Creation of New Characters Available"},
				"Adamantoise": {Status: "Online", CharacterCreationStatus: "Creation of New Characters Available"},
			},
			next: lodestone.Servers{
				"Zalera":      {Status: "Offline", CharacterCreationStatus: "Creation of New Characters Available"},
				"Adamantoise": {Status: "Offline", CharacterCreationStatus: "Creation of New Characters Unavailable"},
			},
			want: []StatusChange{
				{World: "Adamantoise", Field: fieldCharacterCreationStatus, OldValue: "Creation of New Characters Available", NewValue: "Creation of New Characters Unavailable"},
				{World: "Adamantoise", Field: fieldStatus, OldValue: "Online", NewValue: "Offline"},
				{World: "Zalera", Field: fieldStatus, OldValue: "Online", NewValue: "Offline"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := diffServers(test.prev, test.next)

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("diffServers() = %+v, want %+v", got, test.want)
			}
		})
	}
}
