package main

import (
	"age-downloader/age"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
)

func ValidatePrompt(prompt promptui.Prompt) string {
	result, err := prompt.Run()

	if err != nil {
		panic(fmt.Sprintf("Prompt failed %s\n", err))
	}

	return result
}

func ValidateSelect(prompt promptui.Select) int {
	index, _, err := prompt.Run()

	if err != nil {
		panic(fmt.Sprintf("Prompt failed %s\n", err))
	}

	return index
}

func GenerateAnimeSelectOptions(animesSearchResponse age.AnimeSearchResponse) promptui.Select {
	animes := animesSearchResponse.Results
	return promptui.Select{
		Label: "Select Anime",
		Items: animes,
		Templates: &promptui.SelectTemplates{
			Label:    `{{ . }}?`,
			Active:   `👉 {{ .Name | cyan | bold }}`,
			Inactive: `{{ .Name }}`,
			Selected: `{{ "✔" | green | bold }}  {{ .Name | cyan }}`,
			Details: `
			--------- Anime ----------
			{{ "Name:" | faint }}	{{ .Name }}
			{{ "Intro:" | faint }}	{{ .Intro }}`,
		},
	}
}

func GenerateAnimeEpisodeSelectOptions(animeEpisodes []age.AnimeEpisode) promptui.Select {
	return promptui.Select{
		Label: "Select Episode",
		Items: animeEpisodes,
		Templates: &promptui.SelectTemplates{
			Label:    `{{ . }}?`,
			Active:   `👉 {{ .Name | cyan | bold }}`,
			Inactive: `{{ .Name }}`,
			Selected: `{{ "✔" | green | bold }}  {{ .Name | cyan }}`,
		},
	}
}

func main() {

	validateAnime := func(anime string) error {
		if len(strings.Trim(anime, "")) == 0 {
			fmt.Printf("anima name can't be empty!")
			return errors.New("invalid anime name")
		}
		return nil
	}

	validateDir := func(anime string) error {
		if len(strings.Trim(anime, "")) == 0 {
			fmt.Printf("dir name can't be empty!")
			return errors.New("invalid dir name")
		}
		return nil
	}

	dirPrompt := promptui.Prompt{
		Label:   "Dir",
		Default: ".",
	}

	dir := ValidatePrompt(dirPrompt)

	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Fatalf("create dir err: %v", err)
	}

	domainPrompt := promptui.Prompt{
		Label:    "Domain",
		Default:  "https://www.agedm.org",
		Validate: validateDir,
	}

	domain := ValidatePrompt(domainPrompt)

	client := age.AnimeClient{
		Domain: domain,
	}

	animePrompt := promptui.Prompt{
		Label:    "Anime Name",
		Validate: validateAnime,
	}

	anime := ValidatePrompt(animePrompt)

	page := 1

	animesSearchResponse := client.SearchAnime(anime, page)

	animeSelectOptionPrompt := GenerateAnimeSelectOptions(animesSearchResponse)

	selectedIndex := ValidateSelect(animeSelectOptionPrompt)

	animeEpisodes := client.FetchAnimeEpisodes(animesSearchResponse.Results[selectedIndex])

	animeEpisodesSelectOptionPrompt := GenerateAnimeEpisodeSelectOptions(animeEpisodes)

	selectedIndex = ValidateSelect(animeEpisodesSelectOptionPrompt)

	// TODO download all
	client.DownloadAnime(animeEpisodes[selectedIndex], dir)
}
