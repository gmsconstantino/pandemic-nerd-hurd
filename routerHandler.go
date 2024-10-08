package main

import (
	"fmt"
	"net/http"

	"github.com/jroimartin/gocui"

	"github.com/gmsconstantino/pandemic-nerd-hurd/pandemic"
)

// Index Root API handler
func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello Api Pandemic Nerd Hurd")
}

type command struct {
	cmd string
}

// Command receive the request from game and translates it to this simulator
func Command(gameState *pandemic.GameState, view *PandemicView, gui *gocui.Gui) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		/*
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			cmd := string(body)
		*/

		var cmd string

		switch r.Method {
		case "GET":
			http.Error(w, "got GET", http.StatusBadRequest)
		case "POST":
			// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			cmd = r.FormValue("cmd")
		default:
			fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
		}

		consoleView, err := gui.View("Console")
		//commandView, err := gui.View("Commands")
		if err != nil {
			gui.Close()
			view.logger.Fatalln("Console view not found, game view not set up correctly")
			return
		}

		view.runStaticCommand(cmd, gameState, consoleView)

		gui.Update(func(g *gocui.Gui) error {
			return nil
		})
	}
}

// NextTurn handle the change of turn
func NextTurn(gameState *pandemic.GameState, view *PandemicView, gui *gocui.Gui) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		consoleView, err := gui.View("Console")
		//commandView, err := gui.View("Commands")
		if err != nil {
			gui.Close()
			view.logger.Fatalln("Console view not found, game view not set up correctly")
			return
		}
		view.runStaticCommand("n", gameState, consoleView)

		gui.Update(func(g *gocui.Gui) error {
			return nil
		})
	}
}

/*

type person struct {
	Name string
	Age  int
}

func PersonCreate(w http.ResponseWriter, r *http.Request) {
	// Declare a new Person struct.
	var p person

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Do something with the Person struct...
	fmt.Fprintf(w, "Person: %+v", p)
}

*/
