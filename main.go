package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	bubbletea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/fsnotify/fsnotify"
	"github.com/xuri/excelize/v2"
)

// --- Custom Message Types ---

type excelDataMsg struct {
	budget int
	stocks map[string]stockData
}

type stockData struct {
	change  string
	comment string
	extra   string
}

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

// --- Bubble Tea Model ---

type model struct {
	budget int
	stocks map[string]stockData
	err    error
	form   *huh.Form // Form for editing stock data
}

func initialModel() model {
	return model{
		stocks: make(map[string]stockData),
	}
}

// --- Main ---

func main() {
	p := bubbletea.NewProgram(initialModel())
	if err, _ := p.Run(); err != nil {
		log.Fatal(err)
	}
}

// --- Bubble Tea Init, Update, & View ---

func (m model) Init() bubbletea.Cmd {
	return watchExcelCmd("data.xlsx")
}

func (m model) Update(msg bubbletea.Msg) (bubbletea.Model, bubbletea.Cmd) {
	switch msg := msg.(type) {

	case excelDataMsg:
		m.budget = msg.budget
		m.stocks = msg.stocks
		return m, watchExcelCmd("data.xlsx")

	case errMsg:
		m.err = msg.err
		return m, watchExcelCmd("data.xlsx")

	case bubbletea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, bubbletea.Quit
		case "e": // Press 'e' to edit stock data
			return m, m.editStockCmd()
		}
		return m, nil

	default:
		return m, nil
	}
}

func (m model) View() string {
	s := "Stock Data:\n\n"
	s += fmt.Sprintf("Budget: %d\n\n", m.budget)
	s += "Stock\tChange\tComment\tExtra\n"
	for name, data := range m.stocks {
		s += fmt.Sprintf("%-8s%-8d%-16s%-8d\n", name, data.change, data.comment, data.extra)
	}
	if m.err != nil {
		s += "\nError: " + m.err.Error()
	}
	s += "\nPress 'e' to edit, 'q' to quit.\n"
	return s
}

// --- File Watching & Excel Reading ---

func watchExcelCmd(filename string) bubbletea.Cmd {
	return func() bubbletea.Msg {
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

	budgetStr, err := f.GetCellValue("Sheet1", "B2")
	if err != nil {
		return excelDataMsg{}, err
	}
	budget, err := strconv.Atoi(budgetStr)
	if err != nil {
		budget = 0
	}

	stocks := make(map[string]stockData)
	for row := 3; row <= 10; row++ {
		name, err := f.GetCellValue("Sheet1", fmt.Sprintf("A%d", row))
		if err != nil || name == "" {
			continue
		}
		changeStr, _ := f.GetCellValue("Sheet1", fmt.Sprintf("B%d", row))
		change, _ := strconv.Atoi(changeStr)
		comment, _ := f.GetCellValue("Sheet1", fmt.Sprintf("C%d", row))
		extraStr, _ := f.GetCellValue("Sheet1", fmt.Sprintf("D%d", row))
		extra, _ := strconv.Atoi(extraStr)
		stocks[name] = stockData{
			change:  strconv.Itoa(change),
			comment: comment,
			extra:   strconv.Itoa(extra),
		}
	}

	return excelDataMsg{
		budget: budget,
		stocks: stocks,
	}, nil
}

// --- Interactive Editing with Huh ---

func (m *model) editStockCmd() bubbletea.Cmd {
	var selectedStock string
	var change string
	var comment string
	var extra string

	// Create a form for editing stock data
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select Stock").
				Options(getStockOptions(m.stocks)...),
			huh.NewInput().
				Title("Change").
				Value(&change),
			huh.NewInput().
				Title("Comment").
				Value(&comment),
			huh.NewInput().
				Title("Extra").
				Value(&extra),
		))

	return func() bubbletea.Msg {
		err := form.Run()
		if err != nil {
			return errMsg{err}
		}

		// Update the stock data in the model
		m.stocks[selectedStock] = stockData{
			change:  change,
			comment: comment,
			extra:   extra,
		}

		// Write the updated data back to the Excel file
		err = writeExcelData("data.xlsx", m.budget, m.stocks)
		if err != nil {
			return errMsg{err}
		}

		return nil
	}
}

func getStockOptions(stocks map[string]stockData) []huh.Option[string] {
	options := make([]huh.Option[string], 0, len(stocks))
	for name := range stocks {
		options = append(options, huh.NewOption(name, name))
	}
	return options
}

func writeExcelData(filename string, budget int, stocks map[string]stockData) error {
	f, err := excelize.OpenFile(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write budget
	f.SetCellValue("Sheet1", "B2", budget)

	// Write stock data
	row := 3
	for name, data := range stocks {
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", row), name)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", row), data.change)
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", row), data.comment)
		f.SetCellValue("Sheet1", fmt.Sprintf("D%d", row), data.extra)
		row++
	}

	return f.Save()
}
