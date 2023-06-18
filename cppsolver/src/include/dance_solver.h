// A C wrapper for the C++ DanceSolver class, which is defined in
// `dance_solver.hpp`.
#pragma once

#include <stdint.h>

#ifdef __cplusplus
extern "C"
{
    namespace dance_solver_c_api
    {
#endif

// Shared enums
#define DANCE_SOLVER_INTERNAL_INCLUDE
#include "dance-solver/logger.h"
#include "dance-solver/enums.h"
#undef DANCE_SOLVER_INTERNAL_INCLUDE

        typedef struct
        {
            int id;
            int active;
        } Dancer;

        typedef struct
        {
            int position_id;
        } Position;

        typedef struct
        {
            int id;
            Position *positions;
            int num_positions;
        } Dance;

        typedef struct
        {
            int dancer_id;
            int position_id;
            int dance_id;
            DancePreference preference;
            DancerPositionStatus status;
        } DancerPosition;

        typedef struct
        {
            int dance_id;
            int position_id;
            int64_t dancer_id;
        } PositionSolution;

        typedef struct DanceSolutionPriv DanceSolutionPriv;

        typedef struct
        {
            SolverStatus status;
            int num_assignments;
            int num_dances;
            DanceSolutionPriv *priv;
        } DanceSolution;

        typedef struct Solver Solver;

        Solver *dance_solver_new_with_logger(
            logger *l,
            Dancer *dancers, int num_dancers,
            Dance *dances, int num_dances,
            DancerPosition *dancer_positions, int num_dancer_positions);
        void free_dance_solver(Solver *solver);
        DanceSolution *get_possible_dances(Solver *solver);
        void free_dance_solution(DanceSolution *solution);
        int get_dancer_dance_position(DanceSolution *solution, int dance_id, int position_id);
        int is_dance_performed(DanceSolution *solution, int dance_id);

#ifdef __cplusplus
    } // namespace dance_solver_c_api
} // extern "C"
#endif
