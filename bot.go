package main

// Ideas:
// Calculate the standard deviation of top scores
// TOP pp, lowest PP
// Highest Accuracy, lowest accuracy
import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
	API "github.com/ren-/osu/api"
)

var APIConnection API.Config
var db *sqlx.DB

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	//https://discordapp.com/oauth2/authorize?client_id=189474870923362305&scope=bot&permissions=8
	Token := os.Getenv("DG_TOKEN")
	// Create a new Discord session using the provided login information.
	// Use discordgo.New(Token) to just use a token for login.
	dg, err := discordgo.New(Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}
	APIConnection.SetAPIKey(os.Getenv("OSU_TOKEN"))

	db, err = sqlx.Connect("postgres", "user="+os.Getenv("DB_USER")+" dbname="+os.Getenv("DB_DATABASE")+" sslmode=disable")
	if err != nil {
		log.Fatalln(err)
	}

	// Register messageCreate as a callback for the messageCreate events.
	dg.AddHandler(messageCreate)

	// Open the websocket and begin listening.
	dg.Open()

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")

	// channels := getChannelsForScores(dg)

	//scores, err := APIConnection.GetUserBest("asdadsadadadad", API.OSU, 5)
	//fmt.Println(scores)
	if err != nil {
		fmt.Println(err)
	}

	// Simple way to keep program running until CTRL-C is pressed.
	<-make(chan struct{})
	return
}

// This function will fetch all the channels that match the "#scores" name
// Those channels will be used to announce information
func getChannelsByName(s *discordgo.Session, channelName string) []string {
	var guildIDs []string
	var channelIDs []string

	guilds, err := s.UserGuilds()

	if err != nil {
		fmt.Println("Bot is not connected to any guild, ", err)
	}

	for _, element := range guilds {
		guildIDs = append(guildIDs, element.ID)
	}

	for _, element := range guildIDs {
		channels, err := s.GuildChannels(element)
		if err != nil {
			fmt.Println("This guild doesn't have any channels, ", err)
		}
		for _, item := range channels {
			if item.Name == channelName {
				channelIDs = append(channelIDs, item.ID)
			}
		}
	}

	return channelIDs
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	log.Printf("[%5s]: %5s > %s\n", m.ChannelID, m.Author.Username, m.Content)

	if strings.HasPrefix(m.Content, "!info") {

		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)

		results := fmt.Sprintf(
			"```Processes:\t%d\n"+
				"HeapAlloc:\t%.2fMB\n"+
				"Total Sys:\t%.2fMB\n"+
				"```*Written by* ***Ren*** *a.k.a.* ***claymore***",
			runtime.NumGoroutine(), float64(mem.HeapAlloc)/1048576, float64(mem.Sys)/1048576)
		s.ChannelMessageSend(m.ChannelID, results)

	}

	if strings.HasPrefix(m.Content, "!help") {

	}

	// if strings.HasPrefix(m.Content, "!stats ") {
	// 	var second = strings.Split(m.Content, " ")
	// 	if len(second) == 2 {
	// 		scores, err := APIConnection.GetUserBest(second[1], API.OSU, 100)
	// 		if err != nil {
	// 			fmt.Printf("HTTP: %s", err)
	// 		} else {
	// 			if len(scores) == 0 {
	// 				s.ChannelMessageSend(m.ChannelID, "No information available for specified user.")
	// 			} else {
	// 				info := stats.Calculate(scores)
	// 				results := fmt.Sprintf(
	// 					"**Highest PP gained**: %f\n"+
	// 						"**Lowest PP gained**: %f\n"+
	// 						"**Average PP**: %f\n"+
	// 						"**Standard deviation of PP**: %f\n"+
	// 						"**Median absolute deviation**: %f\n",
	// 					info.HighestPP, info.LowestPP, info.AveragePP, info.StandardDeviation, info.MedianAbsoluteDeviation)

	// 				//s.ChannelMessageSend(m.ChannelID, "Ok, <@"+m.Author.ID+">, timer for "+second[1]+" minutes!")
	// 				s.ChannelMessageSend(m.ChannelID, results)
	// 			}
	// 		}
	// 	} else {
	// 		s.ChannelMessageSend(m.ChannelID, "Usage: !stats <osu_username>")
	// 	}
	// }

	// Print message to stdout.
	//fmt.Printf("%20s %20s %20s > %s\n", m.ChannelID, time.Now().Format(time.Stamp), m.Author.Username, m.Content)
}
