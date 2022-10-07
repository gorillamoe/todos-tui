package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type status int

const divisor = 3

const (
	todo status = iota
	inProgress
	done
)

var models []tea.Model

const (
	model status = iota
	form
)

var configFile = "todos.json"
var config Config

func (m *Model) Save() tea.Msg {
	config := Config{}
	for _, list := range m.lists {
		for _, item := range list.Items() {
			task := item.(Task)
			switch task.status {
			case todo:
				config.Todos = append(config.Todos, ConfigItem{Title: task.title, Description: task.description})
			case inProgress:
				config.InProgress = append(config.InProgress, ConfigItem{Title: task.title, Description: task.description})
			case done:
				config.Done = append(config.Done, ConfigItem{Title: task.title, Description: task.description})
			}
		}
	}
	file, _ := json.MarshalIndent(config, "", "  ")
	_ = os.WriteFile(configFile, file, 0644)
	return nil
}

func loadConfig() (Config, error) {
	config := Config{}
	file, err := os.Open(configFile)
	if err != nil {
		return config, error(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return config, error(err)
	}
	return config, nil
}

func NewTask(status status, title, description string) Task {
	return Task{status: status, title: title, description: description}
}

func (m *Model) Delete() tea.Msg {
	selectedItem := m.lists[m.focused].SelectedItem()
	selectedTask := selectedItem.(Task)
	m.lists[selectedTask.status].RemoveItem(m.lists[m.focused].Index())
	m.Save()
	return nil
}

func (t *Task) Prev() {
	if t.status == todo {
		t.status = done
	} else {
		t.status--
	}
}

func (t *Task) Next() {
	if t.status == done {
		t.status = todo
	} else {
		t.status++
	}
}

// implement the list.Item interface
func (t Task) FilterValue() string {
	return t.title
}

func (t Task) Title() string {
	return t.title
}

func (t Task) Description() string {
	return t.description
}

func New() *Model {
	return &Model{}
}

func (m *Model) MoveToPrev() tea.Msg {
	selectedItem := m.lists[m.focused].SelectedItem()
	selectedTask := selectedItem.(Task)
	m.lists[selectedTask.status].RemoveItem(m.lists[m.focused].Index())
	selectedTask.Prev()
	m.lists[selectedTask.status].InsertItem(len(m.lists[selectedTask.status].Items()), list.Item(selectedTask))
	m.Save()
	return nil
}

func (m *Model) MoveToNext() tea.Msg {
	selectedItem := m.lists[m.focused].SelectedItem()
	selectedTask := selectedItem.(Task)
	m.lists[selectedTask.status].RemoveItem(m.lists[m.focused].Index())
	selectedTask.Next()
	m.lists[selectedTask.status].InsertItem(len(m.lists[selectedTask.status].Items()), list.Item(selectedTask))
	m.Save()
	return nil
}

func (m *Model) Next() {
	if m.focused == done {
		m.focused = todo
	} else {
		m.focused++
	}
}

func (m *Model) Prev() {
	if m.focused == todo {
		m.focused = done
	} else {
		m.focused--
	}
}

func (m *Model) initLists() {
	defaultList := list.New([]list.Item{}, list.NewDefaultDelegate(), 5, 5)
	defaultList.SetShowHelp(false)
	m.lists = []list.Model{defaultList, defaultList, defaultList}
	m.lists[todo].Title = "To Do"
	for i := len(config.Todos) - 1; i >= 0; i-- {
		item := config.Todos[i]
		m.lists[todo].InsertItem(0, list.Item(NewTask(todo, item.Title, item.Description)))
	}
	m.lists[inProgress].Title = "In Progress"
	for i := len(config.InProgress) - 1; i >= 0; i-- {
		item := config.InProgress[i]
		m.lists[inProgress].InsertItem(0, list.Item(NewTask(inProgress, item.Title, item.Description)))
	}
	m.lists[done].Title = "Done"
	for i := len(config.Done) - 1; i >= 0; i-- {
		item := config.Done[i]
		m.lists[done].InsertItem(0, list.Item(NewTask(done, item.Title, item.Description)))
	}
}

func (m Model) onWindowSizeMessage(msg tea.WindowSizeMsg) tea.Msg {
	h, v := columnStyle.GetFrameSize()
	globals.vh = msg.Height - v - 1
	globals.vw = msg.Width - h
	columnStyle.Width(globals.vw / divisor)
	focusedStyle.Width(globals.vw / divisor)
	columnStyle.Height(globals.vh)
	focusedStyle.Height(globals.vh)
	m.lists[todo].SetWidth(globals.vw)
	m.lists[todo].SetHeight(globals.vh)
	m.lists[inProgress].SetWidth(globals.vw)
	m.lists[inProgress].SetHeight(globals.vh)
	m.lists[done].SetWidth(globals.vw)
	m.lists[done].SetHeight(globals.vh)
	return nil
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.onWindowSizeMessage(msg)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "D":
			return m, m.Delete
		case "left", "h":
			m.Prev()
		case "right", "l":
			m.Next()
		case "shift+left", "H":
			return m, m.MoveToPrev
		case "shift+right", "L":
			return m, m.MoveToNext
		case "n":
			models[model] = m // save the state of the current model
			models[form] = NewForm(m.focused)
			return models[form].Update(nil)
		}
	case Task:
		task := msg
		m.lists[task.status].InsertItem(len(m.lists[task.status].Items()), task)
		m.Save()
		return m, nil
	}
	var cmd tea.Cmd
	m.lists[m.focused], cmd = m.lists[m.focused].Update(msg)
	return m, cmd
}

func (m Model) View() string {
	todoView := m.lists[todo].View()
	inProgView := m.lists[inProgress].View()
	doneView := m.lists[done].View()
	switch m.focused {
	case inProgress:
		return lipgloss.JoinHorizontal(
			lipgloss.Left,
			columnStyle.Render(todoView),
			focusedStyle.Render(inProgView),
			columnStyle.Render(doneView),
		)
	case done:
		return lipgloss.JoinHorizontal(
			lipgloss.Left,
			columnStyle.Render(todoView),
			columnStyle.Render(inProgView),
			focusedStyle.Render(doneView),
		)
	default:
		return lipgloss.JoinHorizontal(
			lipgloss.Left,
			focusedStyle.Render(todoView),
			columnStyle.Render(inProgView),
			columnStyle.Render(doneView),
		)
	}
}

func main() {
	config, _ = loadConfig()
	models = []tea.Model{New(), NewForm(todo)}
	m := models[model]
	model := m.(*Model)
	model.initLists()
	p := tea.NewProgram(m)
	if err := p.Start(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
