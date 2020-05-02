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

	return p.runStaticCommand(commandBuffer, gameState, consoleView)
}

func (p *PandemicView) runStaticCommand(commandBuffer string, gameState *pandemic.GameState, consoleView *gocui.View) error {

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
		msg, err := gameState.Infect(cityName)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("%v", err))
		} else {
			fmt.Fprintf(consoleView, "Infected %v.\n", cityName)
			fmt.Fprintln(consoleView, p.colorOhFuck("%v", msg))
		}
	case "next-turn", "n":
		turn, err := gameState.NextTurn()
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("Could not move on to next turn: %v", err))
		} else {
			fmt.Fprintf(consoleView, "It is now %v's turn\n", turn.Player.HumanName)
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
			fmt.Fprintf(consoleView, "%v gave %v to %v\n", from.Player.HumanName, cardName, to.HumanName)
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
		msg, err := gameState.Epidemic(city)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("%v", err))
			break
		} else {
			fmt.Fprintf(consoleView, "Epidemic in %v. Please update the infect rate (infect-rate[r] N)\n", city)
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
			fmt.Fprintf(consoleView, "infection rate now %v\n", ir)
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
		fmt.Fprintf(consoleView, "Set infection level in %v to %v\n", city.Name, city.NumInfections)
	case "api-city-infect-level", "aci":
		if len(commandArgs) != 3 {
			break
		}
		il, err := strconv.ParseInt(commandArgs[2], 10, 32)
		if err != nil {
			break
		}
		cityName, err := pandemic.GetCityByPrefix(commandArgs[1], gameState)
		if err != nil {
			break
		}
		city, err := gameState.GetCity(cityName)
		if err != nil {
			break
		}
		city.SetInfections(int(il))
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
		fmt.Fprintf(consoleView, "%v drew %v from city deck\n", curPlayer.HumanName, cardName)
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
			fmt.Fprintf(consoleView, "Quarantined %v\n", cityName)
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
		fmt.Fprintf(consoleView, "%v discarded %v\n", curPlayer.HumanName, cardName)
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
			fmt.Fprintf(consoleView, "Removed quarantine from %v\n", cityName)
		}
	case "save", "s":
		filename := filepath.Join(gameState.GameName, fmt.Sprintf("game_%v_%v.json", time.Now().Format("20060102_030405"), cmd))
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
		fmt.Fprintf(consoleView, "Treated %v infections on %v\n", il, city.Name)
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
		fmt.Fprintf(consoleView, "%v new location %v\n", player.HumanName, cityName)
	case "character-location", "cl":
		if len(commandArgs) != 3 {
			fmt.Fprintln(consoleView, p.colorWarning("Usage: character-location[pl] <character-prefix> <city-prefix>"))
			return nil
		}
		player, err := pandemic.GetPlayerByCharacter(commandArgs[1], gameState)
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
		fmt.Fprintf(consoleView, "%v new location %v\n", player.HumanName, cityName)
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
		fmt.Fprintf(consoleView, "%v moved to %v\n", curPlayer.HumanName, cityName)
	case "start-city-draw", "sc":
		if len(commandArgs) != 2 {
			fmt.Fprintln(consoleView, p.colorWarning("You must pass a city or funded event name to draw"))
			break
		}
		cardName, err := pandemic.GetCardByPrefix(commandArgs[1], gameState)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("%v", err))
			break
		}
		err = gameState.StartDrawCard(cardName)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("%v", err))
			break
		}
		fmt.Fprintf(consoleView, "%v drew %v from city deck\n", curPlayer.HumanName, cardName)
	case "undo", "u":
		break
	case "start":
		gameState.StartGame()
		fmt.Fprintf(consoleView, "Game started\n")
	case "help", "h":
		fmt.Fprintln(consoleView, "Help")
		fmt.Fprintln(consoleView, "")
		fmt.Fprintln(consoleView, "start-city-draw		   sc")
		fmt.Fprintln(consoleView, "city-draw               c")
		fmt.Fprintln(consoleView, "")
		fmt.Fprintln(consoleView, "treat-disease           t")
		fmt.Fprintln(consoleView, "infect                  i")
		fmt.Fprintln(consoleView, "city-infect-level       ci")
		fmt.Fprintln(consoleView, "epidemic                e")
		fmt.Fprintln(consoleView, "infect-rate             r")
		fmt.Fprintln(consoleView, "")
		fmt.Fprintln(consoleView, "give-card               g")
		fmt.Fprintln(consoleView, "discard                 d")
		fmt.Fprintln(consoleView, "quarantine              q")
		fmt.Fprintln(consoleView, "remove-quarantine       rq")
		fmt.Fprintln(consoleView, "")
		fmt.Fprintln(consoleView, "move                    m")
		fmt.Fprintln(consoleView, "next-turn               n")
		fmt.Fprintln(consoleView, "")
		fmt.Fprintln(consoleView, "player-location         l")
		fmt.Fprintln(consoleView, "save                    s")

	default:
		fmt.Fprintln(consoleView, p.colorWarning(fmt.Sprintf("Unrecognized command %v", cmd)))
		return nil
	}

	return nil
}
