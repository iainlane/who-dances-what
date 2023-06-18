package main

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/iainlane/who-dances-what/internal/model"
	"github.com/iainlane/who-dances-what/internal/solver"
)

type danceSetGenerator struct {
	logger      *logrus.Entry
	dancerNames []string
}

func danceSet(logger *log.Entry) *cli.Command {
	generator := danceSetGenerator{
		logger: logger,
	}

	return &cli.Command{
		Name:  "dance-set",
		Usage: "Generate a dance set given a list of dancers",
		Before: func(c *cli.Context) error {
			return generator.handleCommandLineParameters(c)
		},
		Action: func(c *cli.Context) error { return generator.doGenerateDanceSet(c) },
	}
}

func (g *danceSetGenerator) handleCommandLineParameters(c *cli.Context) error {
	// dancer names should be in a positional argument
	dancerNames := c.Args().Slice()

	// check that the slice is not empty
	if len(dancerNames) == 0 {
		return cli.Exit("No dancers specified", 1)
	}

	g.dancerNames = dancerNames

	return nil
}

func (g *danceSetGenerator) doGenerateDanceSet(c *cli.Context) error {
	m, err := model.NewModel(c.String("db"), g.logger)
	if err != nil {
		return err
	}

	dancers, err := m.FetchDancersByName(g.dancerNames)
	if err != nil {
		return err
	}

	dances, positions, err := m.FetchDancerPositionsForDancers(dancers)
	if err != nil {
		return err
	}

	set := solver.Solve(g.logger, positions)
	if set.NumDancesDanced() == 0 {
		fmt.Println("Can't dance any dances")
		return nil
	}

	var sb strings.Builder

	for _, dance := range dances {
		if !dance.IsDanced(set) {
			g.logger.WithField("dance", dance.Name).Debug("not danced")
			continue
		}

		sb.WriteString(dance.Name)
		sb.WriteString("\n")

		for _, position := range dance.Positions {
			sb.WriteString(position.Name)
			sb.WriteString(": ")
			sb.WriteString(set.DancerFor(dance, position).Name)
			sb.WriteString("\n")
		}
	}

	fmt.Print(sb.String())

	return nil
}
