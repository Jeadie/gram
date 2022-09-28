package main

import (
	"encoding/json"
	"fmt"
	"os"
)

const (
	Reset       Colour = "\033[0m"
	Default            = "\033[39m"
	Black              = "\033[30m"
	Red                = "\033[91m"
	DarkRed            = "\033[31m"
	Green              = "\033[92m"
	DarkGreen          = "\033[32m"
	DarkYellow         = "\033[33m"
	Orange             = "\033[38:5:202m" // "\033[93m"
	Blue               = "\033[94m"
	DarkBlue           = "\033[34m"
	Cyan               = "\033[96m"
	Magenta            = "\033[95m"
	DarkMagenta        = "\033[35m"
	DarkCyan           = "\033[36m"
	LightGray          = "\033[37m"
	DarkGray           = "\033[90m"
	White              = "\033[97m"
)

type ColourScheme struct {
	Keyword  Colour `json:"Keyword"`
	Strings  Colour `json:"Strings"`
	Comments Colour `json:"Comments"`
	Numbers  Colour `json:"Numbers"`
	Todos    Colour `json:"Todos"`
	Name     string `json:"Name"`
}

var defaultColourScheme = ColourScheme{
	Keyword: Orange, Strings: Green, Comments: DarkGray, Numbers: Blue, Todos: DarkYellow, Name: "Default",
}

func GetColourScheme() ColourScheme {
	v, exists := os.LookupEnv("GRAM_COLOUR_SCHEME")
	if !exists {
		return defaultColourScheme
	}
	colours := LoadColourSchemesFromFile("colours.json")
	for _, c := range colours {
		if c.Name == v {
			// TODO: Refactor hor colours are stored to allow for better colour scheme formats
			// TODO: Also, can't natively store \033 in file. More reason to migrate colour format
			return defaultColourScheme
		}
	}
	return defaultColourScheme
}

func LoadColourSchemesFromFile(file string) []ColourScheme {
	bytes, err := os.ReadFile(file)
	if err != nil {
		return []ColourScheme{}
	}

	var colours []ColourScheme
	err = json.Unmarshal(bytes, &colours)
	if err != nil {
		panic(fmt.Errorf("%w", err))
	}
	return colours
}
