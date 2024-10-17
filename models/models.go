package models

// Items represents a menu item.
type Items struct {
	Id   string
	Name string
}

type LogsFetchedMsg struct {
	Logs string
}

type Action struct {
	Finished bool
	Error    string
}
