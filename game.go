package main

import (
	"os"
	"path/filepath"

	"./pandemic"
	"github.com/Sirupsen/logrus"
	"github.com/jroimartin/gocui"

	"gopkg.in/alecthomas/kingpin.v2"

	"net/http"

	"github.com/gorilla/mux"
)

var (
	app              = kingpin.New("pandemic–nerd-hurd", "Start a nerd herd game")
	startCmd         = app.Command("start", "Start a new game")
	startNewGameFile = startCmd.Flag("new-game-file", "The file containing initial data about Cities, Players and Funded Events.").Default("data/new_game.json").ExistingFile()
	startMonth       = startCmd.Flag("month", "The name of the month in the game we are playing. If playing the second time in a month, add '2' after the name").Required().Enum(
		"jan",
		"feb",
		"mar",
		"apr",
		"may",
		"jun",
		"jul",
		"aug",
		"sep",
		"oct",
		"nov",
		"dec",
		"jan2",
		"feb2",
		"mar2",
		"apr2",
		"may2",
		"jun2",
		"jul2",
		"aug2",
		"sep2",
		"oct2",
		"nov2",
		"dec2",
	)
	loadCmd  = app.Command("load", "Load a game from an existing saved game")
	loadFile = loadCmd.Flag("file", "The JSON file containing the game state").Required().ExistingFile()
)

func main() {
	cmd := kingpin.MustParse(app.Parse(os.Args[1:]))

	logger := logrus.New()
	fd, err := os.OpenFile("log.txt", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	logger.Out = fd
	wd, _ := os.Getwd()

	var gameState *pandemic.GameState

	switch cmd {
	case "start":
		gameState, err = pandemic.NewGame(filepath.Join(wd, *startNewGameFile), *startMonth)
		if err != nil {
			logger.Fatalln(err)
		}
	case "load":
		gameState, err = pandemic.LoadGame(filepath.Join(wd, *loadFile))
		if err != nil {
			logger.Fatalln(err)
		}
	}

	view := NewView(logger)
	gui, err := gocui.NewGui(gocui.OutputNormal)

	if err != nil {
		view.logger.Errorln("Could not init GUI: %v", err)
	}
	defer gui.Close()

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", Index)
	router.HandleFunc("/command", Command(gameState, view, gui))
	router.HandleFunc("/nextturn", NextTurn(gameState, view, gui))

	// Sub rotine to handle the request from the api
	go func() {
		logger.Fatalln(http.ListenAndServe(":8080", router))
	}()

	view.Start(gameState, gui)

}
