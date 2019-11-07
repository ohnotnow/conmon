package main

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

type container struct {
	image        string
	lastDied     time.Time
	restartCount int
}

func main() {
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
				if duration < 5 && entry.restartCount > 1 {
					fmt.Println("alert!")
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
