package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	logger := log.New()
	logger.SetLevel(log.InfoLevel)

	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "db",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "log-level",
				Value: "info",
				Action: func(c *cli.Context, level string) error {
					l, err := log.ParseLevel(level)
					if err != nil {
						return err
					}
					logger.SetLevel(l)
					return nil
				},
			},
		},
		Commands: []*cli.Command{
			listDances(logger.WithField("command", "list-dances")),
			listActiveDancers(logger.WithField("command", "list-active-dancers")),
			danceSet(logger.WithField("command", "dance-set")),
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Fatal(err)
	}
}
