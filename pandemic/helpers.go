package pandemic

import (
	"fmt"
	"strings"
)

func GetCardByPrefix(entry string, gs *GameState) (CardName, error) {
	card, err := gs.CityDeck.GetCardByPrefix(entry)
	if err != nil {
		return "", err
	}
	return card.Name(), nil
}

func GetCityByPrefix(entry string, gs *GameState) (CityName, error) {
	card, err := gs.CityDeck.GetCardByPrefix(entry)
	if err != nil {
		return CityName(""), err
	}
	if !card.IsCity() {
		return CityName(""), fmt.Errorf("%v is not a city", card.Name())
	}
	return card.CityName, nil
}

func GetPlayerByPrefix(entry string, gs *GameState) (*Player, error) {
	var ret *Player
	for _, player := range gs.GameTurns.PlayerOrder {
		if strings.HasPrefix(strings.ToLower(player.HumanName), strings.ToLower(entry)) {
			if ret != nil {
				return nil, fmt.Errorf("%v is an ambiguous human name", entry)
			} else {
				ret = player
			}
		}
	}
	return ret, nil
}

func GetPlayerByCharacter(entry string, gs *GameState) (*Player, error) {
	var ret *Player
	for _, player := range gs.GameTurns.PlayerOrder {
		if strings.HasPrefix(strings.ToLower(string(player.Character.Type)), strings.ToLower(entry)) {
			if ret != nil {
				return nil, fmt.Errorf("%v is an ambiguous name", entry)
			} else {
				ret = player
			}
		}
	}
	return ret, nil
}
