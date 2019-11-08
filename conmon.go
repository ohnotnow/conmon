package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/ashwanthkumar/slack-go-webhook"
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
	AlertDies      int    `yaml:"alert_dies"`
	AlertMinutes   int    `yaml:"alert_minutes"`
	AlertFrequency int    `yaml:"alert_frequency"`
	SlackWebhook   string `yaml:"slack_webhook"`
	SlackChannel   string `yaml:"slack_channel"`
	Hostname       string `yaml:"hostname"`
	Proxy          string `yaml:"proxy"`
}

func main() {
	hostname, err := os.Hostname()
	confPtr := flag.String("config", "", "Path to config file")
	hostnamePtr := flag.String("hostname", hostname, "Path to config file")
	proxyPtr := flag.String("proxy", "", "Path to config file")
	slackHookPtr := flag.String("slackhook", "", "Slack webhook url")
	slackChannelPtr := flag.String("slackchannel", "", "Slack channel")
	flag.Parse()
	c := conf{
		AlertDies:      1,
		AlertMinutes:   5,
		AlertFrequency: 10,
		SlackWebhook:   *slackHookPtr,
		SlackChannel:   *slackChannelPtr,
		Hostname:       *hostnamePtr,
		Proxy:          *proxyPtr,
	}
	if *confPtr != "" {
		if _, err := os.Stat(*confPtr); err != nil {
			log.Fatalf("Config file not found : %s", *confPtr)
		}

		yamlFile, err := ioutil.ReadFile(*confPtr)
		if err != nil {
			log.Printf("yamlFile.Get err   #%v ", err)
		}
		err = yaml.Unmarshal(yamlFile, &c)
		if err != nil {
			log.Fatalf("Unmarshal: %v", err)
		}
	}

	log.Printf("Starting... Config : %+v\n", c)

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
			log.Println(err)
		case msg := <-msgs:
			logEntry := fmt.Sprintf("ID %s image %s status %s name %s", msg.ID, msg.From, msg.Status, msg.Actor.Attributes["name"])
			log.Println(logEntry)
			entry, cont := containers[msg.From]
			if cont {
				entry.restartCount++
				duration := time.Now().Sub(entry.lastDied).Minutes()
				if duration < float64(c.AlertMinutes) && entry.restartCount > c.AlertDies && time.Now().Sub(entry.lastAlerted).Minutes() > float64(c.AlertFrequency) {
					log.Printf("alert for %s\n", msg.From)
					if c.SlackWebhook != "" {
						err := SendSlackNotification(c, logEntry)
						if err != nil {
							log.Printf("Error sending slack notification : %s", err)
						}
					}
					entry.lastAlerted = time.Now()
					entry.restartCount = 0
				}
			} else {
				entry = container{image: msg.From, restartCount: 0}
			}
			entry.lastDied = time.Now()
			containers[msg.From] = entry
		}
	}
}

func SendSlackNotification(config conf, msg string) []error {
	payload := slack.Payload{
		Text:      config.Hostname + " Container repeatedly dying : " + msg,
		Username:  "conmon",
		Channel:   config.SlackChannel,
		IconEmoji: ":monkey_face:",
	}

	err := slack.Send(config.SlackWebhook, config.Proxy, payload)
	if err != nil {
		return err
	}
	return nil
}
