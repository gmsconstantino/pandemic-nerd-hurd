Many thanks to https://github.com/anthonybishopric/pandemic-nerd-hurd for their amazing work

# Pandemic & Pandemic: Legacy stats collector

Simple code that emits JSON files containing information about playthroughs.
Can do some basic statistical analysis.

## Running

Clone the repo, make sure you have Go 1.6, then:

```
$ go get ./...
$ go build .
$ ./pandemic-nerd-hurd
```

## TODO

_Features_
* Show panic levels in the UI
* Show player turns, which turns caused epidemics
* Track character traits and powerups
* Remind people on their turn what they can do (special abilities)

_Code Fixes_
* Keep pointers to actual epidemic and funded event cards in players / turns
* BUG: current turn on loading a save file is not the correct pointer to a player.

_2020-04-13_
* Added API to receive commands
* Added new Characters from the base game 
* Game setup with hand player empty are now allowed, the card can be drawed in the UI
* Implemented Outbreaks algorithm
* Added Player location to the UI, and _move_ command
* Remove special attributes from blue disease, hardcoded in code
* Added a setup file for the original Pandemic board