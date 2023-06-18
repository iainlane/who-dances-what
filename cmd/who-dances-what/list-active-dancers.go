package main

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/iainlane/who-dances-what/internal/model"
)

func listActiveDancers(logger *logrus.Entry) *cli.Command {
	return &cli.Command{
		Name:   "list-active-dancers",
		Usage:  "List all active dancers",
		Action: func(c *cli.Context) error { return doListActiveDancers(c, logger) },
	}
}

func doListActiveDancers(c *cli.Context, entry *logrus.Entry) error {
	m, err := model.NewModel(c.String("db"), entry)
	if err != nil {
		return err
	}

	dancers, err := m.FetchDancers()
	if err != nil {
		return err
	}

	// build a comma-separated list of dancer names
	var sb strings.Builder
	for _, dancer := range dancers {
		if !dancer.Active {
			continue
		}

		if sb.Len() > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(dancer.Name)
		var roleEmoji string
		switch dancer.Type {
		case model.RoleDancer:
			roleEmoji = "💃"
		case model.RoleMusician:
			roleEmoji = "🎵"
		case model.RoleBoth:
			roleEmoji = "💃🎵"
		}
		sb.WriteString(" (" + roleEmoji + ")")
	}

	fmt.Println(sb.String())

	return nil
}
