package solver

import (
	"testing"

	"github.com/iainlane/who-dances-what/internal/model"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestSolver(t *testing.T) {
	dancer := &model.Dancer{
		ID:     1,
		Name:   "Loner",
		Active: true,
	}
	dance := &model.Dance{
		ID:   1,
		Name: "Loner's Dance",
		Positions: []*model.Position{
			{
				PositionID: 1,
				Name:       "Loner's Position",
			},
		},
	}
	dancerPosition := model.DancerPosition{
		Dancer:     dancer,
		Position:   dance.Positions[0],
		Dance:      dance,
		Preference: model.PreferenceYes,
	}

	set := Solve(logrus.WithField("test-name", t.Name()), []*model.DancerPosition{&dancerPosition})
	require.Equal(t, 1, set.NumDancesDanced())
	require.True(t, dance.IsDanced(set))
	require.Equal(t, dancer, set.DancerFor(dance, dance.Positions[0]))
}
