package solver

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestOneDanceOnePositionOneDancer(t *testing.T) {
	t.Parallel()

	require := require.New(t)

	dancers := []rawDancer{{1, true}}
	dances := []rawDance{{1, []rawPosition{{1}}}}
	dancer_positions := []rawDancerPosition{{1, 1, 1, PreferenceYes}}

	solver := newCDanceSolver(logrus.WithField("test-name", t.Name()), dancers, dances, dancer_positions)
	defer solver.freeCDanceSolver()

	solution := solver.getPossibleDances()
	defer solution.freeCDanceSolution()

	require.Equalf(SolverStatusOptimal, solution.status, "Expected status to be SolverStatusOptimal, got %s", solution.status)
	require.Equal(1, solution.num_assignments, "Expected number of assignments to be 1")

	pos := solution.getDancerDancePosition(1, 1)
	require.Equal(1, pos, "Expected dancer 1 to be assigned to position 1")

	require.Equal(solution.num_dances, 1)
	require.True(solution.isDancePerformed(1))
}

func Test_OneDanceWithOnePositionAndOneDancer(t *testing.T) {
	t.Parallel()

	require := require.New(t)

	dancers := []rawDancer{{1, true}}
	dances := []rawDance{{1, []rawPosition{{1}}}}
	dancer_positions := []rawDancerPosition{
		{1, 1, 1, PreferenceYes},
	}

	solver := newCDanceSolver(logrus.WithField("test-name", t.Name()), dancers, dances, dancer_positions)
	defer solver.freeCDanceSolver()

	solution := solver.getPossibleDances()
	defer solution.freeCDanceSolution()

	require.Equalf(SolverStatusOptimal, solution.status, "Expected status to be SolverStatusOptimal, got %s", solution.status)
	require.Equal(1, solution.num_assignments, "Expected number of assignments to be 1")

	pos := solution.getDancerDancePosition(1, 1)
	require.Equal(1, pos, "Expected dancer 1 to be assigned to position 1")

	require.Equal(solution.num_dances, 1)
	require.True(solution.isDancePerformed(1))
}

func Test_InactiveDancersAreSkipped(t *testing.T) {
	t.Parallel()

	require := require.New(t)

	dancers := []rawDancer{
		{1, false},
		{2, false},
		{3, false},
		{4, false},
	}
	dances := []rawDance{{1, []rawPosition{{1}}}}
	dancer_positions := []rawDancerPosition{
		{1, 1, 1, PreferenceYes},
		{2, 1, 1, PreferenceYes},
		{3, 1, 1, PreferenceYes},
		{4, 1, 1, PreferenceYes},
	}

	solver := newCDanceSolver(logrus.WithField("test-name", t.Name()), dancers, dances, dancer_positions)
	defer solver.freeCDanceSolver()

	solution := solver.getPossibleDances()
	defer solution.freeCDanceSolution()

	require.Equalf(SolverStatusInfeasible, solution.status, "Expected status to be SolverStatusInfeasible, got %s", solution.status)
	require.Equal(0, solution.num_assignments, "Expected number of assignments to be 0")

	require.Equal(solution.num_dances, 1)
	require.False(solution.isDancePerformed(1))
}

func Test_OneDanceWithOnePositionAndFourDancers(t *testing.T) {
	t.Parallel()

	require := require.New(t)

	dancers := []rawDancer{
		{1, true},
		{2, true},
		{3, true},
		{4, true},
	}
	dances := []rawDance{{1, []rawPosition{{1}}}}
	dancer_positions := []rawDancerPosition{
		{1, 1, 1, PreferenceYes},
		{2, 1, 1, PreferenceYes},
		{3, 1, 1, PreferenceYes},
		{4, 1, 1, PreferenceYes},
	}

	solver := newCDanceSolver(logrus.WithField("test-name", t.Name()), dancers, dances, dancer_positions)
	defer solver.freeCDanceSolver()

	solution := solver.getPossibleDances()
	defer solution.freeCDanceSolution()

	require.Equalf(SolverStatusOptimal, solution.status, "Expected status to be SolverStatusOptimal, got %s", solution.status)
	require.Equal(1, solution.num_assignments, "Expected number of assignments to be 1")

	pos := solution.getDancerDancePosition(1, 1)
	require.True(pos >= 1 && pos <= 4, "Expected dancer 1 to be assigned to position 1, 2, 3, or 4")

	require.Equal(solution.num_dances, 1)
	require.True(solution.isDancePerformed(1))
}

