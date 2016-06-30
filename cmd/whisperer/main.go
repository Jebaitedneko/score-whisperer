package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/joho/godotenv"
	"github.com/ren-/score-whisperer"
	"github.com/ren-/score-whisperer/inviteplugin"
	"github.com/ren-/score-whisperer/playingplugin"
	"github.com/ren-/score-whisperer/statsplugin"
	"github.com/ren-/score-whisperer/triesplugin"
)

var discordApplicationClientID string
var discordOwnerUserID string

func init() {
	flag.StringVar(&discordOwnerUserID, "discordowneruserid", "", "Discord owner user id.")
	flag.StringVar(&discordApplicationClientID, "discordapplicationclientid", "", "Discord application client id.")
	fmt.Println(discordOwnerUserID)
	fmt.Println(discordApplicationClientID)

	flag.Parse()

	rand.Seed(time.Now().UnixNano())
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	//https://discordapp.com/oauth2/authorize?client_id=189474870923362305&scope=bot&permissions=8
	DiscordToken := os.Getenv("DG_TOKEN")
	// Create a new Discord session using the provided login information.
	// Use discordgo.New(Token) to just use a token for login.

	// Set our variables.
	bot := whisperer.NewBot()

	// Generally CommandPlugins don't hold state, so we share one instance of the command plugin for all services.
	cp := whisperer.NewCommandPlugin()
	cp.AddCommand("stats", statsplugin.StatsCommand, statsplugin.StatsHelp)
	cp.AddCommand("invite", inviteplugin.InviteCommand, inviteplugin.InviteHelp)

	// Register the Discord service if we have an email or token.
	if DiscordToken != "" {
		var discord *whisperer.Discord
		discord = whisperer.NewDiscord(DiscordToken)

		discord.ApplicationClientID = discordApplicationClientID
		discord.OwnerUserID = discordOwnerUserID

		bot.RegisterService(discord)
		bot.RegisterPlugin(discord, cp)
		bot.RegisterPlugin(discord, playingplugin.New())
		bot.RegisterPlugin(discord, triesplugin.New())

	}

	// Start all our services.
	bot.Open()

	// Wait for a termination signal, while saving the bot state every minute. Save on close.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	t := time.Tick(1 * time.Minute)

	for {
		select {
		case <-c:
			bot.Save()
			return
		case <-t:
			bot.Save()
		}
	}
}
