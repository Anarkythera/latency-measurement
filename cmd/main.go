package main

import (
	"latencychecker/chat"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "latency measurer",
		Usage: "",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "channel",
				Aliases: []string{"c"},
				Usage:   "`NAME` of channel to be created for communication",
				Value:   "testChannel",
			},
			&cli.StringFlag{
				Name:     "api-key",
				EnvVars:  []string{"ABLY_API_KEY"},
				Usage:    "Ably API `KEY` for the channel",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "clientID",
				Aliases: []string{"n"},
				Usage:   "`NAME` to identify the client when sending messages",
				Value:   os.Getenv("HOSTNAME"),
			},
			&cli.IntFlag{
				Name:    "messages",
				Aliases: []string{"m"},
				Usage:   "How many `MESSAGES` to send",
				Value:   5,
			},
			&cli.IntFlag{
				Name:    "wait",
				Aliases: []string{"w"},
				Usage:   "How long to listen for `RESPONSES`",
				Value:   30,
			},
			&cli.IntFlag{
				Name:    "delay",
				Aliases: []string{"d"},
				Usage:   "Delay between sending messages by 'SECONDS'",
				Value:   5,
			},
		},
		Action: func(c *cli.Context) error {
			chat.Start(c.String("api-key"), c.String("channel"), c.String("clientID"), c.Int("messages"), c.Int("delay"), c.Int("wait"))

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
