package models

import "github.com/charmbracelet/lipgloss"

// Items represents a menu item.

type Logs struct {
	Logs        string
	LogsPages   []string
	CurrentPage int
}

type Styles struct {
	BorderColor lipgloss.Color
}

type Items struct {
	Id   string
	Name string
  Command string
}

type LogsFetchedMsg struct {
	Logs string
}

type Action struct {
	Finished bool
	Error    string
}
