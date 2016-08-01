package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/anthonybishopric/pandemic-nerd-hurd/pandemic"
	"github.com/jroimartin/gocui"
)

func (p *PandemicView) runCommand(gameState *pandemic.GameState, consoleView *gocui.View, commandView *gocui.View) error {
	commandBuffer := strings.Trim(commandView.Buffer(), "\n\t\r ")
	if commandBuffer == "" {
		return nil
	}

	commandArgs := strings.Split(commandBuffer, " ")
	cmd := commandArgs[0]

	switch cmd {
	case "infect", "i":
		if len(commandArgs) != 2 {
			fmt.Fprintln(consoleView, p.colorWarning("You must pass a city to the infect command."))
			break
		}
		city := commandArgs[1]
		err := gameState.InfectionDeck.Draw(city)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning(err))
		} else {
			cityData, err := gameState.GetCity(city)
			if err != nil {
				panic(fmt.Sprintf("%v present in infection deck but not game state cities", city))
			}
			if cityData.Infect() {
				fmt.Fprintf(consoleView, p.colorOhFuck(fmt.Sprintf("Infected and outbroke %v\n", city)))
			} else {
				fmt.Fprintf(consoleView, "Infected %v\n", city)
			}
		}
	case "epidemic", "e":
		if len(commandArgs) != 2 {
			fmt.Fprintln(consoleView, p.colorWarning("You must pass a city to the epidemic command.\n"))
			break
		}
		city := commandArgs[1]
		err := gameState.InfectionDeck.PullFromBottom(city)
		if err != nil {
			fmt.Fprintln(consoleView, p.colorWarning(err))
			break
		} else {
			fmt.Fprintf(consoleView, "Epidemic in %v. Please update the infect rate (infect-rate N)\n", city)
			cityData, _ := gameState.GetCity(city)
			cityData.Epidemic()
		}
		gameState.InfectionDeck.ShuffleDrawn()
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
	case "city-infect-level", "l":
		if len(commandArgs) != 3 {
			fmt.Fprintln(consoleView, p.colorWarning("You must pass a city and infection value"))
			break
		}
		il, err := strconv.ParseInt(commandArgs[2], 10, 32)
		if err != nil {
			fmt.Fprintf(consoleView, p.colorWarning(fmt.Sprintf("%v is not a valid infection level\n", commandArgs[1])))
			break
		}
		city, err := gameState.GetCity(commandArgs[1])
		if err != nil {
			fmt.Fprintf(consoleView, p.colorWarning(fmt.Sprintf("Could not get city %v: %v\n", commandArgs[1], err)))
			break
		}
		city.SetInfections(int(il))
		fmt.Fprintf(consoleView, "Set infection level in %v to %v\n", city.Name, city.NumInfections)
	case "city-draw", "c":
	default:
		fmt.Fprintf(consoleView, p.colorWarning(fmt.Sprintf("Unrecognized command %v\n", cmd)))
	}

	commandView.Clear()
	return nil
}