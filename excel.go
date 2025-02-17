package main

import (
	"fmt"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
	"github.com/xuri/excelize/v2"
)

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

	return excelDataMsg{
		expenses:  expenses,
		stonks:    stonks,
		watchList: watchList,
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

//func getStockOptions(stocks map[string]stockData) []huh.Option[string] {
//	options := make([]huh.Option[string], 0, len(stocks))
//	for name := range stocks {
//		options = append(options, huh.NewOption(name, name))
//	}
//	return options
//}

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

	// Clear or overwrite rows
	// For brevity, just overwrite row-by-row
	for i, e := range expenses {
		row := i + 2
		f.SetCellValue("Expenses", fmt.Sprintf("A%d", row), e.Name)
		f.SetCellValue("Expenses", fmt.Sprintf("B%d", row), e.Amount)
	}
	for i, st := range stonks {
		row := i + 2
		f.SetCellValue("Stonks", fmt.Sprintf("A%d", row), st.Symbol)
		f.SetCellValue("Stonks", fmt.Sprintf("B%d", row), st.Change)
		f.SetCellValue("Stonks", fmt.Sprintf("C%d", row), st.Comment)
		f.SetCellValue("Stonks", fmt.Sprintf("D%d", row), st.Extra)
	}
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