func Test_TwoDancesOneDancer(t *testing.T) {
	t.Parallel()

	require := require.New(t)

	dancers := []rawDancer{{1, true}}
	dances := []rawDance{
		{1, []rawPosition{{1}}},
		{2, []rawPosition{{1}}},
	}
	dancer_positions := []rawDancerPosition{
		{1, 1, 1, PreferenceYes},
		{1, 1, 2, PreferenceYes},
	}

	solver := newCDanceSolver(logrus.WithField("test-name", t.Name()), dancers, dances, dancer_positions)
	defer solver.freeCDanceSolver()

	solution := solver.getPossibleDances()
	defer solution.freeCDanceSolution()

	require.Equalf(SolverStatusOptimal, solution.status, "Expected status to be SolverStatusOptimal, got %s", solution.status)
	require.Equal(2, solution.num_assignments, "Expected number of assignments to be 2")

	pos1 := solution.getDancerDancePosition(1, 1)
	pos2 := solution.getDancerDancePosition(2, 1)

	require.Equal(1, pos1, "Expected dancer 1 to be assigned to position 1 in dance 1")
	require.Equal(1, pos2, "Expected dancer 1 to be assigned to position 1 in dance 2")

	require.Equal(solution.num_dances, 2)
	require.True(solution.isDancePerformed(1))
	require.True(solution.isDancePerformed(2))
}

func Test_OneDanceNobodyCanDanceIt(t *testing.T) {
	t.Parallel()

	require := require.New(t)

	dancers := []rawDancer{{1, true}}
	dances := []rawDance{{1, []rawPosition{{1}}}}
	dancer_positions := []rawDancerPosition{}

	solver := newCDanceSolver(logrus.WithField("test-name", t.Name()), dancers, dances, dancer_positions)
	defer solver.freeCDanceSolver()

	solution := solver.getPossibleDances()
	defer solution.freeCDanceSolution()

	require.Equalf(SolverStatusInfeasible, solution.status, "Expected status to be SolverStatusInfeasible, got %s", solution.status)
	require.Equal(0, solution.num_assignments, "Expected number of assignments to be 0")

	require.Equal(solution.num_dances, 1)
	require.False(solution.isDancePerformed(1))
}

func Test_OneDanceTwoPositionsTwoDancers(t *testing.T) {
	t.Parallel()

	require := require.New(t)

	dancers := []rawDancer{{1, true}, {2, true}}
	dances := []rawDance{{1, []rawPosition{{1}, {2}}}}
	dancer_positions := []rawDancerPosition{
		{1, 1, 1, PreferenceYes},
		{2, 2, 1, PreferenceYes},
	}

	solver := newCDanceSolver(logrus.WithField("test-name", t.Name()), dancers, dances, dancer_positions)
	defer solver.freeCDanceSolver()

	solution := solver.getPossibleDances()
	defer solution.freeCDanceSolution()

	require.Equalf(SolverStatusOptimal, solution.status, "Expected status to be SolverStatusOptimal, got %s", solution.status)
	require.Equal(2, solution.num_assignments, "Expected number of assignments to be 2")

	pos1 := solution.getDancerDancePosition(1, 1)
	pos2 := solution.getDancerDancePosition(1, 2)

	require.Equal(1, pos1, "Expected dancer 1 to be assigned to position 1")
	require.Equal(2, pos2, "Expected dancer 2 to be assigned to position 2")

	require.Equal(solution.num_dances, 1)
	require.True(solution.isDancePerformed(1))
}

func Test_OneDanceTwoPositionsTwoDancers_CantDanceBothPositions(t *testing.T) {
	t.Parallel()

	require := require.New(t)

	dancers := []rawDancer{{1, true}, {2, true}}
	dances := []rawDance{{1, []rawPosition{{1}, {2}}}}
	dancer_positions := []rawDancerPosition{
		{1, 1, 1, PreferenceYes},
		{2, 1, 1, PreferenceYes},
	}

	solver := newCDanceSolver(logrus.WithField("test-name", t.Name()), dancers, dances, dancer_positions)
	defer solver.freeCDanceSolver()

	solution := solver.getPossibleDances()
	defer solution.freeCDanceSolution()

	require.Equalf(SolverStatusInfeasible, solution.status, "Expected status to be SolverStatusInfeasible, got %s", solution.status)
	require.Equal(0, solution.num_assignments, "Expected number of assignments to be 0")

	require.Equal(solution.num_dances, 1)
	require.False(solution.isDancePerformed(1))
}

