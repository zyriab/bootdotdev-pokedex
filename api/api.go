package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"

	"github.com/zyriab/pokedex/pokecache"
)

const (
	LocationsURL string = "https://pokeapi.co/api/v2/location-area/"
	PokemonURL   string = "https://pokeapi.co/api/v2/pokemon/"
)

type Env struct {
	Pokedex  map[string]Pokemon
	Next     *string
	Previous *string
	Args     []string
	Cache    *pokecache.Cache
}

type LocationAreaResult struct {
	Next     *string        `json:"next"`
	Previous *string        `json:"previous"`
	Results  []LocationArea `json:"results"`
}

type LocationArea struct {
	Name string `json:"name"`
}

type PokemonsResults struct {
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

type Pokemon struct {
	Name      string `json:"name"`
	Abilities []struct {
		Ability struct {
			Name string `json:"name"`
		} `json:"ability"`
	} `json:"abilities"`
	BaseExperience int `json:"base_experience"`
	Height         int `json:"height"`
	Weight         int `json:"weight"`

	Stats []struct {
		BaseStat int `json:"base_stat"`
		Stat     struct {
			Name string `json:"name"`
		} `json:"stat"`
	} `json:"stats"`

	Types []struct {
		Type struct {
			Name string `json:"name"`
		} `json:"type"`
	} `json:"types"`
}

func calculateCaptureChance(pokemonLevel int) bool {
	baseProbability := math.Max(0.1, 1.0/math.Pow(1.3, float64(pokemonLevel)))

	return rand.Float64() < baseProbability
}

func CatchPokemon(env *Env) (bool, error) {
	if len(env.Args) == 0 {
		return false, errors.New("[CatchPokemon] no pokemon specified")
	}

	name := env.Args[0]

	res, err := http.Get(PokemonURL + name)
	if err != nil {
		return false, fmt.Errorf("[CatchPokemon] %w", err)
	}

	body, err := io.ReadAll(res.Body)
	res.Body.Close()

	if res.StatusCode == 404 {
		return false, fmt.Errorf(
			"[CatchPokemon] %s does not seem to exist",
			name,
		)
	}

	if res.StatusCode > 299 {
		return false, fmt.Errorf(
			"[CatchPokemon] response failed with status code: %d and\nbody: %s",
			res.StatusCode,
			body,
		)
	}

	if err != nil {
		return false, fmt.Errorf("[CatchPokemon] %w", err)
	}

	//nolint:exhaustruct
	pokemon := Pokemon{}

	err = json.Unmarshal(body, &pokemon)
	if err != nil {
		return false, fmt.Errorf("[CatchPokemon] %w", err)
	}

	if ok := calculateCaptureChance(pokemon.BaseExperience); ok {
		env.Pokedex[pokemon.Name] = pokemon

		return true, nil
	}

	return false, nil
}

func GetPokemons(env *Env) ([]string, error) {
	if len(env.Args) == 0 {
		return nil, errors.New("[GetPokemons] no location area specified")
	}

	name := env.Args[0]
	key := LocationsURL + name

	if data, ok := env.Cache.Get(key); ok {
		//noling:exhaustruct
		res := PokemonsResults{}

		err := json.Unmarshal(data, &res)
		if err != nil {
			return nil, fmt.Errorf("[GetPokemons] %w", err)
		}

		pokemons := []string{}
		for _, pokemon := range res.PokemonEncounters {
			pokemons = append(pokemons, pokemon.Pokemon.Name)
		}

		return pokemons, nil
	}

	res, err := http.Get(key)
	if err != nil {
		return nil, fmt.Errorf("[GetPokemons] %w", err)
	}

	body, err := io.ReadAll(res.Body)
	res.Body.Close()

	if res.StatusCode > 299 {
		return nil, fmt.Errorf(
			"[GetPokemons] response failed with status code: %d and\nbody: %s",
			res.StatusCode,
			body,
		)
	}

	if err != nil {
		return nil, fmt.Errorf("[GetPokemons] %w", err)
	}

	env.Cache.Add(key, body)

	//nolint:exhaustruct
	data := PokemonsResults{}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("[GetPokemons] %w", err)
	}

	pokemons := []string{}
	for _, pokemon := range data.PokemonEncounters {
		pokemons = append(pokemons, pokemon.Pokemon.Name)
	}

	return pokemons, nil
}

func GetNextLocationAreas(
	env *Env,
) (LocationAreaResult, error) {
	if env.Next == nil {
		return LocationAreaResult{}, errors.New(
			"next page cannot be nil",
		)
	}

	res, err := getLocationAreas(*env.Next, env.Cache)
	if err != nil {
		return LocationAreaResult{}, err
	}

	return res, nil
}

func GetPreviousLocationAreas(
	env *Env,
) (LocationAreaResult, error) {
	if env.Previous == nil {
		return LocationAreaResult{}, errors.New(
			"next page cannot be nil",
		)
	}

	res, err := getLocationAreas(*env.Previous, env.Cache)
	if err != nil {
		return LocationAreaResult{}, err
	}

	return res, nil
}

func getLocationAreas(
	url string,
	cache *pokecache.Cache,
) (LocationAreaResult, error) {
	if data, ok := cache.Get(url); ok {
		res := LocationAreaResult{}

		err := json.Unmarshal(data, &res)
		if err != nil {
			return LocationAreaResult{}, fmt.Errorf(
				"[getLocationAreas] %w",
				err,
			)
		}

		return res, nil
	}

	res, err := http.Get(url)
	if err != nil {
		return LocationAreaResult{}, fmt.Errorf(
			"[GetLocationAreas] %w",
			err,
		)
	}

	body, err := io.ReadAll(res.Body)
	res.Body.Close()

	if res.StatusCode > 299 {
		return LocationAreaResult{}, fmt.Errorf(
			"[GetLocationAreas] response failed with status code: %d and\nbody: %s",
			res.StatusCode,
			body,
		)
	}

	if err != nil {
		return LocationAreaResult{}, fmt.Errorf(
			"[GetLocationAreas] %w",
			err,
		)
	}

	cache.Add(url, body)

	var loc LocationAreaResult

	err = json.Unmarshal(body, &loc)
	if err != nil {
		return LocationAreaResult{}, fmt.Errorf(
			"[GetLocationAreas] %w",
			err,
		)
	}

	return loc, nil
}
