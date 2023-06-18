#include <catch2/catch.hpp>

#include "dance_solver.hpp"
#include "testlogger.h"

TEST_CASE("One dance with one position and one dancer", "[dance_solver]")
{
    std::vector<Dancer> dancers = {{1, true}};
    std::vector<Dance> dances = {{1, {{1}}}};
    std::vector<DancerPosition> dancer_positions = {
        {1, 1, 1, PreferenceYes}};

    auto logger = new_test_logger();
    DanceSolver solver(logger, dancers, dances, dancer_positions);

    auto solution = solver.GetPossibleDances();
    auto assignment = solution.assignment;

    REQUIRE(solution.status == SolverStatus::SolverStatusOptimal);

    REQUIRE(assignment[1][1] == 1);

    free_test_logger(logger);
}

TEST_CASE("Inactive dancers are skipped", "[dance_solver]")
{
    std::vector<Dancer> dancers = {
        {1, false},
        {2, false},
        {3, false},
        {4, false}};
    std::vector<Dance> dances = {{1, {{1}}}};
    std::vector<DancerPosition> dancer_positions = {
        {1, 1, 1, PreferenceYes},
        {2, 1, 1, PreferenceYes},
        {3, 1, 1, PreferenceYes},
        {4, 1, 1, PreferenceYes}};

    auto logger = new_test_logger();
    DanceSolver solver(logger, dancers, dances, dancer_positions);

    auto solution = solver.GetPossibleDances();
    auto assignment = solution.assignment;

    REQUIRE(solution.status == SolverStatus::SolverStatusInfeasible);
    REQUIRE(assignment.size() == 0);
    REQUIRE(solution.num_assignments == 0);

    free_test_logger(logger);
}

TEST_CASE("One dance with one position and four dancers", "[dance_solver]")
{
    std::vector<Dancer> dancers = {
        {1, true},
        {2, true},
        {3, true},
        {4, true}};
    std::vector<Dance> dances = {{1, {{1}}}};
    std::vector<DancerPosition> dancer_positions = {
        {1, 1, 1, PreferenceYes},
        {2, 1, 1, PreferenceYes},
        {3, 1, 1, PreferenceYes},
        {4, 1, 1, PreferenceYes}};

    std::vector<int> dancer_ids = {1, 2, 3, 4};

    auto logger = new_test_logger();
    DanceSolver solver(logger, dancers, dances, dancer_positions);

    auto solution = solver.GetPossibleDances();
    REQUIRE(solution.status == SolverStatus::SolverStatusOptimal);

    auto assignment = solution.assignment;
    REQUIRE(assignment.size() == 1);
    REQUIRE(solution.num_assignments == 1);

    auto dancer = assignment[1][1];
    REQUIRE_THAT(dancer_ids, Catch::Matchers::VectorContains(dancer));

    auto dances_performed = solution.dance_performed;
    REQUIRE(dances_performed.size() == 1);
    REQUIRE(dances_performed[1]);

    free_test_logger(logger);
}

TEST_CASE("Two dances, one dancer", "[dance_solver]")
{
    std::vector<Dancer> dancers = {{1, true}};
    std::vector<Dance> dances = {
        {1, {{1}}},
        {2, {{1}}}};
    std::vector<DancerPosition> dancer_positions = {
        {1, 1, 1, PreferenceYes},
        {1, 1, 2, PreferenceYes}};

    auto logger = new_test_logger();
    DanceSolver solver(logger, dancers, dances, dancer_positions);

    auto solution = solver.GetPossibleDances();
    REQUIRE(solution.status == SolverStatus::SolverStatusOptimal);

    auto assignment = solution.assignment;
    REQUIRE(assignment.size() == 2);
    REQUIRE(solution.num_assignments == 2);

    auto dances_performed = solution.dance_performed;
    REQUIRE(dances_performed.size() == 2);
    REQUIRE(dances_performed[1]);
    REQUIRE(dances_performed[2]);

    free_test_logger(logger);
}

TEST_CASE("One dance, nobody can dance it", "[dance_solver]")
{
    std::vector<Dancer> dancers = {{1, true}};
    std::vector<Dance> dances = {{1, {{1}}}};
    std::vector<DancerPosition> dancer_positions = {};

    auto logger = new_test_logger();
    DanceSolver solver(logger, dancers, dances, dancer_positions);

    auto solution = solver.GetPossibleDances();
    REQUIRE(solution.status == SolverStatus::SolverStatusInfeasible);

    auto assignment = solution.assignment;
    REQUIRE(assignment.size() == 0);
    REQUIRE(solution.num_assignments == 0);

    auto dances_performed = solution.dance_performed;
    REQUIRE(dances_performed.size() == 1);
    REQUIRE(!dances_performed[1]);

    free_test_logger(logger);
}

