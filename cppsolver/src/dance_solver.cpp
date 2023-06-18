#include <ranges>

#include "ortools/sat/cp_model.h"
#include "ortools/sat/cp_model.pb.h"
#include "ortools/sat/cp_model_solver.h"
#include "ortools/util/sorted_interval_list.h"

#include "dance_solver.hpp"

using namespace operations_research;
using namespace sat;

typedef std::map<int64_t, DancePreference> DancerPreferenceMap;
typedef std::map<int, DancerPreferenceMap> PositionDancerPreferenceMap;
typedef std::map<int, PositionDancerPreferenceMap> DancePositionDancerPreferenceMap;

class DanceSolver::DanceSolverImpl
{
public:
    DanceSolverImpl(
        logger *logger,
        const std::vector<Dancer> &dancers,
        const std::vector<Dance> &dances,
        const std::vector<DancerPosition> &dancer_positions);
    const DanceSolution GetPossibleDances();

private:
    void CreateVariablesAndConstraints();
    void ProcessDancerPositions(
        const std::vector<DancerPosition> &dancer_positions);
    void ProcessDance(
        const Dance &dance);
    const IntVar ProcessDancePosition(
        const Dance &dance,
        const Position &position,
        const DancerPreferenceMap &position_preference_map);
    void ProcessDancerForDancePosition(
        const Dance &dance,
        const Position &position,
        const Dancer &dancer,
        const DancerPreferenceMap &dancer_preference_map,
        const IntVar &dance_position_var,
        const IntVar &dance_position_var_alt);
    void HandleDancerPositionPreference(
        const Dance &dance,
        const Position &position,
        const Dancer &dancer,
        const DancerPreferenceMap &dancer_preference_map,
        const BoolVar &dance_is_danced,
        const BoolVar &dancer_is_assigned);
    const LinearExpr CreateObjective();
    const DanceSolution GetSolution(const CpSolverResponse &response);

    logger *logger_;

    CpModelBuilder cp_model_;

    std::vector<Dancer> dancers_;
    std::vector<Dance> dances_;

    const std::vector<DancerPosition> dancer_positions_;

    DancePositionDancerPreferenceMap dancer_position_preference_map_;
    std::map<int, std::set<int>> dance_dancers_;

    // dance id -> position id -> variable
    std::map<int, std::map<int, IntVar>> dancer_position_map_;
    std::map<int, std::vector<BoolVar>> dances_by_dancer_;
    std::vector<IntVar> dancer_counts_;
    std::map<int, BoolVar> dance_is_danced_vars_;

    std::vector<BoolVar> maybes_;
    std::vector<BoolVar> yeses_;
    std::vector<BoolVar> favourites_;

    IntVar min_dances_;
    IntVar max_dances_;
    IntVar dance_diff_;
    IntVar favourite_count_;
    IntVar yes_count_;
    IntVar maybe_count_;
};

__attribute__((visibility("default")))
DanceSolver::DanceSolver(
    logger *logger,
    std::vector<Dancer> &dancers,
    std::vector<Dance> &dances,
    std::vector<DancerPosition> &dancer_positions)
    : pimpl_(std::make_unique<DanceSolver::DanceSolverImpl>(
          logger,
          dancers,
          dances,
          dancer_positions))
{
}

DanceSolver::~DanceSolver() = default;

__attribute__((visibility("default")))
const DanceSolver::DanceSolution
DanceSolver::GetPossibleDances()
{
    return pimpl_->GetPossibleDances();
}

DanceSolver::DanceSolverImpl::DanceSolverImpl(
    logger *logger,
    const std::vector<Dancer> &dancers,
    const std::vector<Dance> &dances,
    const std::vector<DancerPosition> &dancer_positions)
    : logger_(logger),
      dancers_(dancers),
      dances_(dances),
      dancer_positions_(dancer_positions)
{
}

void DanceSolver::DanceSolverImpl::ProcessDancerPositions(const std::vector<DancerPosition> &dancer_positions)
{
    DancePositionDancerPreferenceMap dancer_position_preference_map;
    std::map<int, std::set<int>> dance_dancers;

    for (const auto &dancer_position : dancer_positions)
    {
        const auto dance_id = dancer_position.DanceID;
        const auto position_id = dancer_position.PositionID;

        const auto positionvars = &dancer_position_preference_map[dance_id][position_id];

        (*positionvars)[(int64_t)dancer_position.DancerID] = dancer_position.Preference;
        dance_dancers[dance_id].insert(dancer_position.DancerID);
    }

    dancer_position_preference_map_ = dancer_position_preference_map;
    dance_dancers_ = dance_dancers;
}

