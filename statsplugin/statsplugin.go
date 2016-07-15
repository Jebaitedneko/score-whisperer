package statsplugin

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/bwmarrin/discordgo"
	humanize "github.com/dustin/go-humanize"
	"github.com/ren-/score-whisperer"
)

var statsStartTime = time.Now()

func getDurationString(duration time.Duration) string {
	return fmt.Sprintf(
		"%0.2d:%02d:%02d",
		int(duration.Hours()),
		int(duration.Minutes())%60,
		int(duration.Seconds())%60,
	)
}

// StatsCommand returns bot statistics.
func StatsCommand(bot *whisperer.Bot, service whisperer.Service, message whisperer.Message, command string, parts []string) {
	stats := runtime.MemStats{}
	runtime.ReadMemStats(&stats)

	w := &tabwriter.Writer{}
	buf := &bytes.Buffer{}

	w.Init(buf, 0, 4, 0, ' ', 0)
	if service.Name() == whisperer.DiscordServiceName {
		fmt.Fprintf(w, "```\n")
	}
	fmt.Fprintf(w, "Score Whisperer: \t%s\n", whisperer.VersionString)
	if service.Name() == whisperer.DiscordServiceName {
		fmt.Fprintf(w, "Discordgo: \t%s\n", discordgo.VERSION)
	}
	fmt.Fprintf(w, "Go: \t%s\n", runtime.Version())
	fmt.Fprintf(w, "Uptime: \t%s\n", getDurationString(time.Now().Sub(statsStartTime)))
	fmt.Fprintf(w, "Memory used: \t%s / %s (%s garbage collected)\n", humanize.Bytes(stats.Alloc), humanize.Bytes(stats.Sys), humanize.Bytes(stats.TotalAlloc))
	fmt.Fprintf(w, "Concurrent tasks: \t%d\n", runtime.NumGoroutine())
	if service.Name() == whisperer.DiscordServiceName {
		fmt.Fprintf(w, "Connected servers: \t%d\n", service.ChannelCount())
		fmt.Fprintf(w, "\n```")
	} else {
		fmt.Fprintf(w, "Connected channels: \t%d\n", service.ChannelCount())
	}
	w.Flush()

	out := buf.String() + "\nMade by claymore. :heart: to iopred. be padangos"

	if service.SupportsMultiline() {
		service.SendMessage(message.Channel(), out)
	} else {
		lines := strings.Split(out, "\n")
		for _, line := range lines {
			if err := service.SendMessage(message.Channel(), line); err != nil {
				break
			}
		}
	}
}

// StatsHelp is the help for the stats command.
var StatsHelp = whisperer.NewCommandHelp("", "Lists bot statistics.")
