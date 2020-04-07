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
			fmt.Fprintln(consoleView, p.colorWarning("You must pass a city to the infect command.\n"))
			break
		}
		cityName, err := pandemic.GetCityByPrefix(commandArgs[1], gameState)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("%v\n", err))
			break
		}
		err = gameState.Infect(cityName)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("%v\n", err))
		} else {
			fmt.Fprintf(consoleView, "Infected %v\n", cityName)
		}
	case "next-turn", "n":
		turn, err := gameState.NextTurn()
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("Could not move on to next turn: %v\n", err))
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
			fmt.Fprintln(consoleView, p.colorWarning("Usage: give-card <human-prefix> <city-prefix>\n"))
			break
		}
		from, err := gameState.GameTurns.CurrentTurn()
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("%v\n", err))
			break
		}
		to, err := pandemic.GetPlayerByPrefix(commandArgs[1], gameState)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("%v\n", err))
			break
		}
		if to == nil {
			fmt.Fprintln(consoleView, p.colorWarning("Player with prefix '%v' not found\n", commandArgs[1]))
			break
		}
		cardName, err := pandemic.GetCardByPrefix(commandArgs[2], gameState)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("%v\n", err))
			break
		}
		err = gameState.ExchangeCard(from.Player, to, cardName)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("%v\n", err))
			break
		} else {
			fmt.Fprintf(consoleView, "%v gave %v to %v\n", from.Player.HumanName, cardName, to.HumanName)
		}
	case "epidemic", "e":
		if len(commandArgs) != 2 {
			fmt.Fprintln(consoleView, p.colorWarning("You must pass a city to the epidemic command.\n"))
			break
		}
		city, err := pandemic.GetCityByPrefix(commandArgs[1], gameState)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("%v", err))
			break
		}
		err = gameState.Epidemic(city)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("%v", err))
			break
		} else {
			fmt.Fprintf(consoleView, "Epidemic in %v. Please update the infect rate (infect-rate[r] N)\n", city)
		}
	case "infect-rate", "r":
		if len(commandArgs) != 2 {
			fmt.Fprintln(consoleView, p.colorWarning("You must pass an integer value to the infect rate\n"))
			break
		}
		ir, err := strconv.ParseInt(commandArgs[1], 10, 32)
		if err != nil {
			fmt.Fprintf(consoleView, p.colorWarning(fmt.Sprintf("%v is not a valid infection rate\n", commandArgs[1])))
		} else {
			fmt.Fprintf(consoleView, "infection rate now %v\n", ir)
			gameState.InfectionRate = int(ir)
		}
	case "city-infect-level", "ci":
		if len(commandArgs) != 3 {
			fmt.Fprintln(consoleView, p.colorWarning("You must pass a city and infection value\n"))
			break
		}
		il, err := strconv.ParseInt(commandArgs[2], 10, 32)
		if err != nil {
			fmt.Fprintf(consoleView, p.colorWarning(fmt.Sprintf("%v is not a valid infection level\n", commandArgs[1])))
			break
		}
		cityName, err := pandemic.GetCityByPrefix(commandArgs[1], gameState)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("%v", err))
			break
		}
		city, err := gameState.GetCity(cityName)
		if err != nil {
			fmt.Fprintf(consoleView, p.colorWarning(fmt.Sprintf("Could not get city %v: %v\n", cityName, err)))
			break
		}
		city.SetInfections(int(il))
		fmt.Fprintf(consoleView, "Set infection level in %v to %v\n", city.Name, city.NumInfections)
	case "city-draw", "c":
		if len(commandArgs) != 2 {
			fmt.Fprintln(consoleView, p.colorWarning("You must pass a city or funded event name to draw\n"))
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
			fmt.Fprintln(consoleView, p.colorWarning("quarantine must be called with a city name\n"))
			break
		}
		cityName, err := pandemic.GetCityByPrefix(commandArgs[1], gameState)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("%v", err))
			break
		}
		err = gameState.Quarantine(cityName)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning(fmt.Sprintf("Could not quarantine %v: %v\n", cityName, err)))
		} else {
			fmt.Fprintf(consoleView, "Quarantined %v\n", cityName)
		}
	case "discard", "d":
		if len(commandArgs) != 2 {
			fmt.Fprintln(consoleView, p.colorWarning("discard must be called with a city name\n"))
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
			fmt.Fprintf(consoleView, p.colorWarning("remove-quarantine must be called with a city name\n"))
		}
		cityName, err := pandemic.GetCityByPrefix(commandArgs[1], gameState)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("%v", err))
			break
		}
		err = gameState.RemoveQuarantine(cityName)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning(fmt.Sprintf("Could not remove quarantine from %v: %v\n", cityName, err)))
		} else {
			fmt.Fprintf(consoleView, "Removed quarantine from %v\n", cityName)
		}
	case "save", "s":
		filename := filepath.Join(gameState.GameName, fmt.Sprintf("game_%v_%v.json", time.Now().UnixNano(), cmd))
		err = os.MkdirAll(gameState.GameName, 0755)
		if err != nil {
			fmt.Fprintf(consoleView, p.colorOhFuck(fmt.Sprintf("Could not create a game name folder: %v\n", err)))
		}
		data, err := json.Marshal(gameState)
		if err != nil {
			fmt.Fprintf(consoleView, p.colorOhFuck(fmt.Sprintf("Could not marshal gamestate as JSON: %v\n", err)))
			return nil
		}
		err = ioutil.WriteFile(filename, data, 0644)
		if err != nil {
			fmt.Fprintf(consoleView, p.colorOhFuck(fmt.Sprintf("Could not save gamestate: %v\n", err)))
			return nil
		}
		fmt.Fprintf(consoleView, "Save succefull\n")
	case "treat-disease", "t":
		if len(commandArgs) < 2 {
			fmt.Fprintf(consoleView, p.colorWarning("treat-disease[t] <cityName> <infetions>\n"))
			return nil
		}
		cityName, err := pandemic.GetCityByPrefix(commandArgs[1], gameState)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("%v", err))
			break
		}
		city, err := gameState.GetCity(cityName)
		if err != nil {
			fmt.Fprintf(consoleView, p.colorWarning(fmt.Sprintf("Could not get city %v: %v\n", cityName, err)))
			break
		}
		il := 1
		if len(commandArgs) == 3 {
			ilp, err := strconv.ParseInt(commandArgs[2], 10, 32)
			if err != nil {
				fmt.Fprintf(consoleView, p.colorWarning(fmt.Sprintf("%v is not a valid infection level\n", commandArgs[1])))
				break
			}
			il = int(ilp)
		}

		city.TreatInfections(il)
		fmt.Fprintf(consoleView, "Treated %v infections on %v\n", il, city.Name)
	case "player-location", "pl":
		if len(commandArgs) != 3 {
			fmt.Fprintf(consoleView, p.colorWarning("Usage: player-location[pl] <human-prefix> <city-prefix>\n"))
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
	case "move", "m":
		if len(commandArgs) != 2 {
			fmt.Fprintf(consoleView, p.colorWarning("move must be called with a city name\n"))
			return nil
		}
		cityName, err := pandemic.GetCityByPrefix(commandArgs[1], gameState)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning("%v", err))
			break
		}
		curPlayer.SetLocation(cityName)
		fmt.Fprintf(consoleView, "%v moved to %v\n", curPlayer.HumanName, cityName)
	case "help", "h":
		fmt.Fprintf(consoleView, "Help\n")
		fmt.Fprintf(consoleView, "infect                  i\n")
		fmt.Fprintf(consoleView, "next-turn               n\n")
		fmt.Fprintf(consoleView, "give-card               g\n")
		fmt.Fprintf(consoleView, "epidemic                e\n")
		fmt.Fprintf(consoleView, "infect-rate             r\n")
		fmt.Fprintf(consoleView, "city-infect-level       ci\n")
		fmt.Fprintf(consoleView, "city-draw               c\n")
		fmt.Fprintf(consoleView, "quarantine              q\n")
		fmt.Fprintf(consoleView, "discard                 d\n")
		fmt.Fprintf(consoleView, "remove-quarantine       rq\n")
		fmt.Fprintf(consoleView, "save                    s\n")
		fmt.Fprintf(consoleView, "treat-disease           t\n")
		fmt.Fprintf(consoleView, "player-location         l\n")
		fmt.Fprintf(consoleView, "move                    m\n")
	default:
		fmt.Fprintf(consoleView, p.colorWarning(fmt.Sprintf("Unrecognized command %v\n", cmd)))
		return nil
	}

	return nil
}