void DanceSolver::DanceSolverImpl::HandleDancerPositionPreference(
    const Dance &dance,
    const Position &position,
    const Dancer &dancer,
    const DancerPreferenceMap &dancer_preference_map,
    const BoolVar &dance_is_danced,
    const BoolVar &dancer_is_assigned)
{
    std::string name;
    BoolVar pref;

    const auto dance_id = dance.ID;
    const auto position_id = position.PositionID;
    const auto dancer_id = dancer.ID;

    const auto dancer_str = "dancer_" + std::to_string(dancer_id);
    const auto dance_pref_str = "_dance_" + std::to_string(dance_id) + "_position_" + std::to_string(position_id) + "_dancer_" + std::to_string(dancer_id);

    if (!dancer.Active)
    {
        name = dancer_str + "_not_active";
        cp_model_.AddEquality(dancer_is_assigned, false).WithName(name);

        return;
    }

    auto preference = DancePreference::PreferenceNo;

    const auto it = dancer_preference_map.find(dancer_id);
    if (it != dancer_preference_map.end())
    {
        preference = it->second;
    }

    switch (preference)
    {
    case DancePreference::PreferenceNo:
        // this means the dancer can't dance it at all
        name = "preference_no" + dance_pref_str;
        cp_model_.AddEquality(dancer_is_assigned, false).WithName(name);

        return;
    case DancePreference::PreferenceMaybe:
        name = "preference_maybe" + dance_pref_str;
        pref = cp_model_.NewBoolVar().WithName(name);
        cp_model_.AddEquality(pref, dancer_is_assigned)
            .OnlyEnforceIf(dance_is_danced)
            .WithName(name);
        maybes_.push_back(pref);

        break;
    case DancePreference::PreferenceYes:
        name = "preference_yes" + dance_pref_str;
        pref = cp_model_.NewBoolVar().WithName(name);
        cp_model_.AddEquality(pref, dancer_is_assigned)
            .OnlyEnforceIf(dance_is_danced)
            .WithName(name);
        yeses_.push_back(pref);

        break;
    case DancePreference::PreferenceFavourite:
        name = "preference_favourite" + dance_pref_str;
        pref = cp_model_.NewBoolVar().WithName(name);
        cp_model_.AddEquality(pref, dancer_is_assigned)
            .OnlyEnforceIf(dance_is_danced)
            .WithName(name);
        favourites_.push_back(pref);

        break;
    }

    cp_model_.AddEquality(pref, false)
        .OnlyEnforceIf(Not(dance_is_danced))
        .WithName(name + "_if_dance_not_danced");
}

void DanceSolver::DanceSolverImpl::ProcessDancerForDancePosition(
    const Dance &dance,
    const Position &position,
    const Dancer &dancer,
    const DancerPreferenceMap &dancer_preference_map,
    const IntVar &dance_position_var,
    const IntVar &dance_position_var_alt)
{
    const auto dance_id = dance.ID;
    const auto dancer_id = dancer.ID;
    const auto dance_is_danced = dance_is_danced_vars_[dance_id];
    const auto position_id = position.PositionID;

    const auto dance_id_str = "dance_" + std::to_string(dance_id);
    const auto dancer_id_str = "dancer_" + std::to_string(dancer_id);
    const auto position_id_str = std::to_string(position_id);

    const auto dancing_position = dancer_id_str + "_dancing_" + dance_id_str + "_position_" + position_id_str;
    const auto dancer_is_assigned =
        cp_model_.NewBoolVar().WithName(dancing_position);
    dances_by_dancer_[dancer_id].push_back(dancer_is_assigned);

    HandleDancerPositionPreference(dance, position, dancer, dancer_preference_map, dance_is_danced, dancer_is_assigned);

    cp_model_.AddEquality(dance_position_var, dancer_id)
        .OnlyEnforceIf(dance_is_danced)
        .OnlyEnforceIf(dancer_is_assigned)
        .WithName(dancer_id_str + "_is_assigned_to_" + dance_id_str);
    cp_model_.AddNotEqual(dance_position_var, dancer_id)
        .OnlyEnforceIf(Not(dancer_is_assigned))
        .WithName(dancer_id_str + "_is_not_assigned_to_" + dance_id_str);

    cp_model_.AddEquality(dance_position_var_alt, dancer_id)
        .OnlyEnforceIf(dance_is_danced)
        .OnlyEnforceIf(dancer_is_assigned)
        .WithName(dancer_id_str + "_is_assigned_to_" + dance_id_str + "_alt");
}

