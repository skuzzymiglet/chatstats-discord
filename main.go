package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/wcharczuk/go-chart"
)

// Variables used for command line parameters
var (
	Token string
)

func init() {

	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

func countByDate(messages []*discordgo.Message) map[time.Time]int {
	byDate := make(map[time.Time]int)
	for _, m := range messages {
		t, err := discordgo.SnowflakeTimestamp(m.ID)
		if err != nil {
			log.Fatal(err)
		}
		d := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
		_, ok := byDate[d]
		if !ok {
			byDate[d] = 1
		} else {
			byDate[d]++
		}
	}
	return byDate
}

var mtime time.Time

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	prefix := "!cs "

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	fmt.Println("Command", m.Content, "in channel", string(m.ChannelID), "in guild", string(m.GuildID))

	// If the message is "ping" reply with "Pong!"
	if m.Content == prefix+"ping" {
		mtime, err := discordgo.SnowflakeTimestamp(m.ID)
		if err != nil {
			log.Fatal(err)
		}
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Time: %v\n", time.Now().Sub(mtime)))
	}

	if m.Content == prefix+"graph" {
		graph := chart.Chart{
			Series: []chart.Series{
				chart.ContinuousSeries{
					XValues: []float64{1.0, 2.0, 3.0, 4.0},
					YValues: []float64{1.0, 2.0, 3.0, 4.0},
				},
			},
		}
		buffer := bytes.NewBuffer([]byte{})
		graph.Render(chart.PNG, buffer)
		_, err := s.ChannelFileSend(m.ChannelID, "graph.png", buffer)
		if err != nil {
			log.Fatal(err)
		}
	}

	if m.Content == "!cs" {
		log.Println("Creating default graph")
		msgs, err := s.ChannelMessages(m.ChannelID, 100, "", "", "")
		if err != nil {
			log.Fatal(err)
		}
		var keys []time.Time
		var values []float64
		for k, v := range countByDate(msgs) {
			keys = append(keys, k)
			values = append(values, float64(v))
		}
		graph := chart.Chart{
			Series: []chart.Series{
				chart.TimeSeries{
					XValues: keys,
					YValues: values,
				},
			},
		}
		buffer := bytes.NewBuffer([]byte{})
		graph.Render(chart.PNG, buffer)
		_, err = s.ChannelFileSend(m.ChannelID, "graph.png", buffer)
		if err != nil {
			log.Fatal(err)
		}
	}
}
