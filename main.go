package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	ltable "github.com/charmbracelet/lipgloss/table"
	"github.com/fsnotify/fsnotify"
	"github.com/xuri/excelize/v2"
)

type screen int

const (
	screenMenu screen = iota
	screenExpenses
	screenStonks
	screenWatchlist
)

var (
	editExpensesTitle = lipgloss.NewStyle().
		MarginLeft(1).
		MarginRight(5).
		Padding(0, 1).
		Bold(true).
		Italic(true).
		Foreground(lipgloss.Color("#FFF7DB")).
		SetString("Edit Expenses Title")

	mainMenu = lipgloss.NewStyle().
		MarginLeft(1).
		MarginRight(5).
		Padding(0, 1).
		Bold(true).
		Italic(true).
		Foreground(lipgloss.Color("#FFF7DB")).
		SetString("Edit Expenses Title")

	expansesMenuTitle = lipgloss.NewStyle().
		MarginLeft(1).
		MarginRight(5).
		Padding(0, 1).
		Bold(true).
		Italic(true).
		Foreground(lipgloss.Color("#FFF7DB")).
		SetString("Expenses")
)

// expenseEditedMsg now includes both an index and the updated expense.
type expenseEditedMsg struct {
	index   int
	expense Expense
}

// Expense Datastructures
type Expense struct {
	Name   string
	Amount float64
}
type Stonk struct {
	Symbol  string
	Change  float64
	Comment string
	Extra   float64
}
type WatchItem struct {
	Symbol string
	Qty    string
	Owned  bool
}

type excelDataMsg struct {
	expenses      []Expense
	stonks        []Stonk
	watchList     []WatchItem
	totalExpenses float64
}

// model is the Bubble Tea model.
type model struct {
	expenses      []Expense
	expensesTable *ltable.Table
	stonks        []Stonk
	watchList     []WatchItem
	err           error
	editing       bool
	currentScreen screen
	totalExpenses float64
}

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

func initialModel() model {
	data, err := readExcelData("data.xlsx")
	if err != nil {
		log.Printf("Error reading Excel data: %v", err)
		data = excelDataMsg{
			expenses:  []Expense{},
			stonks:    []Stonk{},
			watchList: []WatchItem{},
		}
	}

	m := model{
		currentScreen: screenMenu,
		expenses:      data.expenses,
		stonks:        data.stonks,
		watchList:     data.watchList,
		totalExpenses: data.totalExpenses,
	}
	m.updateExpensesTable()
	return m
}

// entry point
func main() {
	p := tea.NewProgram(initialModel())
	if err, _ := p.Run(); err != nil {
		log.Fatal(err)
	}
}

// --- File Watching & Excel Reading ---
func watchExcelCmd(filename string) tea.Cmd {
	return func() tea.Msg {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return errMsg{err}
		}
		defer watcher.Close()

		err = watcher.Add(filename)
		if err != nil {
			return errMsg{err}
		}

		for {
			select {
			case event := <-watcher.Events:
				if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
					time.Sleep(500 * time.Millisecond)
					data, err := readExcelData(filename)
					if err != nil {
						return errMsg{err}
					}
					return data
				}
			case err := <-watcher.Errors:
				return errMsg{err}
			}
		}
	}
}

func readExcelData(filename string) (excelDataMsg, error) {
	f, err := excelize.OpenFile(filename)
	if err != nil {
		return excelDataMsg{}, err
	}
	defer f.Close()

	// read each sheet
	expenses, err := readExpenses(f)
	if err != nil {
		return excelDataMsg{}, err
	}
	stonks, err := readStonks(f)
	if err != nil {
		return excelDataMsg{}, err
	}
	watchList, err := readWatchList(f)
	if err != nil {
		return excelDataMsg{}, err
	}

	f.SetCellFormula("Expenses", "D2", "=SUM(B3:B9)")
	computed, _ := f.CalcCellValue("Expenses", "D2")
	fmt.Printf("Raw computed string: %q\n", computed)
	// Convert to float64
	total, _ := strconv.ParseFloat(computed, 64)

	return excelDataMsg{
		expenses:      expenses,
		stonks:        stonks,
		watchList:     watchList,
		totalExpenses: total,
	}, nil
}