const IntVar DanceSolver::DanceSolverImpl::ProcessDancePosition(
    const Dance &dance,
    const Position &position,
    const DancerPreferenceMap &dancer_preference_map)
{
    Domain domain;

    const auto dance_id = dance.ID;
    const auto dance_is_danced = dance_is_danced_vars_[dance_id];
    const auto position_id = position.PositionID;

    const auto dancer_id_keys = dancer_preference_map | std::views::keys;
    const auto dancer_ids = std::vector<int64_t>(dancer_id_keys.begin(), dancer_id_keys.end());

    // update 264 below, dancer_ids.size() should be all dancer ids for this dance.
    const auto all_dancer_ids_for_dance = dance_dancers_[dance_id];

    if (dancer_preference_map.empty())
    {
        domain = Domain(0, dance.Positions.size() - 1);
        cp_model_.AddEquality(dance_is_danced, false)
            .WithName("dance_" + std::to_string(dance_id) + "_not_danced_no_dancer_for_position_" + std::to_string(position_id));
    }
    else if (all_dancer_ids_for_dance.size() < dance.Positions.size())
    {
        Debug(logger_) << "dance " << dance_id << " has " << dancer_ids.size() << " dancers but needs " << dance.Positions.size();
        domain = Domain(0, dance.Positions.size() - 1);
        cp_model_.AddEquality(dance_is_danced, false)
            .WithName("dance_" + std::to_string(dance_id) + "_not_danced_not_enough_dancers_for_position_" + std::to_string(position_id));
    }
    else
    {
        domain = Domain::FromValues(dancer_ids);
    }

    // the main variables for who is dancing the position
    const auto variable_name =
        "dance_" + std::to_string(dance_id) + "_position_" + std::to_string(position_id);
    const auto dance_position_var =
        cp_model_.NewIntVar(domain).WithName(variable_name);
    dancer_position_map_[dance_id][position_id] = dance_position_var;

    // This is a slightly less constrained version of the variable above. The
    // `AddAllDifferent` constraint which we use to ensure all spots are danced
    // by different people doesn't support `OnlyEnforceIf`. We only need to
    // enforce that constraint if the dance is being danced. If we tried to
    // enforce it all the time, we would end up with unsolveable problems, e.g.
    // in the case where there aren't enough dancers for the dance and it's not
    // being danced - we wouldn't be able to come up with an all different
    // assignment set. So we use this variable to enforce the constraint only
    // when the dance is being danced, and if the dance isn't then we don't care
    // what values the position variables have.
    const auto dance_position_var_alt =
        cp_model_.NewIntVar(domain)
            .WithName(variable_name + "_alt");

    for (const auto &dancer : dancers_)
    {
        ProcessDancerForDancePosition(
            dance,
            position,
            dancer,
            dancer_preference_map,
            dance_position_var,
            dance_position_var_alt);
    }

    return dance_position_var_alt;
}

void DanceSolver::DanceSolverImpl::ProcessDance(
    const Dance &dance)
{
    auto dance_id = dance.ID;
    std::vector<IntVar> variablesForDance;
    std::vector<IntVar> variablesForDanceAlt;

    variablesForDance.reserve(dance.Positions.size());
    variablesForDanceAlt.reserve(dance.Positions.size());

    const auto dance_is_danced =
        cp_model_.NewBoolVar()
            .WithName("is_dance_" + std::to_string(dance_id) + "_danced");
    dance_is_danced_vars_[dance_id] = dance_is_danced;

    const auto position_dancer_preference_map = dancer_position_preference_map_[dance_id];

    for (const auto &position : dance.Positions)
    {
        const auto position_id = position.PositionID;

        DancerPreferenceMap dancer_preference_map;

        const auto it = position_dancer_preference_map.find(position_id);
        if (it != position_dancer_preference_map.end())
        {
            dancer_preference_map = it->second;
        }

        const auto dance_position_var_alt = ProcessDancePosition(dance, position, dancer_preference_map);
        variablesForDanceAlt.push_back(dance_position_var_alt);
    }

    // all positions must be filled by a different dancer
    cp_model_.AddAllDifferent(variablesForDanceAlt)
        .WithName("all_positions_different_" + std::to_string(dance_id));
}

