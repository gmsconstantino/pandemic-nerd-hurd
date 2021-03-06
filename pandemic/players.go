package pandemic

import (
	"fmt"
)

type CharacterType string

const (
	Medic                 = "Medic"
	Dispatcher            = "Dispatcher"
	Researcher            = "Researcher"
	Scientist             = "Scientist"
	Civilian              = "Civilian"
	QuarantineSpecialist  = "QuarantineSpecialist"
	Colonel               = "Colonel"
	OperationsExpert      = "OperationsExpert"
	Generalist            = "Generalist"
	Soldier               = "Soldier"
	Virologist            = "Virologist"
	Epidemiologist        = "Epidemiologist"
	GeneSplicer           = "GeneSplicer"
	FirstResponder        = "FirstResponder"
	Pharmacist            = "Pharmacist"
	LocalLiason           = "LocalLiason"
	FieldDirector         = "FieldDirector"
	Pilot                 = "Pilot"
	FieldOperative        = "FieldOperative"
	Troubleshooter        = "Troubleshooter"
	Archivist             = "Archivist"
	ContainmentSpecialist = "ContainmentSpecialist"
	ContingencyPlanner    = "ContingencyPlanner"
)

type Player struct {
	HumanName  string     `json:"human_name"`
	Character  *Character `json:"character"`
	Location   CityName   `json:"location"`
	StartCards []CardName `json:"start_cards"`
	Cards      []*CityCard
}

func (p *Player) Discard(cardName CardName) error {
	filtered := []*CityCard{}
	for _, card := range p.Cards {
		if card.Name() != cardName {
			filtered = append(filtered, card)
		}
	}
	if len(filtered) == len(p.Cards) {
		return fmt.Errorf("%v does not seem to have %v\n", p.HumanName, cardName)
	}
	p.Cards = filtered
	return nil
}

func (p *Player) SetLocation(location CityName) error {
	p.Location = location
	return nil
}

type Character struct {
	Name        string        `json:"name"`
	Type        CharacterType `json:"type"`
	TurnMessage string        `json:"turn_message"`
}
