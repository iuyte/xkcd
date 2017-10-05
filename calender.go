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
	"encoding/json"
	"io/ioutil"
	"time"
)

const TimeFormat = "2006-01-02 15:04:05"

type Calender struct {
	AuthorID     string `json:"AuthorID`
	ChannelID    string `json:"ChannelID"`
	Date         string `json:"Date"`
	Description  string `json:"Description"`
	Participants string `json:"Participants"`
	ServerID     string `json:"ServerID"`
	Title        string `json:"Title"`
}

var Events []Calender

func NewCalender(title, description, role, date, serverID, channelID, authorID string) (c *Calender, e error) {
	t, err := time.Parse(TimeFormat, date)
	if err != nil {
		return nil, err
	}

	Events = append(Events, Calender{
		Title:        title,
		Description:  description,
		Participants: role,
		ServerID:     serverID,
		ChannelID:    channelID,
		AuthorID:     authorID,
		Date:         t.Format(TimeFormat),
	})

	c = &Events[len(Events)-1]
	go SaveCalenders()
	return
}

func LoadCalenders() error {
	Events = make([]Calender, 0)
	b, err := ioutil.ReadFile("./events.json")
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, &Events)
	if err != nil {
		return err
	}

	return nil
}

func SaveCalenders() error {
	b, err := json.Marshal(Events)
	if err != nil {
		return err
	}

	return ioutil.WriteFile("./events.json", b, 777)
}