func Test_TwoDances_OneCanBeDancedAndTheOtherCant(t *testing.T) {
	t.Parallel()

	require := require.New(t)

	dancers := []rawDancer{{ID: 1, Active: true}}
	dances := []rawDance{
		{1, []rawPosition{{1}}},
		{2, []rawPosition{{1}, {2}}},
	}
	dancer_positions := []rawDancerPosition{
		{1, 1, 1, PreferenceYes},
	}

	solver := newCDanceSolver(logrus.WithField("test-name", t.Name()), dancers, dances, dancer_positions)
	defer solver.freeCDanceSolver()

	solution := solver.getPossibleDances()
	defer solution.freeCDanceSolution()

	require.Equalf(SolverStatusOptimal, solution.status, "Expected status to be SolverStatusOptimal, got %s", solution.status)
	require.Equal(1, solution.num_assignments, "Expected number of assignments to be 1")

	pos := solution.getDancerDancePosition(1, 1)
	require.Equal(1, pos, "Expected dancer 1 to be assigned to dance 1")

	require.Equal(solution.num_dances, 2)
	require.True(solution.isDancePerformed(1))
}

func Test_TwoDancesTwoDancers_BothShouldBeGivenOneOfThem(t *testing.T) {
	t.Parallel()

	require := require.New(t)

	dancers := []rawDancer{{1, true}, {2, true}}
	dances := []rawDance{
		{1, []rawPosition{{1}}},
		{2, []rawPosition{{1}}},
	}
	dancer_positions := []rawDancerPosition{
		{1, 1, 1, PreferenceYes},
		{1, 1, 2, PreferenceYes},
		{2, 1, 1, PreferenceYes},
		{2, 1, 2, PreferenceYes},
	}

	solver := newCDanceSolver(logrus.WithField("test-name", t.Name()), dancers, dances, dancer_positions)
	defer solver.freeCDanceSolver()

	solution := solver.getPossibleDances()
	defer solution.freeCDanceSolution()

	require.Equalf(SolverStatusOptimal, solution.status, "Expected status to be SolverStatusOptimal, got %s", solution.status)
	require.Equal(2, solution.num_assignments, "Expected number of assignments to be 2")

	pos1 := solution.getDancerDancePosition(1, 1)
	pos2 := solution.getDancerDancePosition(2, 1)

	require.True(pos1 == 1 || pos1 == 2, "Expected dancer 1 to be assigned to dance 1 or 2")
	exp2 := 2
	if pos1 == 2 {
		exp2 = 1
	}
	require.Equalf(exp2, pos2, "Expected dancer 1 to be assigned to dance %d, got: %d", exp2, pos2)

	require.Equal(solution.num_dances, 2)
	require.True(solution.isDancePerformed(1))
}

func Test_PreferencesAreTakenIntoAccount_SimpleCase(t *testing.T) {
	t.Parallel()

	require := require.New(t)

	dancers := []rawDancer{{1, true}, {2, true}}
	dances := []rawDance{{1, []rawPosition{{1}}}}
	dancer_positions := []rawDancerPosition{
		{1, 1, 1, PreferenceYes},
		{2, 1, 1, PreferenceFavourite},
	}

	solver := newCDanceSolver(logrus.WithField("test-name", t.Name()), dancers, dances, dancer_positions)
	defer solver.freeCDanceSolver()

	solution := solver.getPossibleDances()
	defer solution.freeCDanceSolution()

	require.Equalf(SolverStatusOptimal, solution.status, "Expected status to be SolverStatusOptimal, got %s", solution.status)
	require.Equal(1, solution.num_assignments, "Expected number of assignments to be 1")

	pos := solution.getDancerDancePosition(1, 1)

	// Dancer 2 is preferred for the dance, so they should be the one assigned to it
	require.Equalf(2, pos, "Expected dancer 2 to be assigned to dance 1, got: %d", pos)

	require.Equal(solution.num_dances, 1)
	require.True(solution.isDancePerformed(1))
}