const LinearExpr DanceSolver::DanceSolverImpl::CreateObjective()
{
    // at least one dance must be performed
    const auto dance_is_danced_vars_values = dance_is_danced_vars_ | std::views::values;
    const auto dance_is_danced_vars = std::vector<BoolVar>(dance_is_danced_vars_values.begin(), dance_is_danced_vars_values.end());
    cp_model_.AddBoolOr(dance_is_danced_vars).WithName("at_least_one_dance_danced");

    // the number of dances that could possibly be assigned to a dancer
    const Domain dance_domain = {0, (int64_t)dances_.size()};

    // to maximise the fairness among dancers, we need to minimise the
    // difference between max and min dance count. This ensures that dancers get
    // an even number of dances.

    // first we find out how many dances each dancer is doing
    for (const auto &dancer_dances : dances_by_dancer_)
    {
        const auto dancer_id = dancer_dances.first;
        const auto dance_vars_for_dancer = dancer_dances.second;

        const auto dance_count_for_dancer =
            cp_model_.NewIntVar(dance_domain)
                .WithName("dance_count_" + std::to_string(dancer_id));

        cp_model_.AddEquality(dance_count_for_dancer, LinearExpr::Sum(dance_vars_for_dancer))
            .WithName("dance_count_" + std::to_string(dancer_id));

        dancer_counts_.push_back(dance_count_for_dancer);
    }

    // then we get the minimum and maximum of those counts
    min_dances_ = cp_model_.NewIntVar(dance_domain).WithName("min_dances");
    max_dances_ = cp_model_.NewIntVar(dance_domain).WithName("max_dances");
    cp_model_.AddMinEquality(min_dances_, dancer_counts_).WithName("min_dances");
    cp_model_.AddMaxEquality(max_dances_, dancer_counts_).WithName("max_dances");

    // the difference between the max and min dance count is what we are trying
    // to minimize however, the objective function is to maximise, so we negate
    // the difference (* -1). maximising this number is the same as reducing the
    // difference, i.e. equalising the spread as far as possible.
    dance_diff_ = cp_model_.NewIntVar({(int64_t)(-1 * dances_.size() * (dancers_.size() - 1)), 0})
                      .WithName("dance_diff");
    cp_model_.AddEquality(dance_diff_, (max_dances_ - min_dances_) * -1).WithName("dance_diff");

    // count the number of preferences that are satisfied, so we can try to
    // assign people to their favourite positions
    favourite_count_ =
        cp_model_.NewIntVar({0, (int64_t)favourites_.size()})
            .WithName("favourite_count");
    cp_model_.AddEquality(favourite_count_, LinearExpr::Sum(favourites_))
        .WithName("favourite_count");

    yes_count_ =
        cp_model_.NewIntVar({0, (int64_t)yeses_.size()})
            .WithName("yes_count");
    cp_model_.AddEquality(yes_count_, LinearExpr::Sum(yeses_))
        .WithName("yes_count");

    maybe_count_ = cp_model_.NewIntVar({0, (int64_t)maybes_.size()})
                       .WithName("maybe_count");
    cp_model_.AddEquality(maybe_count_, LinearExpr::Sum(maybes_))
        .WithName("maybe_count");

    // count the number of dances that are performed. we will try to maximise this too.
    const auto number_of_dances_performed =
        cp_model_.NewIntVar(dance_domain)
            .WithName("number_of_dances_performed");
    cp_model_.AddEquality(number_of_dances_performed, LinearExpr::Sum(dance_is_danced_vars))
        .WithName("number_of_dances_performed");

    // the objective function is a weighted sum of the above variables
    const auto objective = LinearExpr::WeightedSum(
        {dance_diff_, number_of_dances_performed, favourite_count_, yes_count_, maybe_count_},
        {FAIRNESS_WEIGHT, NUM_DANCES_PERFORMED_WEIGHT, PREFERENCE_FAVOURITE_WEIGHT, PREFERENCE_YES_WEIGHT, PREFERENCE_MAYBE_WEIGHT});

    return objective;
}

