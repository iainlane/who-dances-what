#include "dance_solver.h"
#include "testlogger.h"

#include <check.h>
#include <stdlib.h>
#include <stdio.h>

static logger *l;

START_TEST(test_one_dance_one_position_one_dancer)
{
    Dancer dancers[] = {{1, 1}};
    Position positions[] = {{1}};
    Dance dances[] = {{1, positions, 1}};
    DancerPosition dancer_positions[] = {{1, 1, 1, PreferenceYes}};

    Solver *solver = dance_solver_new_with_logger(l, dancers, 1, dances, 1, dancer_positions, 1);
    DanceSolution *solution = get_possible_dances(solver);

    ck_assert_int_eq(solution->status, SolverStatusOptimal);
    ck_assert_int_eq(solution->num_assignments, 1);

    ck_assert_int_eq(get_dancer_dance_position(solution, 1, 1), 1);
    ck_assert_int_eq(get_dancer_dance_position(solution, 1, 2), -1);

    ck_assert_int_eq(solution->num_dances, 1);
    ck_assert_int_eq(is_dance_performed(solution, 1), 1);

    free_dance_solution(solution);
    free_dance_solver(solver);
}
END_TEST

START_TEST(test_preference_no)
{
    Dancer dancers[] = {{1, 1}, {2, 1}};
    Position positions[] = {{1}, {2}};
    Dance dances[] = {{1, positions, 2}};
    DancerPosition dancer_positions[] = {
        {1, 1, 1, PreferenceYes},
        {2, 1, 1, PreferenceYes},
        {2, 1, 2, PreferenceNo},
    };

    Solver *solver = dance_solver_new_with_logger(l, dancers, 2, dances, 1, dancer_positions, 3);
    DanceSolution *solution = get_possible_dances(solver);

    ck_assert_int_eq(solution->status, SolverStatusInfeasible);
    ck_assert_int_eq(solution->num_assignments, 0);

    ck_assert_int_eq(solution->num_dances, 1);
    ck_assert_int_eq(is_dance_performed(solution, 1), 0);

    free_dance_solution(solution);
    free_dance_solver(solver);
}

void setup(void)
{
    l = new_test_logger();
}

void teardown(void)
{
    free_test_logger(l);
}

Suite *solver_suite(void)
{
    Suite *s;
    TCase *tc_core;

    s = suite_create("dance_solver");

    /* Core test case */
    tc_core = tcase_create("Core");

    tcase_add_unchecked_fixture(tc_core, setup, teardown);
    tcase_add_test(tc_core, test_one_dance_one_position_one_dancer);
    tcase_add_test(tc_core, test_preference_no);
    suite_add_tcase(s, tc_core);

    return s;
}

int main(void)
{
    int number_failed;
    Suite *s;
    SRunner *sr;

    s = solver_suite();
    sr = srunner_create(s);

    srunner_set_fork_status(sr, CK_NOFORK);
    srunner_run_all(sr, CK_NORMAL);
    number_failed = srunner_ntests_failed(sr);
    srunner_free(sr);

    return (number_failed == 0) ? EXIT_SUCCESS : EXIT_FAILURE;
}
