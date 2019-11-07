package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"gopkg.in/yaml.v2"
)

type container struct {
	image        string
	lastDied     time.Time
	lastAlerted  time.Time
	restartCount int
}

type conf struct {
	AlertDies      int `yaml:"alert_dies"`
	AlertMinutes   int `yaml:"alert_minutes"`
	AlertFrequency int `yaml:"alert_frequency"`
}

func main() {
	configFilePath := os.Args[1]
	c := conf{
		AlertDies:      1,
		AlertMinutes:   5,
		AlertFrequency: 10,
	}
	if configFilePath != "" {
		if _, err := os.Stat(configFilePath); err == nil {
			yamlFile, err := ioutil.ReadFile(configFilePath)
			if err != nil {
				log.Printf("yamlFile.Get err   #%v ", err)
			}
			err = yaml.Unmarshal(yamlFile, &c)
			if err != nil {
				log.Fatalf("Unmarshal: %v", err)
			}
			fmt.Println(c)
		}
	}
	containers := make(map[string]container)
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	cli.NegotiateAPIVersion(ctx)

	filters := filters.NewArgs()
	filters.Add("type", "container")
	filters.Add("event", "die")

	msgs, errs := cli.Events(ctx, types.EventsOptions{"", "", filters})

	for {
		select {
		case err := <-errs:
			fmt.Println(err)
		case msg := <-msgs:
			fmt.Println(msg.Status)
			fmt.Println(msg.ID)
			fmt.Println(msg.From)
			fmt.Println(msg.Type)
			fmt.Println(msg.Action)
			fmt.Println(msg.Actor)
			entry, cont := containers[msg.From]
			if cont {
				duration := time.Now().Sub(entry.lastDied).Minutes()
				if duration < 5 && entry.restartCount > 1 && time.Now().Sub(entry.lastAlerted).Minutes() > 10 {
					fmt.Println("alert!")
					entry.lastAlerted = time.Now()
				}
			}
			if !cont {
				entry = container{image: msg.From}
			}
			entry.lastDied = time.Now()
			entry.restartCount++
			containers[msg.From] = entry
			fmt.Println(entry.image)
			fmt.Println(entry.lastDied)
			fmt.Println(entry.restartCount)
		}
	}
}
