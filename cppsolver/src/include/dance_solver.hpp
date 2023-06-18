#pragma once

#include <map>
#include <memory>
#include <ranges>
#include <vector>

#define DANCE_SOLVER_INTERNAL_INCLUDE
#include "dance-solver/logger.hpp"
// Shared enums
#include "dance-solver/enums.h"
#undef DANCE_SOLVER_INTERNAL_INCLUDE

// C wrapper
#include "dance_solver.h"

// Weights
#define NUM_DANCES_PERFORMED_WEIGHT 1
#define FAIRNESS_WEIGHT 1

#define PREFERENCE_MAYBE_WEIGHT 1
#define PREFERENCE_YES_WEIGHT 2
#define PREFERENCE_FAVOURITE_WEIGHT 3

struct Dancer
{
    int ID;
    bool Active;

    Dancer() = default;
    Dancer(int id, bool active);

    Dancer(const Dancer &other) = default;
    Dancer &operator=(const Dancer &other) = default;

    Dancer(const dance_solver_c_api::Dancer &dancer);
    Dancer &operator=(const dance_solver_c_api::Dancer &dancer);
};

struct Position
{
    int PositionID;

    Position() = default;
    Position(int position_id);

    Position(const Position &other) = default;
    Position &operator=(const Position &other) = default;

    Position(const dance_solver_c_api::Position &position);
    Position &operator=(const dance_solver_c_api::Position &position);
};

struct Dance
{
    int ID;
    std::vector<Position> Positions;

    Dance() = default;
    Dance(int id, std::vector<Position> positions);

    Dance(const Dance &other) = default;
    Dance &operator=(const Dance &other) = default;

    Dance(const dance_solver_c_api::Dance &dance);
    Dance &operator=(const dance_solver_c_api::Dance &dance);
};

struct DancerPosition
{
    int DancerID;
    int PositionID;
    int DanceID;
    DancePreference Preference;

    DancerPosition() = default;
    DancerPosition(int dancer_id, int position_id, int dance_id, DancePreference preference);

    DancerPosition(const DancerPosition &other) = default;
    DancerPosition &operator=(const DancerPosition &other) = default;

    DancerPosition(const dance_solver_c_api::DancerPosition &dancer_position);
    DancerPosition &operator=(const dance_solver_c_api::DancerPosition &dancer_position);
};

struct PositionSolution
{
    int dance_id;
    int PositionID;
    int64_t DancerID;
};

// Solver
class DanceSolver
{
public:
    typedef int DancerID;
    typedef int DanceID;
    typedef int PositionID;

    // dance_id -> position_id -> dancer_id
    typedef std::map<DanceID, std::map<PositionID, DanceID>> SolutionAssignment;

    // dance_id -> bool
    typedef std::map<DanceID, bool> DancesPerformed;

    struct DanceSolution
    {
        const SolverStatus status;
        const int num_assignments;
        const DancesPerformed dance_performed;
        const SolutionAssignment assignment;
    };

    DanceSolver(
        logger *l,
        std::vector<Dancer> &dancers,
        std::vector<Dance> &dances,
        std::vector<DancerPosition> &dancer_positions);
    ~DanceSolver();
    const DanceSolution GetPossibleDances();

private:
    // hide the or-tools dependency
    class DanceSolverImpl;
    std::unique_ptr<DanceSolverImpl> pimpl_;
};
