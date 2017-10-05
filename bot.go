/*
 * Copyright (c) 2017 Ethan Wells
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	prefix string = ";"
	token  string
	dg     *discordgo.Session
)

func main() {
	err := LoadCalenders()
	if err != nil {
		fmt.Println(err)
	}

	go func() {
		t := time.NewTicker(time.Minute)
		defer t.Stop()
		for {
			<-t.C
			rand.Seed(time.Now().Unix())
		}
	}()

	token = Token()
	if token == "" {
		fmt.Println("No token provided. Please place a token followed by a ';' at token.txt")
		return
	}

	dg, err = discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}

	dg.AddHandler(ready)
	dg.AddHandler(guildCreate)
	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
	}

	go alertEvents()

	fmt.Println("The bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()
	os.Exit(0)
}

func guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	if event.Guild.Unavailable {
		return
	}

	for _, channel := range event.Guild.Channels {
		if channel.ID == "315552571823489024" {
			_, _ = s.ChannelMessageSend(channel.ID, "Connected.")
			return
		}
	}
}

func Token() string {
	b, err := ioutil.ReadFile("token.txt")
	if err != nil {
		fmt.Println(err)
	}
	c := string(b)
	return strings.Split(c, ";")[0]
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateStatus(0, prefix+"help")
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if len(m.Message.Content) < 1 {
		return
	}
	if m.Message.ContentWithMentionsReplaced()[:len(prefix)] != prefix {
		return
	}

	c := strings.Split(strings.TrimSpace(strings.TrimPrefix(m.Message.Content, prefix)), " ")
	c[0] = strings.ToLower(c[0])

	if c[0] == "ping" {
		s.ChannelMessageSend(m.ChannelID, "pong!")
		return
	}

	if c[0] == "help" {
		var e *discordgo.MessageEmbed
		e = &discordgo.MessageEmbed{
			Title:       "Help",
			Description: "How to use this here xkcd bot",
			URL:         "https://xkcd.com/",
			Color:       7506394,
			Type:        "rich",
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:  "help",
					Value: "Display this message",
				},
				{
					Name:  "xkcd <comic number, name, regex or whatever>",
					Value: "Get the designated comic",
				},
				{
					Name:  "latest",
					Value: "The latest and greatest xkcd",
				},
				{
					Name:  "random",
					Value: "A random comic",
				},
				{
					Name:  "event",
					Value: "An interface for interacting with calender events. Say `;help` event for more!",
				},
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text:    "@" + m.Author.String(),
				IconURL: "https://cdn.discordapp.com/avatars/" + m.Author.ID + "/" + m.Author.Avatar + ".png",
			},
		}

		if len(c) > 1 {
			if c[1] == "event" {
				e = &discordgo.MessageEmbed{
					Title:       "Help",
					Description: "How to use `event` commands on this here xkcd bot",
					URL:         "https://xkcd.com/",
					Color:       7506394,
					Type:        "rich",
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:  TimeFormat,
							Value: "When creating events, be sure to use this format for time",
						},
						{
							Name:  "new <title>; <description>; <participants (mention them)>; <date/time>",
							Value: "Create a new event. Note that you can use spaces in the fields, but the fields are seperated by semicolons",
						},
						{
							Name:  "list",
							Value: "List the events that exist",
						},
					},
					Footer: &discordgo.MessageEmbedFooter{
						Text:    "@" + m.Author.String(),
						IconURL: "https://cdn.discordapp.com/avatars/" + m.Author.ID + "/" + m.Author.Avatar + ".png",
					},
				}
			}
		}

		_, err := s.ChannelMessageSendEmbed(m.ChannelID, e)
		if err != nil {
			fmt.Println(err)
		}
		return
	}

	if c[0] == "xkcd" {
		if len(c) < 2 {
			c = append(c, "")
		}
		var (
			xkcd XKCD
			err  error
			r    *regexp.Regexp
		)

		r, err = regexp.Compile("^[0-9]+$")
		if err != nil {
			fmt.Println(err)
			return
		}

		if r.MatchString(c[1]) {
			xkcd, err = GetXkcdNum(c[1])
		} else {
			xkcd, err = GetXkcdTitle(strings.Join(c[1:], " "))
		}
		if err != nil {
			fmt.Println(err)
			return
		}

		e := &discordgo.MessageEmbed{
			Title:       "xkcd #" + strconv.Itoa(xkcd.Num) + ": " + xkcd.Title,
			Description: xkcd.Alt,
			URL:         "https://xkcd.com/" + strconv.Itoa(xkcd.Num),
			Color:       7506394,
			Type:        "rich",
			Image: &discordgo.MessageEmbedImage{
				URL: xkcd.Img,
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text:    "@" + m.Author.String(),
				IconURL: "https://cdn.discordapp.com/avatars/" + m.Author.ID + "/" + m.Author.Avatar + ".png",
			},
		}

		_, err = s.ChannelMessageSendEmbed(m.ChannelID, e)
		if err != nil {
			fmt.Println(err)
			s.ChannelMessageSend(m.ChannelID, xkcd.Img)
		}
		return
	}

	if c[0] == "latest" {
		xkcd, err := GetLatest()
		if err != nil {
			fmt.Println(err)
		}

		e := &discordgo.MessageEmbed{
			Title:       "xkcd #" + strconv.Itoa(xkcd.Num) + ": " + xkcd.Title,
			Description: xkcd.Alt,
			URL:         "https://xkcd.com/" + strconv.Itoa(xkcd.Num),
			Color:       7506394,
			Type:        "rich",
			Image: &discordgo.MessageEmbedImage{
				URL: xkcd.Img,
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text:    "@" + m.Author.String(),
				IconURL: "https://cdn.discordapp.com/avatars/" + m.Author.ID + "/" + m.Author.Avatar + ".png",
			},
		}

		_, err = s.ChannelMessageSendEmbed(m.ChannelID, e)
		if err != nil {
			fmt.Println(err)
			s.ChannelMessageSend(m.ChannelID, xkcd.Img)
		}
		return
	}

	if c[0] == "random" {
		xkcd, err := GetLatest()
		if err != nil {
			fmt.Println(err)
		}
		xkcd, err = GetXkcdNum(strconv.Itoa(rand.Intn(xkcd.Num + 1)))
		if err != nil {
			fmt.Println(err)
		}

		e := &discordgo.MessageEmbed{
			Title:       "xkcd #" + strconv.Itoa(xkcd.Num) + ": " + xkcd.Title,
			Description: xkcd.Alt,
			URL:         "https://xkcd.com/" + strconv.Itoa(xkcd.Num),
			Color:       7506394,
			Type:        "rich",
			Image: &discordgo.MessageEmbedImage{
				URL: xkcd.Img,
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text:    "@" + m.Author.String(),
				IconURL: "https://cdn.discordapp.com/avatars/" + m.Author.ID + "/" + m.Author.Avatar + ".png",
			},
		}

		_, err = s.ChannelMessageSendEmbed(m.ChannelID, e)
		if err != nil {
			fmt.Println(err)
			s.ChannelMessageSend(m.ChannelID, xkcd.Img)
		}
		return
	}

	if c[0] == "play" {
		if len(c) < 2 {
			return
		}
		o := strings.Join(c[1:], " ")

		var (
			err   error
			tch   *discordgo.Channel
			vch   *discordgo.VoiceState
			guild *discordgo.Guild
		)

		tch, err = s.Channel(m.Message.ChannelID)
		if err != nil {
			fmt.Println(err)
			return
		}
		guild, err = s.Guild(tch.GuildID)
		if err != nil {
			fmt.Println(err)
			return
		}

		for _, vs := range guild.VoiceStates {
			if vs.UserID == m.Author.ID {
				vch = vs
				break
			}
		}
		if vch == nil {
			fmt.Println("User not joined channel")
			return
		}

		fmt.Println(o)
		fmt.Println(s.ShardCount)
		fmt.Println(tch.GuildID)
		fmt.Println(vch.ChannelID)

		err = Stream(o, s, tch.GuildID, vch.ChannelID)
		if err != nil {
			fmt.Println(err)
		}
		return
	}

	if c[0] == "event" {
		if len(c) < 2 {
			return
		}

		if c[1] == "new" {
			co := strings.Split(strings.Join(c[2:], " "), ";")
			for i, ci := range co {
				co[i] = strings.TrimSpace(ci)
			}
			if len(co) < 4 {
				return
			}
			f, erro := s.Channel(m.ChannelID)
			if erro != nil {
				fmt.Println(erro)
				return
			}
			if _, err := time.Parse(TimeFormat, co[3]); err != nil {
				co[3] = time.Now().Add(time.Hour * 24).Format(TimeFormat)
			}

			var (
				event *Calender
				err   error
			)
			if len(co) > 4 {
				event, err = NewCalender(co[0], co[1], co[2], co[3], f.GuildID, c[4], m.Author.ID)
			} else {
				event, err = NewCalender(co[0], co[1], co[2], co[3], f.GuildID, f.ID, m.Author.ID)
			}
			if err != nil {
				fmt.Println(err)
				return
			}
			e := &discordgo.MessageEmbed{
				Color: 7506394,
				Type:  "rich",
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Event Created: " + event.Title,
						Value:  event.Description,
						Inline: true,
					},
					{
						Name:   "Time",
						Value:  event.Date,
						Inline: true,
					},
					{
						Name:   "Participants",
						Value:  event.Participants,
						Inline: true,
					},
				},
				Footer: &discordgo.MessageEmbedFooter{
					Text:    "@" + m.Author.String(),
					IconURL: "https://cdn.discordapp.com/avatars/" + m.Author.ID + "/" + m.Author.Avatar + ".png",
				},
			}

			_, err = s.ChannelMessageSendEmbed(m.ChannelID, e)
			if err != nil {
				fmt.Println(err)
			}
			return
		}

		if c[1] == "list" {
			var fields []*discordgo.MessageEmbedField
			for _, event := range Events {
				fields = append(fields,
					&discordgo.MessageEmbedField{
						Name:  event.Title,
						Value: event.Date,
					})
			}
			e := &discordgo.MessageEmbed{
				Color:  7506394,
				Type:   "rich",
				Fields: fields,
				Footer: &discordgo.MessageEmbedFooter{
					Text:    "@" + m.Author.String(),
					IconURL: "https://cdn.discordapp.com/avatars/" + m.Author.ID + "/" + m.Author.Avatar + ".png",
				},
			}

			_, err := s.ChannelMessageSendEmbed(m.ChannelID, e)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func alertEvents() {
	t := time.NewTicker(5 * time.Second)
	var e *discordgo.MessageEmbed
	for {
		for i, event := range Events {
			et, err := time.Parse(TimeFormat, event.Date)
			if err != nil {
				fmt.Println(err)
				continue
			}

			author, err := dg.User(event.AuthorID)
			if err != nil {
				fmt.Println(err)
				e = &discordgo.MessageEmbed{
					Color: 7506394,
					Type:  "rich",
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:   event.Title,
							Value:  event.Description,
							Inline: true,
						},
						{
							Name:   "Time",
							Value:  event.Date,
							Inline: true,
						},
						{
							Name:   "Participants",
							Value:  "@" + event.Participants,
							Inline: true,
						},
					},
				}
			} else if time.Now().After(et) {
				e = &discordgo.MessageEmbed{
					Color: 7506394,
					Type:  "rich",
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:   event.Title,
							Value:  event.Description,
							Inline: true,
						},
						{
							Name:   "Time",
							Value:  event.Date,
							Inline: true,
						},
						{
							Name:   "Participants",
							Value:  event.Participants,
							Inline: true,
						},
					},
					Footer: &discordgo.MessageEmbedFooter{
						Text:    "@" + author.String(),
						IconURL: "https://cdn.discordapp.com/avatars/" + author.ID + "/" + author.Avatar + ".png",
					},
				}

				_, err := dg.ChannelMessageSendEmbed(event.ChannelID, e)
				if err != nil {
					fmt.Println(err)
				}
				Events = append(Events[:i], Events[i+1:]...)
				go SaveCalenders()
			}
		}
		<-t.C
	}
}
