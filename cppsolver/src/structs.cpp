#include "dance_solver.hpp"

/*
 * Conversion functions from the C to the C++ API
 */
Dancer::Dancer(int id, bool active) : ID(id), Active(active) {}

Dancer::Dancer(const dance_solver_c_api::Dancer &dancer)
    : ID(dancer.id), Active(dancer.active != 0) {}

Dancer &Dancer::operator=(const dance_solver_c_api::Dancer &dancer)
{
    ID = dancer.id;
    Active = dancer.active != 0;
    return *this;
}

Position::Position(int position_id) : PositionID(position_id) {}

Position::Position(const dance_solver_c_api::Position &position)
    : PositionID(position.position_id) {}

Position &Position::operator=(const dance_solver_c_api::Position &position)
{
    PositionID = position.position_id;
    return *this;
}

Dance::Dance(int id, std::vector<Position> positions) : ID(id), Positions(positions) {}

Dance::Dance(const dance_solver_c_api::Dance &dance)
    : ID(dance.id), Positions(dance.positions, dance.positions + dance.num_positions) {}

Dance &Dance::operator=(const dance_solver_c_api::Dance &dance)
{
    ID = dance.id;
    Positions = std::vector<Position>(dance.positions, dance.positions + dance.num_positions);
    return *this;
}

DancerPosition::DancerPosition(int dancer_id, int position_id, int dance_id, DancePreference preference)
    : DancerID(dancer_id), PositionID(position_id), DanceID(dance_id), Preference(preference) {}

DancerPosition::DancerPosition(const dance_solver_c_api::DancerPosition &dancer_position)
    : DancerID(dancer_position.dancer_id),
      PositionID(dancer_position.position_id),
      DanceID(dancer_position.dance_id),
      Preference(static_cast<DancePreference>(dancer_position.preference)) {}

DancerPosition &DancerPosition::operator=(const dance_solver_c_api::DancerPosition &dancer_position)
{
    DancerID = dancer_position.dancer_id;
    PositionID = dancer_position.position_id;
    DanceID = dancer_position.dance_id;
    Preference = static_cast<DancePreference>(dancer_position.preference);
    return *this;
}
