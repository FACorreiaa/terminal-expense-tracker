# Stock Terminal Dashboard

A terminal application built in Go that displays and allows you to interactively edit stock investment data while keeping it in sync with an underlying Excel file. This app leverages [Bubble Tea](https://github.com/charmbracelet/bubbletea) for a rich terminal UI, [excelize](https://github.com/xuri/excelize) for reading from Excel, and [fsnotify](https://github.com/fsnotify/fsnotify) to watch for changes in the Excel file.

## Features

- **Live Excel Sync:** Automatically updates the terminal view when changes are made to the Excel file.
- **Interactive Editing:** Use the terminal interface (powered by Bubble Tea and optionally Huh) to update stock data and budgets.
- **Data Management:** Displays stock data including budget, change values, comments, and extra information.
- **Error Reporting:** Displays error messages if the Excel file cannot be read or if other issues occur.

## Requirements

- **Go:** Version 1.16 or higher is recommended.
- **Excel File:** An Excel file (e.g., `data.xlsx`) formatted as follows:
    - **Sheet1:**
        - Budget value in cell `B2`
        - Stock data starting from row 3 with columns:
            - **A:** Stock Name
            - **B:** Change value
            - **C:** Comment
            - **D:** Extra value
- **Dependencies:**
    - [Bubble Tea](https://github.com/charmbracelet/bubbletea)
    - [excelize](https://github.com/xuri/excelize/v2)
    - [fsnotify](https://github.com/fsnotify/fsnotify)
    - Optionally, [Huh](https://github.com/charmbracelet/huh) for enhanced components

