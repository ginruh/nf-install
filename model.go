package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Font struct {
	Name        string
	DownloadURL string
}

func (f *Font) Download() error {
	downloadURL := f.DownloadURL
	tmpDirPath := os.TempDir()
	programDir := "nfinstall"
	baseDir := fmt.Sprintf("%s%s%s", tmpDirPath, "/", programDir)
	if err := os.Mkdir(baseDir, 0755); err != nil {
		if !errors.Is(err, fs.ErrExist) {
			return err
		}
	}
	res, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	fileBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if err = os.WriteFile(
		fmt.Sprintf("%s%s%s", baseDir, "/", f.Name),
		fileBytes,
		0644,
	); err != nil {
		return err
	}
	return nil
}

type Fonts []Font

func (f Fonts) Get(name string) (Font, error) {
	for _, font := range f {
		if font.Name == name {
			return font, nil
		}
	}
	return Font{}, fmt.Errorf("%s font not found", name)
}

type GetFontResult Font
type GetFontError struct{ err error }

func getFont(name string, fonts Fonts) tea.Cmd {
	return func() tea.Msg {
		font, err := fonts.Get(name)
		if err != nil {
			return GetFontError{err}
		}
		return GetFontResult(font)
	}
}

type FetchFontsResult Fonts
type FetchFontsError struct{ err error }

func fetchFonts() tea.Msg {
	var fonts []Font
	release, err := fetchLatestRelease()
	if err != nil {
		return FetchFontsError{err}
	}
	for _, asset := range release.Assets {
		if strings.HasSuffix(asset.Name, ".zip") {
			fonts = append(
				fonts,
				Font{Name: asset.Name, DownloadURL: asset.DownloadURL},
			)

		}
	}
	return FetchFontsResult(fonts)
}

type State int

const (
	Fetch State = iota
	Choose
	Download
	Install
)

type Model struct {
	State        State
	Fonts        []Font
	SelectedFont Font
	Spinner      spinner.Model
	FontInput    textinput.Model
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		fetchFonts,
		m.Spinner.Tick,
	)
}

func Setup() Model {
	// fonts fetching loader
	s := spinner.New()
	s.Spinner = spinner.Dot
	// input text
	fontInput := textinput.New()
	fontInput.Placeholder = "Enter font name"
	fontInput.Focus()
	return Model{
		Spinner:      s,
		State:        Fetch,
		FontInput:    fontInput,
		Fonts:        []Font{},
		SelectedFont: Font{},
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			switch m.State {
			case Choose:
				return m, getFont(m.FontInput.Value(), m.Fonts)
			default:
				return m, tea.Quit
			}
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		default:
			switch m.State {
			case Choose:
				m.FontInput, cmd = m.FontInput.Update(msg)
				return m, cmd
			}
		}
	case FetchFontsResult:
		m.Fonts = msg
		m.State = Choose
		return m, cmd
	case GetFontResult:
		m.State = Download
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd
	default:
		switch m.State {
		case Fetch, Download:
			m.Spinner, cmd = m.Spinner.Update(msg)
			return m, cmd
		}
		return m, cmd
	}
	return m, cmd
}

func (m Model) View() string {
	switch m.State {
	case Fetch:
		return fmt.Sprintf("%sFetching Nerd fonts", m.Spinner.View())
	case Choose:
		return fmt.Sprintf("Enter font name: %s\n\n%s\n", m.FontInput.View(), "(esc to quit)")
	case Download:
		return fmt.Sprintf("%sDownloading %s", m.Spinner.View(), m.FontInput.Value())
	default:
		return ""
	}
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
