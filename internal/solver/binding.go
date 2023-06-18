package solver

/*
#cgo CFLAGS: -I../../cppsolver/src/include -I../../cppsolver/build/
#cgo LDFLAGS: -L../../cppsolver/build/ -Wl,-rpath,${SRCDIR}/../../cppsolver/build/ -ldance_solver

#include <stdlib.h>
#include <stdint.h>

#include "dance_solver.h"
*/
import "C"
import (
	"fmt"
	"runtime/cgo"
	"unsafe"

	"github.com/iainlane/who-dances-what/internal/loggerbinding"
	"github.com/sirupsen/logrus"
)

type SolverStatus int

// map the SolverStatus enum from the C library to Go
const (
	SolverStatusUnknown      SolverStatus = C.SolverStatusUnknown
	SolverStatusModelInvalid SolverStatus = C.SolverStatusModelInvalid
	SolverStatusFeasible     SolverStatus = C.SolverStatusFeasible
	SolverStatusInfeasible   SolverStatus = C.SolverStatusInfeasible
	SolverStatusOptimal      SolverStatus = C.SolverStatusOptimal
)

func (s SolverStatus) String() string {
	switch s {
	case SolverStatusUnknown:
		return "Unknown"
	case SolverStatusModelInvalid:
		return "ModelInvalid"
	case SolverStatusFeasible:
		return "Feasible"
	case SolverStatusInfeasible:
		return "Infeasible"
	case SolverStatusOptimal:
		return "Optimal"
	default:
		return fmt.Sprintf("Unknown SolverStatus: %d", s)
	}
}

type DancePreference int

// map the DancePreference enum from the C library to Go
const (
	PreferenceNo        DancePreference = C.PreferenceNo
	PreferenceMaybe     DancePreference = C.PreferenceMaybe
	PreferenceYes       DancePreference = C.PreferenceYes
	PreferenceFavourite DancePreference = C.PreferenceFavourite
)

func (p DancePreference) String() string {
	switch p {
	case PreferenceNo:
		return "No"
	case PreferenceMaybe:
		return "Maybe"
	case PreferenceYes:
		return "Yes"
	case PreferenceFavourite:
		return "Favourite"
	default:
		return fmt.Sprintf("Unknown DancePreference: %d", p)
	}
}

type DancerPositionStatus int

const (
	DancerPositionStatusUnknown DancerPositionStatus = C.DancerPositionStatusUnknown
	DancerPositionStatusNo      DancerPositionStatus = C.DancerPositionStatusNo
	DancerPositionStatusYes     DancerPositionStatus = C.DancerPositionStatusYes
)

func (dps DancerPositionStatus) String() string {
	switch dps {
	case DancerPositionStatusUnknown:
		return "Unknown"
	case DancerPositionStatusNo:
		return "No"
	case DancerPositionStatusYes:
		return "Yes"
	default:
		return fmt.Sprintf("Unknown DancerPositionStatus: %d", dps)
	}
}

// The raw structs are used to convert the Go structs to C structs and back
type rawDancer struct {
	ID     int
	Active bool
}

type rawPosition struct {
	PositionID int
}

type rawDance struct {
	ID        int
	Positions []rawPosition
}

type rawDancerPosition struct {
	DancerID   int
	PositionID int
	DanceID    int
	Preference DancePreference
}

func toCDancer(d rawDancer) C.Dancer {
	active := 0
	if d.Active {
		active = 1
	}

	return C.Dancer{
		id:     C.int(d.ID),
		active: C.int(active),
	}
}

func toCDance(d rawDance) C.Dance {
	positions := C.calloc(C.size_t(len(d.Positions)), C.sizeof_Position)

	positionSlice := (*[1<<30 - 1]C.Position)(positions)
	for i, position := range d.Positions {
		positionSlice[i] = C.Position{position_id: C.int(position.PositionID)}
	}

	return C.Dance{
		id:            C.int(d.ID),
		positions:     &positionSlice[0],
		num_positions: C.int(len(d.Positions)),
	}
}

func toCDancerPosition(dp rawDancerPosition) C.DancerPosition {
	return C.DancerPosition{
		dancer_id:   C.int(dp.DancerID),
		position_id: C.int(dp.PositionID),
		dance_id:    C.int(dp.DanceID),
		preference:  C.DancePreference(dp.Preference),
	}
}

func freeCDancePositions(d *C.Dance) {
	C.free(unsafe.Pointer(d.positions))
	d.positions = nil
}

type cDanceSolver struct {
	loggerHandle cgo.Handle
	solver       *C.Solver

	dancers         unsafe.Pointer
	dances          unsafe.Pointer
	num_dances      int
	dancerPositions unsafe.Pointer
}

func newCDanceSolver(logger *logrus.Entry, dancers []rawDancer, dances []rawDance, dancer_positions []rawDancerPosition) cDanceSolver {
	handle := cgo.NewHandle(logger)

	cDancers := C.calloc(C.size_t(len(dancers)), C.sizeof_Dancer)
	dancerSlice := (*[1<<30 - 1]C.Dancer)(cDancers)
	for i, d := range dancers {
		dancerSlice[i] = toCDancer(d)
	}

	cDances := C.calloc(C.size_t(len(dances)), C.sizeof_Dance)
	danceSlice := (*[1<<30 - 1]C.Dance)(cDances)
	for i, d := range dances {
		danceSlice[i] = toCDance(d)
	}

	cDancerPositions := C.calloc(C.size_t(len(dancer_positions)), C.sizeof_DancerPosition)
	dancerPositionSlice := (*[1<<30 - 1]C.DancerPosition)(cDancerPositions)
	for i, dp := range dancer_positions {
		dancerPositionSlice[i] = toCDancerPosition(dp)
	}

	lb := loggerbinding.PopulateLogger(handle)

	solver := C.dance_solver_new_with_logger(
		(*C.logger)(unsafe.Pointer(lb)),
		&dancerSlice[0],
		C.int(len(dancers)),
		&danceSlice[0],
		C.int(len(dances)),
		&dancerPositionSlice[0],
		C.int(len(dancer_positions)),
	)

	return cDanceSolver{handle, solver, cDancers, cDances, len(dances), cDancerPositions}
}

func (solver cDanceSolver) freeCDanceSolver() {
	solver.loggerHandle.Delete()
	C.free(solver.dancers)
	for i := 0; i < solver.num_dances; i++ {
		dance := (*C.Dance)(unsafe.Pointer(uintptr(solver.dances) + uintptr(i)*C.sizeof_Dance))
		freeCDancePositions(dance)
	}
	C.free(solver.dances)
	C.free(solver.dancerPositions)
	C.free_dance_solver(solver.solver)
}

type cDanceSolution struct {
	num_assignments int
	num_dances      int
	solution        *C.DanceSolution
	status          SolverStatus
}

func (solver cDanceSolver) getPossibleDances() cDanceSolution {
	solution := C.get_possible_dances(solver.solver)
	return cDanceSolution{
		num_assignments: int(solution.num_assignments),
		num_dances:      int(solution.num_dances),
		solution:        solution,
		status:          SolverStatus(solution.status),
	}
}

func (solution cDanceSolution) freeCDanceSolution() {
	C.free_dance_solution(solution.solution)
}

func (solution cDanceSolution) getDancerDancePosition(dance_id int, position_id int) int {
	position := C.get_dancer_dance_position(solution.solution, C.int(dance_id), C.int(position_id))

	return int(position)
}

func (solution cDanceSolution) isDancePerformed(dance_id int) bool {
	performed := C.is_dance_performed(solution.solution, C.int(dance_id))

	return performed == 1
}
