package main

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/iainlane/who-dances-what/internal/model"
)

func listDances(logger *logrus.Entry) *cli.Command {
	return &cli.Command{
		Name:  "list-dances",
		Usage: "List all dances",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "no-colour",
				Aliases: []string{"n", "no-color"},
				Usage:   "Disable colour output",
				Action: func(c *cli.Context, disable bool) error {
					color.NoColor = disable
					return nil
				},
			},
		},
		Action: func(c *cli.Context) error { return doListDances(c, logger) },
	}
}

func doListDances(c *cli.Context, logger *logrus.Entry) error {
	m, err := model.NewModel(c.String("db"), logger)
	if err != nil {
		return err
	}

	// make sprintf functions for each preference
	yellow := color.New(color.FgYellow).SprintfFunc()
	green := color.New(color.FgGreen).SprintfFunc()
	magenta := color.New(color.FgMagenta, color.Bold).SprintfFunc()
	red := color.New(color.FgRed).SprintfFunc()

	// get the positions for the dancers
	dances, err := m.FetchDances()
	if err != nil {
		return err
	}

	var sb strings.Builder

	for _, dance := range dances {
		sb.WriteString("Dance: ")
		sb.WriteString(dance.Name)
		sb.WriteString("\n")

		for _, position := range dance.Positions {
			var dps []*model.DancerPosition
			for _, dp := range position.DancerPositions {
				if dp.Preference == model.PreferenceNo {
					continue
				}

				dps = append(dps, dp)
			}

			if len(dps) == 0 {
				continue
			}

			sb.WriteString(" ")
			sb.WriteString(position.Name)
			sb.WriteString(": ")

			for i, pref := range dps {
				if i > 0 {
					sb.WriteString(", ")
				}

				var sprintf func(format string, a ...interface{}) string
				switch pref.Preference {
				case model.PreferenceFavourite:
					sprintf = magenta
				case model.PreferenceYes:
					sprintf = green
				case model.PreferenceMaybe:
					sprintf = yellow
				case model.PreferenceNo:
					sprintf = red
				}

				sb.WriteString(sprintf(pref.Dancer.Name))
			}

			sb.WriteString("\n")
		}
	}

	fmt.Print(sb.String())

	return nil
}