TEST_CASE("One dance, two positions, two dancers", "[dance_solver]")
{
    std::vector<Dancer> dancers = {{1, true}, {2, true}};
    std::vector<Dance> dances = {{1, {{1}, {2}}}};
    std::vector<DancerPosition> dancer_positions = {
        {1, 1, 1, PreferenceYes},
        {2, 2, 1, PreferenceYes}};

    auto logger = new_test_logger();
    DanceSolver solver(logger, dancers, dances, dancer_positions);

    auto solution = solver.GetPossibleDances();
    REQUIRE(solution.status == SolverStatus::SolverStatusOptimal);
    REQUIRE(solution.num_assignments == 2);

    auto assignment = solution.assignment;
    REQUIRE(assignment.size() == 1);
    REQUIRE(assignment[1][1] == 1);
    REQUIRE(assignment[1][2] == 2);

    auto dances_performed = solution.dance_performed;
    REQUIRE(dances_performed.size() == 1);
    REQUIRE(dances_performed[1]);

    free_test_logger(logger);
}

TEST_CASE("One dance, two positions, two dancers, but they can't dance both positions", "[dance_solver]")
{
    std::vector<Dancer> dancers = {{1, true}, {2, true}};
    std::vector<Dance> dances = {{1, {{1}, {2}}}};
    std::vector<DancerPosition> dancer_positions = {
        {1, 1, 1, PreferenceYes},
        {2, 1, 1, PreferenceYes}};

    auto logger = new_test_logger();
    DanceSolver solver(logger, dancers, dances, dancer_positions);

    auto solution = solver.GetPossibleDances();
    REQUIRE(solution.status == SolverStatus::SolverStatusInfeasible);

    auto assignment = solution.assignment;
    REQUIRE(assignment.size() == 0);
    REQUIRE(solution.num_assignments == 0);

    auto dances_performed = solution.dance_performed;
    REQUIRE(dances_performed.size() == 1);
    REQUIRE(!dances_performed[1]);

    free_test_logger(logger);
}

TEST_CASE("One dance, one position, one dancer, but they can't dance the position", "[dance_solver]")
{
    std::vector<Dancer> dancers = {{1, true}};
    std::vector<Dance> dances = {{1, {{1}}}};
    std::vector<DancerPosition> dancer_positions = {
        {1, 1, 1, PreferenceNo}};

    auto logger = new_test_logger();
    DanceSolver solver(logger, dancers, dances, dancer_positions);

    auto solution = solver.GetPossibleDances();
    REQUIRE(solution.status == SolverStatus::SolverStatusInfeasible);

    auto assignment = solution.assignment;
    REQUIRE(assignment.size() == 0);
    REQUIRE(solution.num_assignments == 0);

    auto dances_performed = solution.dance_performed;
    REQUIRE(dances_performed.size() == 1);
    REQUIRE(!dances_performed[1]);

    free_test_logger(logger);
}

TEST_CASE("Two dances: one can be danced and the other can't", "[dance_solver]")
{
    std::vector<Dancer> dancers = {{1, true}};
    std::vector<Dance> dances = {
        {1, {{1}}},
        // we have two positions in this dance, to check that the "all
        // different" constraint doesn't apply to non-danced dances
        {2, {{1}, {2}}}};
    std::vector<DancerPosition> dancer_positions = {
        {1, 1, 1, PreferenceYes}};

    auto logger = new_test_logger();
    DanceSolver solver(logger, dancers, dances, dancer_positions);

    auto solution = solver.GetPossibleDances();
    REQUIRE(solution.status == SolverStatus::SolverStatusOptimal);
    REQUIRE(solution.num_assignments == 1);

    auto assignment = solution.assignment;
    REQUIRE(assignment.size() == 1);
    REQUIRE(assignment[1][1] == 1);

    auto dances_performed = solution.dance_performed;
    REQUIRE(dances_performed.size() == 2);
    REQUIRE(dances_performed[1]);
    REQUIRE(!dances_performed[2]);

    free_test_logger(logger);
}

