package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"./pandemic"
	"github.com/jroimartin/gocui"
)

func (p *PandemicView) runCommand(gameState *pandemic.GameState, consoleView *gocui.View, commandView *gocui.View) error {
	commandBuffer := strings.Trim(commandView.Buffer(), "\n\t\r ")
	if commandBuffer == "" {
		return nil
	}
	defer commandView.SetCursor(commandView.Origin())
	defer commandView.Clear()

	commandArgs := strings.Split(commandBuffer, " ")
	cmd := commandArgs[0]

	curTurn, err := gameState.GameTurns.CurrentTurn()
	if err != nil {
		return err
	}
	curPlayer := curTurn.Player

	switch cmd {
	case "infect", "i":
		if len(commandArgs) < 2 {
			fmt.Fprintln(consoleView, p.colorWarning("You must pass a city to the infect command."))
			break
		}
		cityName, err := pandemic.GetCityByPrefix(commandArgs[1], gameState)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("%v", err))
			break
		}
		msg, err = gameState.Infect(cityName)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("%v", err))
		} else {
			fmt.Fprintln(consoleView, "Infected %v.", cityName)
			fmt.Fprintln(consoleView, p.colorOhFuck("%v", msg))
		}
	case "next-turn", "n":
		turn, err := gameState.NextTurn()
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("Could not move on to next turn: %v", err))
		} else {
			fmt.Fprintln(consoleView, "It is now %v's turn", turn.Player.HumanName)
			//message := []string{turn.Player.HumanName}
			//if turn.Player.Character != nil && turn.Player.Character.TurnMessage != "" {
			//	message = append(message, strings.Split(turn.Player.Character.TurnMessage, " ")...)
			//}
			//err = exec.Command("say", message...).Run()
			//if err != nil {
			//	fmt.Fprintln(consoleView, p.colorOhFuck("Could not say message out loud: %v", strings.Join(message, " ")))
			//}
		}
	case "give-card", "g":
		if len(commandArgs) != 3 {
			fmt.Fprintln(consoleView, p.colorWarning("Usage: give-card <human-prefix> <city-prefix>"))
			break
		}
		from, err := gameState.GameTurns.CurrentTurn()
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("%v", err))
			break
		}
		to, err := pandemic.GetPlayerByPrefix(commandArgs[1], gameState)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("%v", err))
			break
		}
		if to == nil {
			fmt.Fprintln(consoleView, p.colorWarning("Player with prefix '%v' not found", commandArgs[1]))
			break
		}
		cardName, err := pandemic.GetCardByPrefix(commandArgs[2], gameState)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("%v", err))
			break
		}
		err = gameState.ExchangeCard(from.Player, to, cardName)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("%v", err))
			break
		} else {
			fmt.Fprintln(consoleView, "%v gave %v to %v", from.Player.HumanName, cardName, to.HumanName)
		}
	case "epidemic", "e":
		if len(commandArgs) != 2 {
			fmt.Fprintln(consoleView, p.colorWarning("You must pass a city to the epidemic command."))
			break
		}
		city, err := pandemic.GetCityByPrefix(commandArgs[1], gameState)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("%v", err))
			break
		}
		msg, err = gameState.Epidemic(city)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("%v", err))
			break
		} else {
			fmt.Fprintln(consoleView, "Epidemic in %v. Please update the infect rate (infect-rate[r] N)", city)
			fmt.Fprintln(consoleView, p.colorOhFuck("%v", msg))
		}
	case "infect-rate", "r":
		if len(commandArgs) != 2 {
			fmt.Fprintln(consoleView, p.colorWarning("You must pass an integer value to the infect rate"))
			break
		}
		ir, err := strconv.ParseInt(commandArgs[1], 10, 32)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning(fmt.Sprintf("%v is not a valid infection rate", commandArgs[1])))
		} else {
			fmt.Fprintln(consoleView, "infection rate now %v", ir)
			gameState.InfectionRate = int(ir)
		}
	case "city-infect-level", "ci":
		if len(commandArgs) != 3 {
			fmt.Fprintln(consoleView, p.colorWarning("You must pass a city and infection value"))
			break
		}
		il, err := strconv.ParseInt(commandArgs[2], 10, 32)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning(fmt.Sprintf("%v is not a valid infection level", commandArgs[1])))
			break
		}
		cityName, err := pandemic.GetCityByPrefix(commandArgs[1], gameState)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("%v", err))
			break
		}
		city, err := gameState.GetCity(cityName)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning(fmt.Sprintf("Could not get city %v: %v", cityName, err)))
			break
		}
		city.SetInfections(int(il))
		fmt.Fprintln(consoleView, "Set infection level in %v to %v", city.Name, city.NumInfections)
	case "city-draw", "c":
		if len(commandArgs) != 2 {
			fmt.Fprintln(consoleView, p.colorWarning("You must pass a city or funded event name to draw"))
			break
		}
		cardName, err := pandemic.GetCardByPrefix(commandArgs[1], gameState)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("%v", err))
			break
		}
		err = gameState.DrawCard(cardName)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("%v", err))
			break
		}
		fmt.Fprintln(consoleView, "%v drew %v from city deck", curPlayer.HumanName, cardName)
	case "quarantine", "q":
		if len(commandArgs) != 2 {
			fmt.Fprintln(consoleView, p.colorWarning("quarantine must be called with a city name"))
			break
		}
		cityName, err := pandemic.GetCityByPrefix(commandArgs[1], gameState)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("%v", err))
			break
		}
		err = gameState.Quarantine(cityName)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning(fmt.Sprintf("Could not quarantine %v: %v", cityName, err)))
		} else {
			fmt.Fprintln(consoleView, "Quarantined %v", cityName)
		}
	case "discard", "d":
		if len(commandArgs) != 2 {
			fmt.Fprintln(consoleView, p.colorWarning("discard must be called with a city name"))
			break
		}
		cardName, err := pandemic.GetCardByPrefix(commandArgs[1], gameState)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("%v", err))
			break
		}
		err = curPlayer.Discard(cardName)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("%v", err))
			break
		}
		fmt.Fprintln(consoleView, "%v discarded %v", curPlayer.HumanName, cardName)
	case "remove-quarantine", "rq":
		if len(commandArgs) != 2 {
			fmt.Fprintln(consoleView, p.colorWarning("remove-quarantine must be called with a city name"))
		}
		cityName, err := pandemic.GetCityByPrefix(commandArgs[1], gameState)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("%v", err))
			break
		}
		err = gameState.RemoveQuarantine(cityName)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning(fmt.Sprintf("Could not remove quarantine from %v: %v", cityName, err)))
		} else {
			fmt.Fprintln(consoleView, "Removed quarantine from %v", cityName)
		}
	case "save", "s":
		filename := filepath.Join(gameState.GameName, fmt.Sprintf("game_%v_%v.json", time.Now().UnixNano(), cmd))
		err = os.MkdirAll(gameState.GameName, 0755)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorOhFuck(fmt.Sprintf("Could not create a game name folder: %v", err)))
		}
		data, err := json.Marshal(gameState)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorOhFuck(fmt.Sprintf("Could not marshal gamestate as JSON: %v", err)))
			return nil
		}
		err = ioutil.WriteFile(filename, data, 0644)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorOhFuck(fmt.Sprintf("Could not save gamestate: %v", err)))
			return nil
		}
		fmt.Fprintln(consoleView, "Save succefull")
	case "treat-disease", "t":
		if len(commandArgs) < 2 {
			fmt.Fprintln(consoleView, p.colorWarning("treat-disease[t] <cityName> <infetions>"))
			return nil
		}
		cityName, err := pandemic.GetCityByPrefix(commandArgs[1], gameState)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("%v", err))
			break
		}
		city, err := gameState.GetCity(cityName)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning(fmt.Sprintf("Could not get city %v: %v", cityName, err)))
			break
		}
		il := 1
		if len(commandArgs) == 3 {
			ilp, err := strconv.ParseInt(commandArgs[2], 10, 32)
			if err != nil {
				fmt.Fprintln(consoleView, p.colorWarning(fmt.Sprintf("%v is not a valid infection level", commandArgs[1])))
				break
			}
			il = int(ilp)
		}

		city.TreatInfections(il)
		fmt.Fprintln(consoleView, "Treated %v infections on %v", il, city.Name)
	case "player-location", "pl":
		if len(commandArgs) != 3 {
			fmt.Fprintln(consoleView, p.colorWarning("Usage: player-location[pl] <human-prefix> <city-prefix>"))
			return nil
		}
		player, err := pandemic.GetPlayerByPrefix(commandArgs[1], gameState)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("%v", err))
			break
		}
		cityName, err := pandemic.GetCityByPrefix(commandArgs[2], gameState)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("%v", err))
			break
		}
		player.SetLocation(cityName)
		fmt.Fprintln(consoleView, "%v new location %v", player.HumanName, cityName)
	case "move", "m":
		if len(commandArgs) != 2 {
			fmt.Fprintln(consoleView, p.colorWarning("move must be called with a city name"))
			return nil
		}
		cityName, err := pandemic.GetCityByPrefix(commandArgs[1], gameState)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("%v", err))
			break
		}
		curPlayer.SetLocation(cityName)
		fmt.Fprintln(consoleView, "%v moved to %v", curPlayer.HumanName, cityName)
	case "help", "h":
		fmt.Fprintln(consoleView, "Help")
		fmt.Fprintln(consoleView, "infect                  i")
		fmt.Fprintln(consoleView, "next-turn               n")
		fmt.Fprintln(consoleView, "give-card               g")
		fmt.Fprintln(consoleView, "epidemic                e")
		fmt.Fprintln(consoleView, "infect-rate             r")
		fmt.Fprintln(consoleView, "city-infect-level       ci")
		fmt.Fprintln(consoleView, "city-draw               c")
		fmt.Fprintln(consoleView, "quarantine              q")
		fmt.Fprintln(consoleView, "discard                 d")
		fmt.Fprintln(consoleView, "remove-quarantine       rq")
		fmt.Fprintln(consoleView, "save                    s")
		fmt.Fprintln(consoleView, "treat-disease           t")
		fmt.Fprintln(consoleView, "player-location         l")
		fmt.Fprintln(consoleView, "move                    m")
	default:
		fmt.Fprintln(consoleView, p.colorWarning(fmt.Sprintf("Unrecognized command %v", cmd)))
		return nil
	}

	return nil
}
