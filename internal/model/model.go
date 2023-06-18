package model

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// 0 = no
// 1 = maybe
// 2 = yes
// 3 = favourite
type DancePreference int

const (
	PreferenceNo        DancePreference = 0
	PreferenceMaybe     DancePreference = 1
	PreferenceYes       DancePreference = 2
	PreferenceFavourite DancePreference = 3
)

func (dp *DancePreference) Scan(value interface{}) error {
	if value == nil {
		*dp = PreferenceNo
		return fmt.Errorf("invalid dance preference value: %v", value)
	}

	switch value.(int64) {
	case 0:
		*dp = PreferenceNo
	case 1:
		*dp = PreferenceMaybe
	case 2:
		*dp = PreferenceYes
	case 3:
		*dp = PreferenceFavourite
	default:
		return fmt.Errorf("invalid dance preference value: %v", value)
	}
	return nil
}

// 1 = Dancer
// 2 = Musician
// 3 = Both
type Role int

const (
	RoleDancer   Role = 1
	RoleMusician Role = 2
	RoleBoth     Role = 3
)

func (r *Role) Scan(value interface{}) error {
	if value == nil {
		*r = RoleDancer
		return fmt.Errorf("invalid role value: %v", value)
	}

	switch value.(int64) {
	case 1:
		*r = RoleDancer
	case 2:
		*r = RoleMusician
	case 3:
		*r = RoleBoth
	default:
		return fmt.Errorf("invalid role value: %v", value)
	}
	return nil
}

func (dp DancePreference) Value() (interface{}, error) {
	return dp, nil
}

func (dp DancePreference) String() string {
	switch dp {
	case PreferenceNo:
		return "no"
	case PreferenceMaybe:
		return "maybe"
	case PreferenceYes:
		return "yes"
	case PreferenceFavourite:
		return "favourite"
	default:
		return "unknown"
	}
}

type Dancer struct {
	ID              int
	Name            string
	DancerPositions []*DancerPosition `gorm:"foreignKey:DancerID"`
	Active          bool
	Type            Role `gorm:"type:integer;default:1"`
}

type Position struct {
	PositionID      int `gorm:"column:position;primaryKey"`
	Name            string
	DanceID         int               `gorm:"column:dance;primaryKey"`
	Dance           *Dance            `gorm:"foreignKey:DanceID"`
	DancerPositions []*DancerPosition `gorm:"foreignKey:DanceID,PositionID;references:DanceID,PositionID"`
}

func (p Position) String() string {
	return p.Name
}

type Dance struct {
	ID        int
	Active    bool
	Name      string
	Note      string
	Positions []*Position `gorm:"foreignKey:DanceID"`
}

type DancerPosition struct {
	DancerID   int `gorm:"column:dancer;primaryKey"`
	PositionID int `gorm:"column:position;primaryKey"`
	DanceID    int `gorm:"column:dance;primaryKey"`

	Dance      *Dance          `gorm:"foreignKey:DanceID"`
	Position   *Position       `gorm:"foreignKey:PositionID,DanceID"`
	Dancer     *Dancer         `gorm:"foreignKey:DancerID"`
	Preference DancePreference `gorm:"type:integer;default:0"`
}

func (DancerPosition) TableName() string {
	return "dancerposition"
}

func (dp DancerPosition) String() string {
	return fmt.Sprintf("%s: %s: %s (%s)", dp.Dance.Name, dp.Dancer.Name, dp.Position.Name, dp.Preference)
}

// AssignmentSet is a map of dance to position to dancer
type Assignments map[*Dance]map[*Position]*Dancer
type DancesDanced map[*Dance]struct{}

type AssignmentSet struct {
	dancesDanced DancesDanced
	assignments  Assignments
}

func NewAssignmentSet(assignments Assignments, dancesDanced DancesDanced) AssignmentSet {
	return AssignmentSet{
		assignments:  assignments,
		dancesDanced: dancesDanced,
	}
}

func (as AssignmentSet) DancerFor(d *Dance, p *Position) *Dancer {
	return as.assignments[d][p]
}

func (as AssignmentSet) NumDancesDanced() int {
	return len(as.dancesDanced)
}

func (d *Dance) IsDanced(as AssignmentSet) bool {
	_, ok := as.dancesDanced[d]

	return ok
}

func (as AssignmentSet) String() string {
	var sb strings.Builder

	for dance, positions := range as.assignments {
		sb.WriteString(fmt.Sprintf("%s:\n", dance.Name))
		for position, dancer := range positions {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", position.Name, dancer.Name))
		}
	}

	return sb.String()
}

