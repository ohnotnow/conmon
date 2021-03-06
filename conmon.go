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
	c := BuildConfig()

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
			panic(err)
		case msg := <-msgs:
			logEntry := fmt.Sprintf("ID %s image %s status %s name %s", msg.ID, msg.From, msg.Status, msg.Actor.Attributes["name"])
			log.Println(logEntry)
			entry, cont := containers[msg.From]
			if cont {
				entry.restartCount++
				if ShouldAlert(c, entry) {
					entry = Alert(c, entry, logEntry)
				}
			} else {
				entry = container{image: msg.From, restartCount: 0}
			}
			entry.lastDied = time.Now()
			containers[msg.From] = entry
		}
	}
}

func Alert(c conf, entry container, logEntry string) container {
	log.Printf("alert for %s\n", entry.image)
	if c.SlackWebhook != "" {
		err := SendSlackNotification(c, logEntry)
		if err != nil {
			log.Printf("Error sending slack notification : %s", err)
		}
	}
	entry.lastAlerted = time.Now()
	entry.restartCount = 0
	return entry
}

func BuildConfig() conf {
	hostname, _ := os.Hostname()
	confPtr := flag.String("config", "", "Path to config file")
	hostnamePtr := flag.String("hostname", hostname, "Hostname to use when pinging slack")
	proxyPtr := flag.String("proxy", "", "HTTP proxy url")
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
	c = ParseConfigFile(*confPtr, c)
	return c
}

func ParseConfigFile(confFile string, c conf) conf {
	if confFile == "" {
		return c
	}
	if _, err := os.Stat(confFile); err != nil {
		log.Fatalf("Config file not found : %s", confFile)
	}

	yamlFile, err := ioutil.ReadFile(confFile)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	return c
}

func ShouldAlert(c conf, entry container) bool {
	duration := time.Now().Sub(entry.lastDied).Minutes()
	return duration < float64(c.AlertMinutes) && entry.restartCount > c.AlertDies && time.Now().Sub(entry.lastAlerted).Minutes() > float64(c.AlertFrequency)
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
