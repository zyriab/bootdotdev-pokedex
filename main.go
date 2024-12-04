//nolint:forbidigo
package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/zyriab/pokedex/api"
	"github.com/zyriab/pokedex/pokecache"
)

type cliCommand struct {
	callback    func(env *api.Env) error
	name        string
	description string
}

func commands() map[string]cliCommand {
	return map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"exit": {
			name:        "exit",
			description: "Exits the Pokedex",
			callback:    commandExit,
		},
		"map": {
			name:        "map",
			description: "Display the next 20 location areas in the Pokemon world",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Display the previous 20 location areas in the Pokemon world",
			callback:    commandMapb,
		},
		"explore": {
			name:        "explore",
			description: "Explore a given area",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Catch the given pokemon (random)",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "Inspect the given Pokémon in your Pokédex",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "List all your Pokémons",
			callback:    commandPokedex,
		},
	}
}

func commandHelp(_ *api.Env) error {
	fmt.Printf("\n\n")
	fmt.Println("Welcome to the Pokedex!")
	fmt.Printf("usage:\n\n")

	for _, v := range commands() {
		fmt.Printf("%s: %s\n", v.name, v.description)
	}

	fmt.Printf("\n\n")

	return nil
}

func commandExit(_ *api.Env) error {
	os.Exit(0)

	return nil
}

func commandMap(env *api.Env) error {
	res, err := api.GetNextLocationAreas(env)
	if err != nil {
		return fmt.Errorf("[commandMap] %w", err)
	}

	env.Next = res.Next
	env.Previous = res.Previous

	fmt.Printf("\n\n")

	for _, loc := range res.Results {
		fmt.Println(loc.Name)
	}

	fmt.Printf("\n\n")

	return nil
}

func commandMapb(env *api.Env) error {
	if env.Previous == nil {
		//nolint:err113
		return errors.New("[commandMapb] no `previous` page")
	}

	res, err := api.GetPreviousLocationAreas(env)
	if err != nil {
		return fmt.Errorf("[commandMapb] %w", err)
	}

	env.Next = res.Next
	env.Previous = res.Previous

	fmt.Printf("\n\n")

	for _, loc := range res.Results {
		fmt.Println(loc.Name)
	}

	fmt.Printf("\n\n")

	return nil
}

func commandExplore(env *api.Env) error {
	if len(env.Args) == 0 {
		return errors.New("[commandExplore] no location specified")
	}

	pokemons, err := api.GetPokemons(env)
	if err != nil {
		return fmt.Errorf("[commandExplore] %w", err)
	}

	fmt.Printf("\n\n")

	for _, pokemon := range pokemons {
		fmt.Println(pokemon)
	}

	fmt.Printf("\n\n")

	return nil
}

func commandCatch(env *api.Env) error {
	if len(env.Args) == 0 {
		return errors.New("[commandCatch] no location specified")
	}

	ok, err := api.CatchPokemon(env)
	if err != nil {
		return fmt.Errorf("[commandCatch] %w", err)
	}

	name := env.Args[0]

	fmt.Printf("Throwin a Pokéball at %s...\n", name)

	switch ok {
	case true:
		fmt.Printf("%s was caught!\n", name)
	default:
		fmt.Printf("%s escaped!\n", name)
	}

	return nil
}

func commandInspect(env *api.Env) error {
	if len(env.Args) == 0 {
		return errors.New("[commandInspect] no location specified")
	}

	name := env.Args[0]
	pokemon, ok := env.Pokedex[name]
	if !ok {
		fmt.Printf("You haven't catched %s, yet!\n", name)
		return nil
	}

	fmt.Println("Name:", name)
	fmt.Println("Weight:", pokemon.Weight)
	fmt.Println("Height:", pokemon.Height)
	fmt.Println("Stats:")

	for _, stat := range pokemon.Stats {
		fmt.Printf("\t- %s: %d\n", stat.Stat.Name, stat.BaseStat)
	}

	fmt.Println("Types:")

	for _, tpe := range pokemon.Types {
		fmt.Printf("\t- %s\n", tpe.Type.Name)
	}

	return nil
}

func commandPokedex(env *api.Env) error {
	if len(env.Pokedex) == 0 {
		fmt.Println("You haven't catched any Pokémon, yet!")
		return nil
	}

	fmt.Println("Your Pokedex:")

	for k := range env.Pokedex {
		fmt.Printf("\t- %s\n", k)
	}

	return nil
}

func main() {
	fmt.Print("Pokedex > ")

	scanner := bufio.NewScanner(os.Stdin)

	baseURL := "https://pokeapi.co/api/v2/location-area/"

	//nolint:exhaustruct
	env := &api.Env{
		Pokedex: make(map[string]api.Pokemon),
		Next:    &baseURL,
		Args:    []string{},
		Cache:   pokecache.NewCache(5 * time.Second),
	}

	for scanner.Scan() {
		args := strings.Split(scanner.Text(), " ")

		cmd, ok := commands()[args[0]]
		if !ok {
			_ = commandHelp(nil)

			fmt.Print("Pokedex > ")

			continue
		}

		env.Args = args[1:]

		err := cmd.callback(env)
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Print("Pokedex > ")
	}

	if err := scanner.Err(); err != nil {
		log.Fatalln(err)
	}
}
