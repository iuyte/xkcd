package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var token string

func main() {
	token = Token()
	if token == "" {
		fmt.Println("No token provided. Please place a token followed by a ';' at ../token.txt")
		return
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}

	dg.AddHandler(ready)
	dg.AddHandler(guildCreate)
	dg.AddHandler(messageCreate)
	dg.AddHandler(messageEdit)
	dg.AddHandler(messageDelete)
	dg.AddHandler(messageBulkDelete)

	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
	}

	fmt.Println("The bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()
}

func Token() string {
	b, err := ioutil.ReadFile("../token.txt")
	if err != nil {
		fmt.Println(err)
	}
	c := string(b)
	c = "MzU2NDMxMjk4NjEwMzMxNjU5.DJbQjA.bmm2oT9IhyzGdHDIAaIv6cImONY"
	return strings.Split(c, ";")[0]
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateStatus(0, "Always watching...")
}

func guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {

	if event.Guild.Unavailable {
		return
	}

	for _, channel := range event.Guild.Channels {
		if channel.ID == "291598452331118592" {
			_, _ = s.ChannelMessageSend(channel.ID, "Gobot the Gopher has *ARRIVED!*")
			return
		}
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	b, err := json.Marshal(m.Message)
	if err != nil {
		return
	}
	ioutil.WriteFile("db.json", b, os.ModeAppend)
}

func messageDelete(s *discordgo.Session, m *discordgo.MessageDelete) {
	b, err := json.Marshal(m.Message)
	if err != nil {
		return
	}
	ioutil.WriteFile("db.json", b, os.ModeAppend)
}

func messageEdit(s *discordgo.Session, m *discordgo.MessageEdit) {
	b, err := json.Marshal(m)
	if err != nil {
		return
	}
	ioutil.WriteFile("db.json", b, os.ModeAppend)
}

func messageBulkDelete(s *discordgo.Session, m *discordgo.MessageDeleteBulk) {
	b, err := json.Marshal(m)
	if err != nil {
		return
	}
	ioutil.WriteFile("db.json", b, os.ModeAppend)
}
