package triesplugin

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/jmoiron/sqlx"
	"github.com/ren-/osu/api"
	"github.com/ren-/score-whisperer"
)

var db *sqlx.DB

func triesHelpFunc(bot *whisperer.Bot, service whisperer.Service, message whisperer.Message, detailed bool) []string {
	return whisperer.CommandHelp(service, "tries", "<beatmap_id>, <username>", fmt.Sprintf("Get information about plays on specific beatmap during the last 24 hours"))
}

func triesMessageFunc(bot *whisperer.Bot, service whisperer.Service, message whisperer.Message) {
	if !service.IsMe(message) {
		if whisperer.MatchesCommand(service, "tries", message) {
			query, _ := whisperer.ParseCommand(service, message)
			//fmt.Printf("Query: %s", query)

			split := strings.Split(query, ",")

			if len(split) < 2 {
				return
			}
			beatmap_id := strings.Trim(split[0], " ")
			username := strings.Trim(split[1], " ")

			db, err := sqlx.Connect("postgres", "host="+os.Getenv("DB_HOST")+" user="+os.Getenv("DB_USER")+" dbname="+os.Getenv("DB_DATABASE")+" password="+os.Getenv("DB_PASSWORD")+" sslmode=disable")
			if err != nil {
				log.Fatalln(err)
			}

			stmt, err := db.Preparex("SELECT beatmap_id, score, max_combo, count50, count100, count300, count_miss, count_katu, count_geki, enabled_mods, user_id, date, rank, username FROM plays WHERE LOWER(username)=LOWER($1) AND beatmap_id=$2 AND date at time zone 'UTC+8' > current_timestamp - interval '24 hours' order by date asc")
			songs := []api.Song{}
			err = stmt.Select(&songs, username, beatmap_id)
			if err != nil {
				fmt.Println(err)
			}

			if len(songs) == 0 {
				return
			}
			totalPoints := 0
			numberHits := 0
			combos := []string{}
			for _, element := range songs {
				totalPoints += (element.Count50*50 + element.Count100*100 + element.Count300*300)
				numberHits += (element.CountMiss + element.Count50 + element.Count100 + element.Count300)
				combos = append(combos, strconv.Itoa(element.MaxCombo))
				fmt.Printf("%v\n%v \n\n\n", totalPoints, numberHits)
			}
			var accuracy float64
			accuracy = (float64(totalPoints) / float64((numberHits * 300))) * 100
			fmt.Println(accuracy)

			w := &tabwriter.Writer{}
			buf := &bytes.Buffer{}

			w.Init(buf, 0, 4, 0, ' ', 0)
			fmt.Fprintf(w, "```\n")
			fmt.Fprintf(w, "Number of plays: \t%d\n", len(songs))
			fmt.Fprintf(w, "Average accuracy: \t%f%%\n", accuracy)
			fmt.Fprintf(w, "Combo: \t%s\n", strings.Join(combos, ", "))
			fmt.Fprintf(w, "```\n")
			w.Flush()

			out := buf.String()

			service.SendMessage(message.Channel(), out)
			fmt.Println(songs)

			fmt.Println(beatmap_id, username)
		}
	}
}

// New will create a new Reminder plugin.
func New() whisperer.Plugin {
	p := whisperer.NewSimplePlugin("Tries")
	p.HelpFunc = triesHelpFunc
	p.MessageFunc = triesMessageFunc

	return p
}
