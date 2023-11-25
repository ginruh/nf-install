package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Font struct {
	Name        string
	DownloadURL string
}

type State struct {
	FontInput    textinput.Model
	Fonts        []Font
	SelectedFont Font
}

func (s State) Init() tea.Cmd {
	return textinput.Blink
}

func Setup() State {
	fontInput := textinput.New()
	fontInput.Placeholder = "Enter font name"
	fontInput.Focus()
	return State{
		FontInput: fontInput,
		Fonts:     []Font{},
	}
}

func (s State) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter, tea.KeyCtrlC, tea.KeyEsc:
			return s, tea.Quit
		}
	}
	s.FontInput, cmd = s.FontInput.Update(msg)
	return s, cmd
}

func (s State) View() string {
	return fmt.Sprintf("Enter font name: %s\n\n%s\n", s.FontInput.View(), "(esc to quit)")
}
