package solver

import (
	"github.com/iainlane/who-dances-what/internal/model"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
)

// This is a wrapper around the C solver. It takes in the data from the model
// and converts it into the format the C solver expects.
// It then converts the output from the C solver back into the format the model
// expects.
func Solve(logger *logrus.Entry, dps []*model.DancerPosition) model.AssignmentSet {
	// Convert the model data into the format the C solver expects
	dancers := make(map[*model.Dancer]rawDancer)
	dancersById := make(map[int]*model.Dancer)
	dances := make(map[*model.Dance]rawDance)
	dancerPositions := make([]rawDancerPosition, 0, len(dps))

	// check if the dancer is already in the map and if not, add it
	for _, dp := range dps {
		dancer := dp.Dancer
		if _, ok := dancers[dancer]; !ok {
			dancers[dancer] = rawDancer{
				Active: dancer.Active,
				ID:     int(dancer.ID),
			}
			dancersById[dancer.ID] = dancer
		}

		dance := dp.Dance
		if _, ok := dances[dance]; !ok {
			positions := make([]rawPosition, 0, len(dance.Positions))
			for _, position := range dance.Positions {
				positions = append(positions, rawPosition{PositionID: int(position.PositionID)})
			}
			dances[dance] = rawDance{
				ID:        int(dance.ID),
				Positions: positions,
			}
		}

		dancerPosition := dp
		dancerPositions = append(dancerPositions, rawDancerPosition{
			DancerID:   int(dancer.ID),
			PositionID: int(dancerPosition.Position.PositionID),
			DanceID:    int(dance.ID),
			Preference: DancePreference(dancerPosition.Preference),
		})
	}

	solver := newCDanceSolver(logger, maps.Values(dancers), maps.Values(dances), dancerPositions)
	defer solver.freeCDanceSolver()
	solution := solver.getPossibleDances()
	defer solution.freeCDanceSolution()

	// Convert the output from the C solver back into the format the model expects
	as := make(model.Assignments)
	dd := make(model.DancesDanced)
	assignments := model.NewAssignmentSet(as, dd)
	for dance, rawDance := range dances {
		danceID := rawDance.ID
		as[dance] = make(map[*model.Position]*model.Dancer)
		if solution.isDancePerformed(dance.ID) {
			dd[dance] = struct{}{}
		}
		for _, position := range dance.Positions {
			positionID := position.PositionID
			idOfDancer := solution.getDancerDancePosition(danceID, positionID)
			dancer := dancersById[idOfDancer]
			as[dance][position] = dancer
		}
	}

	return assignments
}