TEST_CASE("Two dance, two dancers, both should be given one of them", "[dance_solver]")
{
    std::vector<Dancer> dancers = {{1, true}, {2, true}};
    std::vector<Dance> dances = {
        {1, {{1}}},
        {2, {{1}}}};
    std::vector<DancerPosition> dancer_positions = {
        {1, 1, 1, PreferenceYes},
        {1, 1, 2, PreferenceYes},
        {2, 1, 1, PreferenceYes},
        {2, 1, 2, PreferenceYes}};

    auto logger = new_test_logger();
    DanceSolver solver(logger, dancers, dances, dancer_positions);

    auto solution = solver.GetPossibleDances();
    REQUIRE(solution.status == SolverStatus::SolverStatusOptimal);
    REQUIRE(solution.num_assignments == 2);

    auto assignment = solution.assignment;
    REQUIRE(assignment.size() == 2);
    REQUIRE(((assignment[1][1] == 1 && assignment[2][1] == 2) || (assignment[1][1] == 2 && assignment[2][1] == 1)));

    auto dances_performed = solution.dance_performed;
    REQUIRE(dances_performed.size() == 2);
    REQUIRE(dances_performed[1]);
    REQUIRE(dances_performed[2]);

    free_test_logger(logger);
}

TEST_CASE("Preferences are taken into account, simple case")
{
    std::vector<Dancer> dancers = {{1, true}, {2, true}};
    std::vector<Dance> dances = {{1, {{1}}}};
    std::vector<DancerPosition> dancer_positions = {
        {1, 1, 1, PreferenceYes},
        {2, 1, 1, PreferenceFavourite}};

    auto logger = new_test_logger();
    DanceSolver solver(logger, dancers, dances, dancer_positions);

    auto solution = solver.GetPossibleDances();
    REQUIRE(solution.status == SolverStatus::SolverStatusOptimal);
    REQUIRE(solution.num_assignments == 1);

    auto assignment = solution.assignment;
    REQUIRE(assignment.size() == 1);
    REQUIRE(assignment[1][1] == 2);

    auto dances_performed = solution.dance_performed;
    REQUIRE(dances_performed.size() == 1);
    REQUIRE(dances_performed[1]);

    free_test_logger(logger);
}

// These are four dancers who can dance "Strip The Willow" as of 2023-06-29
TEST_CASE("Real world")
{
    std::vector<Dancer> dancers = {{3, true}, {9, true}, {10, true}, {13, true}};
    std::vector<Dance> dances = {{9, {0, 1, 2, 3}}, {2, {0, 1, 2, 3, 4, 5}}};
    std::vector<DancerPosition> dancer_positions = {
        {10, 0, 9, PreferenceMaybe},
        {10, 1, 9, PreferenceMaybe},
        {10, 2, 9, PreferenceMaybe},
        {10, 3, 9, PreferenceMaybe},
        {3, 0, 9, PreferenceYes},
        {3, 1, 9, PreferenceMaybe},
        {3, 2, 9, PreferenceMaybe},
        {3, 3, 9, PreferenceYes},
        {13, 0, 9, PreferenceNo},
        {13, 1, 9, PreferenceNo},
        {13, 2, 9, PreferenceNo},
        {13, 3, 9, PreferenceMaybe},
        {9, 0, 9, PreferenceNo},
        {9, 1, 9, PreferenceMaybe},
        {9, 2, 9, PreferenceMaybe},
        {9, 3, 9, PreferenceNo},
        {10, 0, 2, PreferenceMaybe},
        {10, 1, 2, PreferenceMaybe},
        {10, 2, 2, PreferenceMaybe},
        {10, 3, 2, PreferenceMaybe},
        {10, 4, 2, PreferenceMaybe},
        {10, 5, 2, PreferenceMaybe},
    };

    auto logger = new_test_logger();
    DanceSolver solver(logger, dancers, dances, dancer_positions);

    auto solution = solver.GetPossibleDances();
    REQUIRE(solution.status == SolverStatus::SolverStatusOptimal);
    REQUIRE(solution.num_assignments == 4);

    auto assignment = solution.assignment;
    REQUIRE(assignment.size() == 1);
    REQUIRE(assignment[9].size() == 4);
    REQUIRE(assignment[9][0] == 3);
    REQUIRE(assignment[9][1] == 9);
    REQUIRE(assignment[9][2] == 10);
    REQUIRE(assignment[9][3] == 13);

    auto dances_performed = solution.dance_performed;
    REQUIRE(dances_performed.size() == 2);
    REQUIRE(dances_performed[9]);
    REQUIRE(!dances_performed[11]);

    free_test_logger(logger);
}
