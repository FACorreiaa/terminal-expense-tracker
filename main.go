package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	ltable "github.com/charmbracelet/lipgloss/table"
	"github.com/fsnotify/fsnotify"
	"github.com/xuri/excelize/v2"
)

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
		SetString("Expenses")

	expansesMenuTitle = lipgloss.NewStyle().
		MarginLeft(1).
		MarginRight(5).
		Padding(0, 1).
		Bold(true).
		Italic(true).
		Foreground(lipgloss.Color("#FFF7DB")).
		SetString("Expenses")

	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type menuItem string

func (m menuItem) Title() string       { return string(m) }
func (m menuItem) Description() string { return "" }
func (m menuItem) FilterValue() string { return string(m) }

type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(menuItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type screen int

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
	list          list.Model
	selectedRow   int
}

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

func initialModel() *model {
	data, err := readExcelData("data.xlsx")
	if err != nil {
		log.Printf("Error reading Excel data: %v", err)
		data = excelDataMsg{
			expenses:  []Expense{},
			stonks:    []Stonk{},
			watchList: []WatchItem{},
			// Ensure a default total if needed:
			totalExpenses: 0,
		}
	}

	// Create menu items.
	items := []list.Item{
		menuItem("Expenses"),
		menuItem("Stonks"),
		menuItem("Watchlist"),
	}

	// Create the list model. Adjust the width and height as needed.
	l := list.New(items, itemDelegate{}, 20, 7)
	l.Title = "Main Menu"
	l.SetFilteringEnabled(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)

	m := model{
		currentScreen: screenMenu,
		expenses:      data.expenses,
		stonks:        data.stonks,
		watchList:     data.watchList,
		totalExpenses: data.totalExpenses,
		list:          l,
		editing:       false,
	}
	m.updateExpensesTable()
	return &m
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
	total, _ := strconv.ParseFloat(computed, 64)

	return excelDataMsg{
		expenses:      expenses,
		stonks:        stonks,
		watchList:     watchList,
		totalExpenses: total,
	}, nil
}

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
		err := writeExcelData("data.xlsx", exp, st, wl)
		if err != nil {
			return errMsg{err}
		}
		time.Sleep(500 * time.Millisecond)
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

// Init --- Bubble Tea Init, Update, & View ---
func (m *model) Init() tea.Cmd {
	return watchExcelCmd("data.xlsx")
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case excelDataMsg:
		m.expenses = msg.expenses
		m.stonks = msg.stonks
		m.watchList = msg.watchList
		m.totalExpenses = msg.totalExpenses
		return m, watchExcelCmd("data.xlsx")
	case errMsg:
		m.err = msg.err
		return m, watchExcelCmd("data.xlsx")
	}

	if m.currentScreen == screenMenu {
		m.list, cmd = m.list.Update(msg)
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "enter":
				selected := m.list.SelectedItem().(menuItem)
				fmt.Println("You selected:", selected)
				switch selected {
				case "Expenses":
					m.currentScreen = screenExpenses
				case "Stonks":
					m.currentScreen = screenStonks
				case "Watchlist":
					m.currentScreen = screenWatchlist
				}
			}
		}
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up":
			if m.selectedRow > 0 {
				m.selectedRow--
				m.updateExpensesTable()
			}
		case "down":
			if m.selectedRow < len(m.expenses)-1 {
				m.selectedRow++
				m.updateExpensesTable()

			}
		case "b":
			m.currentScreen = screenMenu
			return m, nil
		case "e":
			if m.currentScreen == screenExpenses && !m.editing && len(m.expenses) > 0 {
				m.editing = true
				return m, m.editExpenseForm(m.selectedRow)
			}
		case "n":
			if m.currentScreen == screenExpenses && !m.editing {
				m.editing = true
				return m, m.newExpenseForm()
			}
		}
	case expenseEditedMsg:
		if msg.index == -1 {
			m.expenses = append(m.expenses, msg.expense)
		} else {
			m.expenses[msg.index] = msg.expense
		}
		m.updateExpensesTable()
		m.editing = false
		m.currentScreen = screenExpenses

		return m, writeExcelCmd(m.expenses, m.stonks, m.watchList)
	}

	return m, nil
}

func (m *model) View() string {
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

func (m *model) viewMenu() string {
	return m.list.View() + "\nPress q to quit.\n"
}

func (m *model) viewExpenses() string {
	var buffer bytes.Buffer
	buffer.WriteString("\n")
	buffer.WriteString(editExpensesTitle.String())
	buffer.WriteString("\n")
	buffer.WriteString(m.expensesTable.String())

	buffer.WriteString("\nUse ↑/↓ to move, 'e' to edit the selected row, 'n' to insert a new expense, 'q' to quit.\n")
	buffer.WriteString("\nPress 'b' to go back.\n")
	buffer.WriteString("\nPress 'e' to edit.\n")
	buffer.WriteString("\nPress 'n' to insert new expense.\n")

	return buffer.String()
}

func (m *model) viewStonks() string {
	s := "=== STONKS ===\n"
	// ...
	s += "\nPress 'b' to go back.\n"
	return s
}

func (m *model) viewWatchlist() string {
	s := "=== WATCHLIST ===\n"
	// ...
	s += "\nPress 'b' to go back.\n"
	return s
}

func (m *model) updateExpensesTable() {
	headers := []string{"#", "Expense", "Amount"}

	var data [][]string
	for i, e := range m.expenses {
		// i+1 is row number for display
		row := []string{strconv.Itoa(i + 1), e.Name, fmt.Sprintf("%.2f", e.Amount)}
		data = append(data, row)
	}

	// Base styles
	re := lipgloss.NewRenderer(os.Stdout)
	baseStyle := re.NewStyle().Padding(0, 1)
	headerStyle := baseStyle.Foreground(lipgloss.Color("252")).Bold(true)
	rowStyle := baseStyle.Foreground(lipgloss.Color("252"))

	// Define a highlight style for the selected row
	highlightStyle := baseStyle.
		Background(lipgloss.Color("57")).
		Foreground(lipgloss.Color("229")).
		Bold(true)

	// Build the table
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
			if row == m.selectedRow {
				return highlightStyle
			}

			if row%2 == 0 {
				return rowStyle.Foreground(lipgloss.Color("245"))
			}
			return rowStyle
		})

	m.expensesTable = t
}

func (m *model) editExpenseForm(index int) tea.Cmd {
	var newName string = m.expenses[index].Name
	var newAmount string = fmt.Sprintf("%.2f", m.expenses[index].Amount)

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
		amt, err := strconv.ParseFloat(newAmount, 64)
		if err != nil {
			return errMsg{err}
		}
		updated := Expense{Name: newName, Amount: amt}

		return expenseEditedMsg{index: index, expense: updated}
	}
}

func (m *model) newExpenseForm() tea.Cmd {
	var newName string = ""
	var newAmount string = "0.00"

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
		amt, err := strconv.ParseFloat(newAmount, 64)
		if err != nil {
			return errMsg{err}
		}
		updated := Expense{Name: newName, Amount: amt}
		return expenseEditedMsg{index: -1, expense: updated}
	}
}