void DanceSolver::DanceSolverImpl::CreateVariablesAndConstraints()
{
    ProcessDancerPositions(dancer_positions_);

    for (const auto &dance : dances_)
    {
        ProcessDance(dance);
    }

    cp_model_.Maximize(CreateObjective());
}

const DanceSolver::DanceSolution DanceSolver::DanceSolverImpl::GetSolution(const CpSolverResponse &response)
{
    const auto status = static_cast<SolverStatus>(response.status());

    DancesPerformed dances_performed;
    if (status == SolverStatus::SolverStatusInfeasible)
    {
        for (const auto &dance : dances_)
        {
            dances_performed[dance.ID] = false;
        }
        return {status, 0, dances_performed, {}};
    }

    std::map<int64_t, Dancer> dancer_map;
    for (const auto &dancer : dancers_)
    {
        dancer_map[dancer.ID] = dancer;
    }

    SolutionAssignment positions;
    int num_assignments = 0;

    for (const auto &dance : dances_)
    {
        auto dance_id = dance.ID;

        auto it = dancer_position_map_.find(dance_id);

        if (it == dancer_position_map_.end())
        {
            Debug(logger_) << "Dance: " << dance_id << " not danced, not enough dancers";
            dances_performed[dance_id] = false;
            continue;
        }

        auto var = dance_is_danced_vars_[dance_id];
        const auto is_danced = SolutionBooleanValue(response, var);

        dances_performed[dance_id] = is_danced;

        if (!is_danced)
        {
            Debug(logger_) << "Dance: " << dance_id << " not danced";
            continue;
        }

        auto position_map = it->second;
        for (const auto &position : dance.Positions)
        {
            auto position_id = position.PositionID;
            auto variable = position_map[position_id];
            auto value = SolutionIntegerValue(response, variable);

            Debug(logger_) << "Dance: " << dance_id << " Position: " << position_id << " Dancer: " << value;

            positions[dance_id][position_id] = dancer_map[value].ID;
            num_assignments++;
        }
    }

    // print min, max, diff
    auto min_dances = SolutionIntegerValue(response, min_dances_);
    auto max_dances = SolutionIntegerValue(response, max_dances_);
    auto dance_diff = SolutionIntegerValue(response, dance_diff_);

    Debug(logger_) << "min dances: " << min_dances;
    Debug(logger_) << "max dances: " << max_dances;
    Debug(logger_) << "dance diff: " << dance_diff;

    // print the preferences satisfied
    auto favourite_count = SolutionIntegerValue(response, favourite_count_);
    auto yes_count = SolutionIntegerValue(response, yes_count_);
    auto maybe_count = SolutionIntegerValue(response, maybe_count_);

    Debug(logger_) << "favourite count: " << favourite_count;
    Debug(logger_) << "yes count: " << yes_count;
    Debug(logger_) << "maybe count: " << maybe_count;

    return {status, num_assignments, dances_performed, positions};
}

const DanceSolver::DanceSolution DanceSolver::DanceSolverImpl::GetPossibleDances()
{
    CreateVariablesAndConstraints();

    // Build a solver and configure it
    Model model;

    SatParameters parameters;
    parameters.fill_additional_solutions_in_response();
    parameters.set_instantiate_all_variables(true);
    // parameters.set_enumerate_all_solutions(true);
    // parameters.set_log_search_progress(true);

    model.Add(NewSatParameters(parameters));

    auto b = cp_model_.Build();

    // This is useful when something is broken, you often get errors referring
    // to variables/constraints by index
    for (int i = 0; i < b.variables_size(); ++i)
    {
        auto var = b.variables(i);
        Trace(logger_) << "variable " << std::to_string(i) << " : " << var.name();
    }

    for (int i = 0; i < b.constraints_size(); ++i)
    {
        auto constraint = b.constraints(i);
        Trace(logger_) << "constraint " << std::to_string(i) << " : " << constraint.name();
    }

    // Solve the model.
    const CpSolverResponse response = SolveCpModel(b, &model);
    Debug(logger_) << "Finished: " << CpSolverResponseStats(response);

    return GetSolution(response);
}