// --- Interactive Editing with Huh ---
func readExpenses(f *excelize.File) ([]Expense, error) {
	rows, err := f.GetRows("Expenses")
	if err != nil {
		return nil, err
	}
	var expenses []Expense
	for i := 1; i < len(rows); i++ {
		line := rows[i]
		if len(line) < 2 {
			continue
		}
		name := line[0]
		amt, _ := strconv.ParseFloat(line[1], 64)
		expenses = append(expenses, Expense{Name: name, Amount: amt})
	}
	return expenses, nil
}
func readStonks(f *excelize.File) ([]Stonk, error) {
	rows, err := f.GetRows("Stonks")
	if err != nil {
		return nil, err
	}
	var stonks []Stonk
	for i := 1; i < len(rows); i++ {
		line := rows[i]
		if len(line) < 4 {
			continue
		}
		sym := line[0]
		chg, _ := strconv.ParseFloat(line[1], 64)
		cmt := line[2]
		ext, _ := strconv.ParseFloat(line[3], 64)
		stonks = append(stonks, Stonk{Symbol: sym, Change: chg, Comment: cmt, Extra: ext})
	}
	return stonks, nil
}
func readWatchList(f *excelize.File) ([]WatchItem, error) {
	rows, err := f.GetRows("WatchList")
	if err != nil {
		return nil, err
	}
	var items []WatchItem
	for i := 1; i < len(rows); i++ {
		line := rows[i]
		if len(line) < 3 {
			continue
		}
		symbol := line[0]
		qty := line[1]
		owned := (line[2] == "Yes")
		items = append(items, WatchItem{Symbol: symbol, Qty: qty, Owned: owned})
	}
	return items, nil
}

func writeExcelCmd(exp []Expense, st []Stonk, wl []WatchItem) tea.Cmd {
	return func() tea.Msg {
		// do the actual write
		err := writeExcelData("data.xlsx", exp, st, wl)
		if err != nil {
			return errMsg{err}
		}
		// Wait a moment so fsnotify sees the file change
		time.Sleep(500 * time.Millisecond)
		// Then read fresh data again
		data, err := readExcelData("data.xlsx")
		if err != nil {
			return errMsg{err}
		}
		return data
	}
}

func writeExcelData(filename string,
	expenses []Expense, stonks []Stonk, watchList []WatchItem) error {
	f, err := excelize.OpenFile(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// Overwrite rows for Expenses
	for i, e := range expenses {
		row := i + 2
		f.SetCellValue("Expenses", fmt.Sprintf("A%d", row), e.Name)
		f.SetCellValue("Expenses", fmt.Sprintf("B%d", row), e.Amount)
	}
	// Overwrite rows for Stonks
	for i, st := range stonks {
		row := i + 2
		f.SetCellValue("Stonks", fmt.Sprintf("A%d", row), st.Symbol)
		f.SetCellValue("Stonks", fmt.Sprintf("B%d", row), st.Change)
		f.SetCellValue("Stonks", fmt.Sprintf("C%d", row), st.Comment)
		f.SetCellValue("Stonks", fmt.Sprintf("D%d", row), st.Extra)
	}
	// Overwrite rows for WatchList
	for i, w := range watchList {
		row := i + 2
		f.SetCellValue("WatchList", fmt.Sprintf("A%d", row), w.Symbol)
		f.SetCellValue("WatchList", fmt.Sprintf("B%d", row), w.Qty)
		if w.Owned {
			f.SetCellValue("WatchList", fmt.Sprintf("C%d", row), "Yes")
		} else {
			f.SetCellValue("WatchList", fmt.Sprintf("C%d", row), "No")
		}
	}
	return f.Save()
}

// --- Bubble Tea Init, Update, & View ---
func (m model) Init() tea.Cmd {
	return watchExcelCmd("data.xlsx")
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case excelDataMsg:
		m.expenses = msg.expenses
		m.stonks = msg.stonks
		m.watchList = msg.watchList
		m.totalExpenses = msg.totalExpenses
		fmt.Printf("Received totalExpenses: %f\n", msg.totalExpenses)

		// Optionally update other data here.
		return m, watchExcelCmd("data.xlsx")

	case errMsg:
		m.err = msg.err
		return m, watchExcelCmd("data.xlsx")

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "e": // Enter editing mode
			if len(m.expenses) > 0 && !m.editing {
				m.editing = true
				// Launch edit form for the first expense for demonstration.
				return m, m.editExpenseForm(0)
			}
		case "n": // Press 'n' to create a new expense.
			if !m.editing {
				m.editing = true
				return m, m.newExpenseForm()
			}
		//case "1":
		//	// For example, you might write back to Excel here.
		//	return m, writeExcelCmd(m.expenses, m.stonks, m.watchList)
		case "1":
			m.currentScreen = screenExpenses
		case "2":
			m.currentScreen = screenStonks
		case "3":
			m.currentScreen = screenWatchlist
		case "b": // go back to the main menu
			m.currentScreen = screenMenu
		}
		return m, nil

	case expenseEditedMsg:
		// Update the edited expense, then exit editing mode.
		if msg.index == -1 {
			m.expenses = append(m.expenses, msg.expense)
		} else {
			m.expenses[msg.index] = msg.expense
		}
		m.updateExpensesTable()
		m.editing = false
		// update excel later here
		return m, writeExcelCmd(m.expenses, m.stonks, m.watchList)

	default:
		return m, nil
	}
}
func (m model) View() string {
	switch m.currentScreen {
	case screenMenu:
		return m.viewMenu()
	case screenExpenses:
		return m.viewExpenses()
	case screenStonks:
		return m.viewStonks()
	case screenWatchlist:
		return m.viewWatchlist()
	default:
		return "Unknown screen"
	}
}

