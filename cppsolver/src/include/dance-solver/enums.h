// don't allow this to be included directly, users must include either
// dance_solver.h or dance_solver.hpp
#ifndef DANCE_SOLVER_INTERNAL_INCLUDE
#error "Please include dance_solver.h or dance_solver.hpp instead of enums.h"
#endif

typedef enum
{
    PreferenceNo = 0,
    PreferenceMaybe = 1,
    PreferenceYes = 2,
    PreferenceFavourite = 3,
} DancePreference;

typedef enum
{
    SolverStatusUnknown = 0,
    SolverStatusModelInvalid = 1,
    SolverStatusFeasible = 2,
    SolverStatusInfeasible = 3,
    SolverStatusOptimal = 4,
} SolverStatus;

typedef enum
{
    DancerPositionStatusUnknown = 0,
    DancerPositionStatusNo = 1,
    DancerPositionStatusYes = 2,
} DancerPositionStatus;
