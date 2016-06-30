package playingplugin

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/ren-/score-whisperer"
)

type playingPlugin struct {
	whisperer.SimplePlugin
	Game string
	URL  string
}

// Name returns the name of the plugin.
func (p *playingPlugin) Name() string {
	return "Playing"
}

// Load will load plugin state from a byte array.
func (p *playingPlugin) Load(bot *whisperer.Bot, service whisperer.Service, data []byte) error {
	if service.Name() != whisperer.DiscordServiceName {
		panic("Playing Plugin only supports Discord.")
	}

	if data != nil {
		if err := json.Unmarshal(data, p); err != nil {
			log.Println("Error loading data", err)
		}
	}

	service.(*whisperer.Discord).Session.UpdateStreamingStatus(0, p.Game, p.URL)

	return nil
}

// Save will save plugin state to a byte array.
func (p *playingPlugin) Save() ([]byte, error) {
	return json.Marshal(p)
}

// Help returns a list of help strings that are printed when the user requests them.
func (p *playingPlugin) helpFunc(bot *whisperer.Bot, service whisperer.Service, message whisperer.Message, detailed bool) []string {
	if detailed {
		return nil
	}

	if !service.IsBotOwner(message) {
		return nil
	}

	return whisperer.CommandHelp(service, "playing", "<game>, <url>", fmt.Sprintf("Set which game %s is playing.", service.UserName()))
}

// Message handler.
func (p *playingPlugin) messageFunc(bot *whisperer.Bot, service whisperer.Service, message whisperer.Message) {
	if !service.IsMe(message) {
		if whisperer.MatchesCommand(service, "playing", message) {
			if !service.IsBotOwner(message) {
				return
			}
			query, _ := whisperer.ParseCommand(service, message)

			split := strings.Split(query, ",")

			p.Game = strings.Trim(split[0], " ")
			if len(split) > 1 {
				p.URL = strings.Trim(split[1], " ")
			} else {
				p.URL = ""
			}

			service.(*whisperer.Discord).Session.UpdateStreamingStatus(0, p.Game, p.URL)
		}
	}
}

// New will create a new top streamers plugin.
func New() whisperer.Plugin {
	p := &playingPlugin{}
	p.MessageFunc = p.messageFunc
	p.HelpFunc = p.helpFunc
	return p
}