type Model struct {
	DB *gorm.DB

	logger *logrus.Entry
}

func NewModel(databaseName string, logger *logrus.Entry) (*Model, error) {
	db, err := gorm.Open(sqlite.Open(databaseName), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if logger != nil {
		db.Logger = NewLogrusLogger(logger)
	}

	/*
		err = db.AutoMigrate(&Dancer{}, &Position{}, &Dance{}, &DancerPosition{})
		if err != nil {
			return nil, err
		}
	*/

	return &Model{DB: db, logger: logger}, nil
}

func (m *Model) FetchDances() ([]*Dance, error) {
	var dances []*Dance
	result := m.DB.Debug().
		Preload("Positions").
		// There are no FK constraints, so we check here that the dance/dancers referred to still exist in the DB.
		Preload(
			"Positions.DancerPositions",
			"dancer IN (SELECT id from dancers) AND dance IN (SELECT id from dances)",
		).
		Preload("Positions.DancerPositions.Dancer").
		Preload("Positions.DancerPositions.Dance").
		Preload("Positions.DancerPositions.Position").
		Order("name").
		Find(&dances)

	return dances, result.Error
}

/*
func (m *Model) FetchDancerPositionsForDancers(dancers []*Dancer) ([]*DancerPosition, error) {
	var dancerids []int
	for _, dancer := range dancers {
		dancerids = append(dancerids, dancer.ID)
	}

	var dancerPositions []*DancerPosition
	result := m.DB.
		Preload("Dance").
		Preload("Dance.Positions").
		Preload("Position").
		Preload("Dancer").
		Where("dancer IN ?", dancerids).
		Find(&dancerPositions)

	return dancerPositions, result.Error
}
*/

func (m *Model) FetchDancerPositionsForDancers(dancers []*Dancer) ([]*Dance, []*DancerPosition, error) {
	var dancerids []int
	for _, dancer := range dancers {
		dancerids = append(dancerids, dancer.ID)
	}

	// Fetch DancerPosition without Position preload
	var dancerPositions []*DancerPosition
	result := m.DB.
		Where("dancer IN ?", dancerids).
		Find(&dancerPositions)

	if result.Error != nil {
		return nil, nil, result.Error
	}

	// Fetch Dance with Positions
	var dances []*Dance
	result = m.DB.
		Preload("Positions").
		Order("name").
		Find(&dances)

	if result.Error != nil {
		return nil, nil, result.Error
	}

	// Create a map of Position pointers for each DanceID
	danceMap := make(map[int]*Dance)
	for _, dance := range dances {
		danceMap[dance.ID] = dance
	}

	dancerMap := make(map[int]*Dancer)
	for _, dancer := range dancers {
		dancerMap[dancer.ID] = dancer
	}

	// Iterate over DancerPosition and assign the correct Position. This ensures
	// that the same Position pointer is used in Dance and DancerPosition.
	for _, dp := range dancerPositions {
		dp.Dance = danceMap[dp.DanceID]
		positions := dp.Dance.Positions
		for _, position := range positions {
			position.Dance = dp.Dance
			if position.PositionID == dp.PositionID {
				dp.Position = position
				break
			}
		}
		dp.Dancer = dancerMap[dp.DancerID]
	}

	return dances, dancerPositions, nil
}

func (m *Model) FetchDancers() ([]*Dancer, error) {
	var dancers []*Dancer
	result := m.DB.
		Order("name").
		Preload("DancerPositions.Dance").
		Preload("DancerPositions.Position").
		Find(&dancers)
	return dancers, result.Error
}

// FetchDancersByName returns a list of dancers with the given names. If a name
// is not found, an error is returned.
func (m *Model) FetchDancersByName(names []string) ([]*Dancer, error) {
	var dancers []*Dancer
	result := m.DB.
		Where("name IN ?", names).
		Order("name").
		Preload("DancerPositions.Dance").
		Preload("DancerPositions.Position").
		Find(&dancers)

	if len(dancers) != len(names) {
		nameSet := make(map[string]struct{})

		for _, name := range names {
			nameSet[name] = struct{}{}
		}

		for _, dancer := range dancers {
			delete(nameSet, dancer.Name)
		}

		missing := maps.Keys(nameSet)

		return nil, fmt.Errorf("missing dancers: %s", strings.Join(missing, ", "))
	}

	return dancers, result.Error

}
