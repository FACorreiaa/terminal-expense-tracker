package main

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

// Datastructures

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

// initialModel init bubbletea model
func initialModel() model {
	return model{}
}

// entry point
func main() {
	p := tea.NewProgram(initialModel())
	if err, _ := p.Run(); err != nil {
		log.Fatal(err)
	}
}

// Init --- Bubble Tea Init, Update, & View ---
func (m model) Init() tea.Cmd {
	return watchExcelCmd("data.xlsx")
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case excelDataMsg:
		m.expenses = msg.expenses
		m.stonks = msg.stonks
		return m, watchExcelCmd("data.xlsx")

	case errMsg:
		m.err = msg.err
		return m, watchExcelCmd("data.xlsx")

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "e": // Press 'e' to edit stock data
			return m, writeExcelCmd(m.expenses, m.stonks, m.watchList)
		case "1":
			// Pretend we edit expenses
			// ... do something ...
			// Then write back to Excel
			return m, writeExcelCmd(m.expenses, m.stonks, m.watchList)
		}

		return m, nil

	default:
		return m, nil
	}
}

func (m model) View() string {
	s := "\n=== DAILY EXPENSES ===\n"
	for _, e := range m.expenses {
		s += fmt.Sprintf("%-20s %10.2f\n", e.Name, e.Amount)
	}
	s += "\n=== STONKS ===\n"
	for _, st := range m.stonks {
		s += fmt.Sprintf("%-10s %10.2f %-20s %10.2f\n",
			st.Symbol, st.Change, st.Comment, st.Extra)
	}
	s += "\n=== WATCH LIST ===\n"
	for _, w := range m.watchList {
		s += fmt.Sprintf("%-10s %-5s %v\n", w.Symbol, w.Qty, w.Owned)
	}

	if m.err != nil {
		s += "\nError: " + m.err.Error()
	}
	s += "\nPress 1=edit expenses, 2=edit stonks, 3=edit watchlist, q=quit.\n"
	return s
}