extern "C"
{
    struct dance_solver_c_api::Solver
    {
        std::unique_ptr<DanceSolver> impl;
    };

    // C wrapper for the public API so we can call it from Go
    __attribute__((visibility("default")))
    dance_solver_c_api::Solver *
    dance_solver_new_with_logger(
        logger *l,
        dance_solver_c_api::Dancer *dancers, int num_dancers,
        dance_solver_c_api::Dance *dances, int num_dances,
        dance_solver_c_api::DancerPosition *dancer_positions, int num_dancer_positions)
    {
        std::vector<Dancer> cpp_dancers(dancers, dancers + num_dancers);
        std::vector<Dance> cpp_dances(dances, dances + num_dances);
        std::vector<DancerPosition> cpp_dancer_positions(dancer_positions, dancer_positions + num_dancer_positions);

        // log the inputs
        for (const auto &dancer : cpp_dancers)
        {
            Debug(l) << "Dancer: " << dancer.ID;
        }

        for (const auto &dance : cpp_dances)
        {
            Debug(l) << "Dance: " << dance.ID;
            for (const auto &position : dance.Positions)
            {
                Debug(l) << "Dance: " << dance.ID << " Position: " << position.PositionID;
            }
        }
        for (const auto &dancer_position : cpp_dancer_positions)
        {
            Debug(l) << "Dance: " << dancer_position.DanceID << " Position: " << dancer_position.PositionID << " Dancer: " << dancer_position.DancerID << " Preference: " << std::to_string(dancer_position.Preference);
        }

        return new dance_solver_c_api::Solver{
            std::make_unique<DanceSolver>(l, cpp_dancers, cpp_dances, cpp_dancer_positions)};
    }

    struct dance_solver_c_api::DanceSolutionPriv
    {
        DanceSolver::DancesPerformed dances_performed;
        DanceSolver::SolutionAssignment assignments;
    };

    dance_solver_c_api::DanceSolution *dance_solution_new(DanceSolver::DanceSolution solution)
    {
        auto sol = new dance_solver_c_api::DanceSolution();
        sol->priv = new dance_solver_c_api::DanceSolutionPriv();
        sol->status = static_cast<dance_solver_c_api::SolverStatus>(solution.status);
        sol->num_assignments = solution.num_assignments;
        sol->num_dances = solution.dance_performed.size();
        sol->priv->dances_performed = solution.dance_performed;
        sol->priv->assignments = solution.assignment;

        return sol;
    }

    __attribute__((visibility("default"))) void free_dance_solver(dance_solver_c_api::Solver *solver)
    {
        delete solver;
    }

    // C wrapper for the C++ public API, mainly so we can call it from Go
    // Invokes the solver and then flattens the solution into a C struct.  On
    // the Go side we will be copying back into managed memory (Go structs) so
    // we don't bother about making a nice nested structure. It is important to
    // free the memory allocated here after you're done by calling
    // `delete_dance_solution`.
    __attribute__((visibility("default")))
    dance_solver_c_api::DanceSolution *
    get_possible_dances(dance_solver_c_api::Solver *solver_ptr)
    {
        auto solver = solver_ptr->impl.get();
        auto cpp_solution = solver->GetPossibleDances();

        return dance_solution_new(cpp_solution);
    }

    __attribute__((visibility("default"))) int dance_solver_c_api::get_dancer_dance_position(
        dance_solver_c_api::DanceSolution *assignments, int dance_id, int position_id)
    {
        auto assignment = assignments->priv->assignments;
        // the assignment is a map dance id -> position id -> dancer id. we need
        // to look up in this nested map. return -1 if either key is not found.
        auto dance_it = assignment.find(dance_id);
        if (dance_it == assignment.end())
        {
            return -1;
        }

        auto position_map = dance_it->second;
        auto position_it = position_map.find(position_id);
        if (position_it == position_map.end())
        {
            return -1;
        }

        auto dancer_id = position_it->second;

        return dancer_id;
    }

    __attribute__((visibility("default"))) int dance_solver_c_api::is_dance_performed(
        dance_solver_c_api::DanceSolution *solution, int dance_id)
    {
        const auto dances_performed = solution->priv->dances_performed;

        const auto it = dances_performed.find(dance_id);
        if (it == dances_performed.end())
        {
            return 0;
        }

        return it->second ? 1 : 0;
    }

    __attribute__((visibility("default"))) void free_dance_solution(dance_solver_c_api::DanceSolution *solution)
    {
        delete solution->priv;
        solution->priv = nullptr;
        delete solution;
    }
}
