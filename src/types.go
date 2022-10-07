package main

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
)

type ConfigItem struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type Config struct {
	Todos      []ConfigItem `json:"todos"`
	InProgress []ConfigItem `json:"in_progress"`
	Done       []ConfigItem `json:"done"`
}

type Model struct {
	loaded   bool
	focused  status
	lists    []list.Model
	err      error
	quitting bool
}

type Task struct {
	status      status
	title       string
	description string
}

type Form struct {
	focused     status
	title       textinput.Model
	description textarea.Model
}
