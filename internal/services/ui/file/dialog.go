package file

import (
	"fmt"
	"gosheet/internal/services/cell"
	"gosheet/internal/utils"
	"os"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ShowUnifiedFileDialog displays a unified file browser/selector for open and save operations
func ShowUnifiedFileDialog(app *tview.Application, returnTo tview.Primitive, mode string, globalData map[[2]int]*cell.Cell, table *tview.Table, SetCurrentFilename func(table *tview.Table, filename string), MarkAsSaved func(table *tview.Table), HasUnsavedChanges func() bool, currentFilename string) {
	homeDir, _ := os.UserHomeDir()
	docsPath := filepath.Join(homeDir, "Documents")
	if _, err := os.Stat(docsPath); os.IsNotExist(err) {
		docsPath = homeDir
	}
	
	currentPath := docsPath
	if mode == "save" {
		currentPath = filepath.Dir(currentFilename)
	}
	var selectedFormat int = 0

	quickAccessList := tview.NewList()
	quickAccessList.SetBorder(true).
		SetTitle(" Quick Access ").
		SetBorderColor(tcell.ColorLightBlue)
	quickAccessList.SetSelectedBackgroundColor(tcell.ColorDarkCyan)
	quickAccessList.SetSelectedTextColor(tcell.ColorWhite)		

	// File list
	fileList := tview.NewList()
	fileList.SetBorder(true).
		SetTitle(" Files ").
		SetBorderColor(tcell.ColorLightBlue)
	fileList.SetSelectedBackgroundColor(tcell.ColorDarkCyan)
	fileList.SetSelectedTextColor(tcell.ColorWhite)
	
	pathLabel := tview.NewTextView().
		SetDynamicColors(true).
		SetText(fmt.Sprintf("[yellow]Location:[::-] %s", utils.PrettyPath(filepath.Dir(currentPath), "no")))
	pathLabel.SetBorder(true).SetBorderColor(tcell.ColorLightBlue)
	
	form := tview.NewForm()
	
	filenameInput := tview.NewInputField().
		SetLabel("Filename: ").
		SetFieldWidth(0)
	
	if mode == "save" {
		filenameInput.SetText(strings.TrimSpace(filepath.Base(currentFilename)))
	}
	
	form.AddFormItem(filenameInput)
	
	var formatDropdown *tview.DropDown
	if mode == "save" {
		formatOptions := make([]string, len(FileFormats))
		for i, format := range FileFormats {
			formatOptions[i] = fmt.Sprintf("%s - %s", format.Extension, format.Description)
		}
		
		formatDropdown = tview.NewDropDown().
			SetLabel("Format: ").
			SetOptions(formatOptions, func(option string, index int) {
				selectedFormat = index
				currentName := filenameInput.GetText()
				
				for _, fmt := range FileFormats {
					if strings.HasSuffix(currentName, fmt.Extension) {
						currentName = strings.TrimSuffix(currentName, fmt.Extension)
						break
					}
				}
				
				filenameInput.SetText(currentName + FileFormats[index].Extension)
			}).
			SetCurrentOption(0)
		
		form.AddFormItem(formatDropdown)
	}
	
	var updateList func(path string)
	updateList = func(path string) {
		fileList.Clear()
		currentPath = path
		pathLabel.SetText(fmt.Sprintf("[yellow]Location:[::-] %s", utils.PrettyPath(path, "no")))
		
		if path != filepath.Dir(path) {
			fileList.AddItem("..", "Parent directory", '↑', func() {
				updateList(filepath.Dir(path))
			})
		}
		
		entries, err := os.ReadDir(path)
		if err != nil {
			return
		}
		
		for _, entry := range entries {
			if entry.IsDir() {
				name := entry.Name()
				fullPath := filepath.Join(path, name)
				fileList.AddItem(fmt.Sprintf("# %s", name), "Directory", 0, func() {
					updateList(fullPath)
				})
			}
		}
		
		for _, entry := range entries {
			if !entry.IsDir() {
				name := entry.Name()
				ext := strings.ToLower(filepath.Ext(name))
				
				if mode == "open" {
					if ext != ".gsheet" && ext != ".json" && ext != ".txt" && ext != ".xlsx" {
						continue
					}
				}
				
				//fullPath := filepath.Join(path, name)
				info, _ := entry.Info()
				var sizeStr, icon string
				
				if info != nil {
					size := info.Size()
					if size < 1024 {
						sizeStr = fmt.Sprintf("%d B", size)
					} else if size < 1024*1024 {
						sizeStr = fmt.Sprintf("%.1f KB", float64(size)/1024)
					} else {
						sizeStr = fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
					}
				}
				
				switch ext {
				case ".gsheet":
					icon = ""
				case ".json":
					icon = ""
				case ".txt":
					icon = ""
				case ".csv":
					icon = ""
				case ".html":
					icon = ""
				case ".xlsx":
					icon = ""
				default:
					icon = ""
				}
				
				displayName := fmt.Sprintf("%s %s", icon, name)
				
				fileList.AddItem(displayName, sizeStr, 0, func() {
					filenameInput.SetText(filepath.Join(path, name))
				})
			}
		}
	}
	populateQuickAccess := func() {
		quickAccessList.Clear()
		
		quickAccessList.AddItem("Home", homeDir, 'h', func() {
			updateList(homeDir)
		})
		
		if _, err := os.Stat(docsPath); err == nil {
			quickAccessList.AddItem("Documents", docsPath, 'd', func() {
				updateList(docsPath)
			})
		}
		
		desktopPath := filepath.Join(homeDir, "Desktop")
		if _, err := os.Stat(desktopPath); err == nil {
			quickAccessList.AddItem("Desktop", desktopPath, 0, func() {
				updateList(desktopPath)
			})
		}
		
		downloadsPath := filepath.Join(homeDir, "Downloads")
		if _, err := os.Stat(downloadsPath); err == nil {
			quickAccessList.AddItem("↓ Downloads", downloadsPath, 0, func() {
				updateList(downloadsPath)
			})
		}
		
		if cwd, err := os.Getwd(); err == nil {
			quickAccessList.AddItem("Current Dir", cwd, 'c', func() {
				updateList(cwd)
			})
		}
		
		if filepath.Separator == '\\' {
			for _, drive := range []string{"C:", "D:", "E:", "F:", "G:"} {
				drivePath := drive + "\\"
				if _, err := os.Stat(drivePath); err == nil {
					driveLabel := fmt.Sprintf("Drive %s", drive)
					capturedPath := drivePath
					quickAccessList.AddItem(driveLabel, drivePath, 0, func() {
						updateList(capturedPath)
					})
				}
			}
		} else {
			quickAccessList.AddItem("Root (/)", "/", 0, func() {
				updateList("/")
			})
			
			mountPoints := []string{"/mnt", "/media", "/Volumes"}
			for _, mount := range mountPoints {
				if entries, err := os.ReadDir(mount); err == nil && len(entries) > 0 {
					quickAccessList.AddItem(fmt.Sprintf("%s", filepath.Base(mount)), mount, 0, func() {
						updateList(mount)
					})
					
					for _, entry := range entries {
						if entry.IsDir() {
							volumePath := filepath.Join(mount, entry.Name())
							capturedPath := volumePath
							quickAccessList.AddItem(fmt.Sprintf("%s", entry.Name()), volumePath, 0, func() {
								updateList(capturedPath)
							})
						}
					}
				}
			}
		}
	}

	populateQuickAccess()
	updateList(currentPath)
	
	if mode == "save" {
		form.AddButton("Save", func() {
			filename := strings.TrimSpace(filenameInput.GetText())
			if filename == "" {
				ShowErrorModal(app, returnTo, "Filename cannot be empty!")
				return
			}
			
			if !strings.Contains(filename, string(filepath.Separator)) {
				filename = filepath.Join(currentPath, filename)
			}
			
			if strings.HasPrefix(filename, "~") {
				filename = filepath.Join(homeDir, filename[1:])
			}
			
			absPath, err := filepath.Abs(filename)
			if err == nil {
				filename = absPath
			}
			
			selectedFormat := FileFormats[selectedFormat]
			if !strings.HasSuffix(strings.ToLower(filename), selectedFormat.Extension) {
				filename += selectedFormat.Extension
			}
			
			dir := filepath.Dir(filename)
			if err := os.MkdirAll(dir, 0755); err != nil {
				ShowErrorModal(app, returnTo, fmt.Sprintf("Cannot create directory: %s", err))
				return
			}
			
			if fileExists(filename) {
				showOverwriteConfirmation(app, table, filename, selectedFormat, false, globalData)
			} else {
				performSave(app, table, filename, selectedFormat, false, globalData)
			}

			SetCurrentFilename(table, filename)
			MarkAsSaved(table)
		})
		
		form.AddButton("Save & Exit", func() {
			filename := strings.TrimSpace(filenameInput.GetText())
			if filename == "" {
				ShowErrorModal(app, returnTo, "Filename cannot be empty!")
				return
			}
			
			if !strings.Contains(filename, string(filepath.Separator)) {
				filename = filepath.Join(currentPath, filename)
			}
			
			if strings.HasPrefix(filename, "~") {
				filename = filepath.Join(homeDir, filename[1:])
			}
			
			absPath, err := filepath.Abs(filename)
			if err == nil {
				filename = absPath
			}
			
			selectedFormat := FileFormats[selectedFormat]
			if !strings.HasSuffix(strings.ToLower(filename), selectedFormat.Extension) {
				filename += selectedFormat.Extension
			}
			
			dir := filepath.Dir(filename)
			if err := os.MkdirAll(dir, 0755); err != nil {
				ShowErrorModal(app, returnTo, fmt.Sprintf("Cannot create directory: %s", err))
				return
			}
			
			if fileExists(filename) {
				showOverwriteConfirmation(app, table, filename, selectedFormat, true, globalData)
			} else {
				performSave(app, table, filename, selectedFormat, true, globalData)
			}
		})
	} else {
		form.AddButton("Open", func() {
			filename := strings.TrimSpace(filenameInput.GetText())
			if filename == "" {
				ShowErrorModal(app, returnTo, "Filename cannot be empty!")
				return
			}
			
			if !strings.Contains(filename, string(filepath.Separator)) {
				filename = filepath.Join(currentPath, filename)
			}
			
			if strings.HasPrefix(filename, "~") {
				filename = filepath.Join(homeDir, filename[1:])
			}
			
			absPath, err := filepath.Abs(filename)
			if err == nil {
				filename = absPath
			}
			
			if _, err := os.Stat(filename); os.IsNotExist(err) {
				ShowErrorModal(app, returnTo, fmt.Sprintf("File not found:\n%s", filename))
				return
			}
			
			ext := strings.ToLower(filepath.Ext(filename))
			if ext != ".gsheet" && ext != ".json" && ext != ".txt" && ext != ".xlsx" {
				ShowErrorModal(app, returnTo, fmt.Sprintf("Unsupported file format: %s\n\nSupported: .gsheet, .json, .txt, .xlsx", ext))
				return
			}
			
			selectFile(app, filename)
		})
	}
	
	form.AddButton("Cancel", func() {
		app.SetRoot(returnTo, true).SetFocus(returnTo)
	})

	form.AddButton("Exit", func() {
		if HasUnsavedChanges() && mode == "save"{
		modal := tview.NewModal().
			SetText("You have unsaved changes.\n\nWhat would you like to do?").
			AddButtons([]string{"Save", "Discard", "Cancel"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				switch buttonLabel {
				case "Save":
					ShowUnifiedFileDialog(app, table, "save", globalData, table, SetCurrentFilename, MarkAsSaved, HasUnsavedChanges, currentFilename)
				case "Discard":
					app.Stop()
				case "Cancel":
					app.SetRoot(table, true).SetFocus(table)
				}
			})
		modal.SetBorder(true).SetTitle(" Unsaved Changes ").SetBorderColor(tcell.ColorYellow)
		app.SetRoot(modal, true).SetFocus(modal)
		} else { app.Stop() }
	})
	
	form.SetBorder(true).SetBorderColor(tcell.ColorYellow)
	if mode == "save" {
		form.SetTitle(" Options ")
	} else {
		form.SetTitle(" Open File ")
	}
	
	// Layout
	leftSection := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(form, 0, 1, true).
		AddItem(quickAccessList, 0, 2, false)
	
	topSection := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(pathLabel, 3, 0, false).
		AddItem(leftSection, 0, 1, true)
	
	mainLayout := tview.NewFlex().
		AddItem(topSection, 0, 1, true).
		AddItem(fileList, 0, 1, false)

	instructions := tview.NewTextView().
		SetText(" [yellow::b]Ctrl+←/→[::-] Switch Panel  [yellow::b]Enter[::-] Select  [yellow::b]Tab/Shift+Tab[::-] Navigate form buttons  [yellow::b]Esc[::-] Cancel").
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	
	container := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(mainLayout, 0, 1, true).
		AddItem(instructions, 1, 0, false)
	
	container.SetBorder(true).SetBorderColor(tcell.ColorLightBlue)
	if mode == "save" {
		container.SetTitle(" Save Spreadsheet ")
	} else {
		container.SetTitle(" Open Spreadsheet ")
	}
	
	focusables := []tview.Primitive{quickAccessList, fileList, form}
	currentFocus := 2
	
	container.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case event.Key() == tcell.KeyEscape:
			app.SetRoot(returnTo, true).SetFocus(returnTo)
			return nil
		case event.Modifiers()&tcell.ModCtrl != 0 && event.Key() == tcell.KeyRight:
			currentFocus = (currentFocus + 1) % len(focusables)
			app.SetFocus(focusables[currentFocus])
			return nil
		case event.Modifiers()&tcell.ModCtrl != 0 && event.Key() == tcell.KeyLeft:
			currentFocus = (currentFocus - 1 + len(focusables)) % len(focusables)
			app.SetFocus(focusables[currentFocus])
			return nil
		}
		return event
	})

	app.SetRoot(container, true).SetFocus(form)
}
