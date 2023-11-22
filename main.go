package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func main() {
	fmt.Println("Fetching Nerd Fonts latest release...")
	release, err := fetchLatestRelease()
	if err != nil {
		log.Fatalln("unable to fetch latest release")
	}

	fmt.Printf("Found release %s.\n", release.Name)

	// Looping through assets
	fonts := release.Assets
	totalFonts := 0
	for i := 0; i < len(fonts); i++ {
		fontName := fonts[i].Name
		if strings.HasSuffix(fontName, ".zip") {
			totalFonts += 1
			fmt.Println(fontName)
		}
	}

	fmt.Printf("\nFound %v fonts.\n", totalFonts)
}

type NFRelease struct {
	Name    string           `json:"name"`
	TagName string           `json:"tag_name"`
	Assets  []NFReleaseAsset `json:"assets"`
}

type NFReleaseAsset struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
}

func fetchLatestRelease() (NFRelease, error) {
	var release NFRelease
	response, err := http.Get("https://api.github.com/repos/ryanoasis/nerd-fonts/releases/latest")
	if err != nil {
		return release, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return release, fmt.Errorf("status code received: %v", response.StatusCode)
	}

	if err = json.NewDecoder(response.Body).Decode(&release); err != nil {
		return release, fmt.Errorf("unable to parse json")
	}
	return release, nil
}
