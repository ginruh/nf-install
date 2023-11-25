package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(Setup())
	if _, err := p.Run(); err != nil {
		log.Fatalln(err)
		return
	}
}

func oldMain() {
	fmt.Println("Fetching fonts")
	fontAssets, err := fetchFonts()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("Fetched all fonts")
	fmt.Printf("Choose font you want to download: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		log.Fatalln(err)
	}
	selectedFont := scanner.Text()

	var requiredFont *Font
	for _, font := range fontAssets {
		if font.Name == selectedFont {
			requiredFont = &font
			break
		}
	}
	if requiredFont == nil {
		log.Fatalln("user font not found")
	}
	fmt.Println("Beginning font download")
	if err := requiredFont.Download(); err != nil {
		log.Fatalln(err)
	}
}

func (f *Font) Download() error {
	downloadURL := f.DownloadURL
	tmpDirPath := os.TempDir()
	fmt.Println(tmpDirPath)
	// create tmp directory for installer
	if err := os.Mkdir(fmt.Sprintf("%s%s%s", tmpDirPath, "/", "nfinstall"), 0755); err != nil {
		if !errors.Is(err, fs.ErrExist) {
			return err
		}
	}
	// Get request
	res, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	// Read all body
	fileBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	// Save as file to tmp directory
	if err = os.WriteFile(
		fmt.Sprintf("%s%s%s%s%s", tmpDirPath, "/", "nfinstall", "/", f.Name),
		fileBytes,
		0644,
	); err != nil {
		return err
	}
	return nil
}

type Fonts []Font

func (f Fonts) Get(release NFRelease) {
	var fonts []Font
	for _, asset := range release.Assets {
		if strings.HasSuffix(asset.Name, ".zip") {
			fonts = append(
				fonts,
				Font{Name: asset.Name, DownloadURL: asset.DownloadURL},
			)

		}
	}
	f = fonts
}

func fetchFonts() ([]Font, error) {
	var fonts []Font
	release, err := fetchLatestRelease()
	if err != nil {
		return fonts, err
	}
	for _, asset := range release.Assets {
		if strings.HasSuffix(asset.Name, ".zip") {
			fonts = append(
				fonts,
				Font{Name: asset.Name, DownloadURL: asset.DownloadURL},
			)

		}
	}
	return fonts, nil
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