func (m model) viewMenu() string {
	s := mainMenu.String()
	s += "\n1) Expenses\n"
	s += "2) Stonks\n"
	s += "3) Watchlist\n"
	s += "\nPress q to quit.\n"
	return s
}

func (m model) viewExpenses() string {
	s := "=== EXPENSES ===\n"
	for _, e := range m.expenses {
		s += fmt.Sprintf("%-20s %10.2f\n", e.Name, e.Amount)
	}
	fmt.Printf("LALLALALALA %.4f", m.totalExpenses)
	s += fmt.Sprintf("\nTOTAL: %.4f\n", m.totalExpenses)
	s += "\nPress 'b' to go back.\n"
	s += "\nPress 'e' to edit.\n"
	return s
}

func sumExpenses(expenses []Expense) float64 {
	var total float64
	for _, e := range expenses {
		total += e.Amount
	}
	return total
}

func (m model) viewStonks() string {
	s := "=== STONKS ===\n"
	// ...
	s += "\nPress 'b' to go back.\n"
	return s
}

func (m model) viewWatchlist() string {
	s := "=== WATCHLIST ===\n"
	// ...
	s += "\nPress 'b' to go back.\n"
	return s
}

func (m *model) updateExpensesTable() {
	headers := []string{"#", "Expense", "Amount"}

	var data [][]string
	for i, e := range m.expenses {
		row := []string{strconv.Itoa(i + 1), e.Name, fmt.Sprintf("%.2f", e.Amount)}
		data = append(data, row)
	}

	// Change style later
	re := lipgloss.NewRenderer(os.Stdout)
	baseStyle := re.NewStyle().Padding(0, 1)
	headerStyle := baseStyle.Foreground(lipgloss.Color("252")).Bold(true)
	rowStyle := baseStyle.Foreground(lipgloss.Color("252"))

	t := ltable.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(re.NewStyle().Foreground(lipgloss.Color("238"))).
		Headers(headers...).
		Width(80).
		Rows(data...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == ltable.HeaderRow {
				return headerStyle
			}
			// Alternate row colors.
			if row%2 == 0 {
				return rowStyle.Foreground(lipgloss.Color("245"))
			}
			return rowStyle
		})

	m.expensesTable = t
}

func (m *model) editExpenseForm(index int) tea.Cmd {
	// Pre-fill form fields with the current expense data.
	var newName string = m.expenses[index].Name
	var newAmount string = fmt.Sprintf("%.2f", m.expenses[index].Amount)

	// Wrap inputs in a group (required by Huh).
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Expense Name").Value(&newName),
			huh.NewInput().Title("Amount").Value(&newAmount),
		),
	)

	return func() tea.Msg {
		if err := form.Run(); err != nil {
			return errMsg{err}
		}
		// Convert the amount to float64.
		amt, err := strconv.ParseFloat(newAmount, 64)
		if err != nil {
			return errMsg{err}
		}
		updated := Expense{Name: newName, Amount: amt}
		return expenseEditedMsg{index: index, expense: updated}
	}
}

func (m *model) newExpenseForm() tea.Cmd {
	// Default empty values for a new expense.
	var newName string = ""
	var newAmount string = "0.00"

	// Wrap inputs in a group using Huh.
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Expense Name").Value(&newName),
			huh.NewInput().Title("Amount").Value(&newAmount),
		),
	)

	return func() tea.Msg {
		if err := form.Run(); err != nil {
			return errMsg{err}
		}
		// Convert the amount to float64.
		amt, err := strconv.ParseFloat(newAmount, 64)
		if err != nil {
			return errMsg{err}
		}
		updated := Expense{Name: newName, Amount: amt}
		// Use index -1 to signal this is a new expense.
		return expenseEditedMsg{index: -1, expense: updated}
	}
}
