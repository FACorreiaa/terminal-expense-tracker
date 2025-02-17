package main

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
	expenses  []Expense
	stonks    []Stonk
	watchList []WatchItem
}

// model Bubbletea model
type model struct {
	expenses  []Expense
	stonks    []Stonk
	watchList []WatchItem
	err       error
}
