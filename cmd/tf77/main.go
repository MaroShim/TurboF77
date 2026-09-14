package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gdamore/tcell/v2"
	"tf77/internal/compiler"
	"tf77/internal/debugger"
	"tf77/internal/sound"
	"tf77/internal/ui"
	"tf77/internal/ui/dialogs"
)

// matchMenuHotKey returns the menu index (0..8) for a given hotkey rune, or -1 if not matched.
func matchMenuHotKey(ch rune) int {
	switch ch {
	case 'f', 'F': // File
		return 0
	case 'e', 'E': // Edit
		return 1
	case 's', 'S': // Search
		return 2
	case 'r', 'R': // Run
		return 3
	case 'c', 'C': // Compile
		return 4
	case 'd', 'D': // Debug
		return 5
	case 'o', 'O': // Options
		return 6
	case 'w', 'W': // Window
		return 7
	case 'h', 'H': // Help
		return 8
	default:
		return -1
	}
}

func main() {
	var initialFile string
	if len(os.Args) > 1 {
		initialFile = os.Args[1]
	}

	app, err := ui.NewApp(initialFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing Turbo F77: %v\n", err)
		os.Exit(1)
	}
	defer app.Stop()

	// Initialize dialogs
	compileDlg := dialogs.NewCompileDialog()
	errListDlg := dialogs.NewErrorListDialog()
	openDlg := dialogs.NewOpenFileDialog()
	saveDlg := dialogs.NewSaveFileDialog()
	aboutDlg := dialogs.NewAboutDialog()
	gotoDlg := dialogs.NewGotoLineDialog()
	findDlg := dialogs.NewFindDialog()
	searchResDlg := dialogs.NewSearchResultsDialog()

	app.SetDialogs(compileDlg, errListDlg, openDlg, saveDlg, aboutDlg, gotoDlg)
	app.SetFindDialog(findDlg)
	app.SetSearchResultsDialog(searchResDlg)

	screen := app.Screen()
	editor := app.GetEditor()
	userScreen := app.GetUserScreen()

	// Action dispatcher
	var dispatchAction func(actionID string)
	dispatchAction = func(actionID string) {
		switch actionID {
		case "file_new":
			*editor = *ui.NewEditor("", editor.WindowNumber)
		case "file_open":
			openDlg.Show(".", func(path string) {
				if err := editor.LoadFile(path); err != nil {
					sound.PlayError()
					app.SetStatusMessage("Error opening " + filepath.Base(path) + ": " + err.Error())
				} else {
					app.SetStatusMessage("Opened " + editor.FileName)
				}
			})
		case "file_save":
			if editor.FilePath == "" || editor.FilePath == "NONAME00.FOR" {
				saveDlg.Show("main.for", func(path string) {
					if err := editor.SaveAs(path); err != nil {
						sound.PlayError()
						app.SetStatusMessage("Error saving " + filepath.Base(path) + ": " + err.Error())
					} else {
						sound.PlaySuccess()
						app.SetStatusMessage("Saved " + editor.FileName)
					}
				})
			} else {
				if err := editor.SaveFile(); err != nil {
					sound.PlayError()
					app.SetStatusMessage("Error saving " + editor.FileName + ": " + err.Error())
				} else {
					sound.PlaySuccess()
					app.SetStatusMessage("Saved " + editor.FileName)
				}
			}
		case "file_save_as":
			defaultName := editor.FileName
			if defaultName == "" || defaultName == "NONAME00.FOR" {
				defaultName = "main.for"
			}
			saveDlg.Show(defaultName, func(path string) {
				if err := editor.SaveAs(path); err != nil {
					sound.PlayError()
					app.SetStatusMessage("Error saving " + filepath.Base(path) + ": " + err.Error())
				} else {
					sound.PlaySuccess()
					app.SetStatusMessage("Saved " + editor.FileName)
				}
			})
		case "app_exit":
			app.Stop()
			os.Exit(0)
		case "run_run":
			bRes, _ := app.RunCurrent()
			if !bRes.Success {
				sound.PlayError()
				compileDlg.Show(editor.FileName, bRes.LinesCompiled, bRes)
			} else {
				sound.PlaySuccess()
			}
		case "run_userscreen":
			app.SyncDebuggerState()
			userScreen.Show()
		case "compile_compile", "compile_make", "compile_buildall":
			bRes := app.CompileCurrent()
			if bRes.Success {
				sound.PlaySuccess()
			} else {
				sound.PlayError()
			}
			compileDlg.Show(editor.FileName, bRes.LinesCompiled, bRes)
		case "debug_continue":
			if !app.GetDebugger().IsActive() {
				err := app.StartDebugging()
				if err != nil {
					sound.PlayError()
					app.SetStatusMessage(fmt.Sprintf("Debug start error: %v", err))
				} else {
					sound.PlayBreakpoint()
				}
			} else {
				if err := app.DebugContinue(); err != nil {
					sound.PlayError()
					app.SetStatusMessage("Debug error: " + err.Error())
				} else {
					sound.PlayBreakpoint()
				}
			}
		case "debug_step_over":
			if !app.GetDebugger().IsActive() {
				err := app.StartDebugging()
				if err != nil {
					sound.PlayError()
					app.SetStatusMessage(fmt.Sprintf("Debug start error: %v", err))
				} else {
					sound.PlayBreakpoint()
				}
			} else {
				if err := app.DebugStepOver(); err != nil {
					sound.PlayError()
					app.SetStatusMessage("Debug error: " + err.Error())
				} else {
					sound.PlayBreakpoint()
				}
			}
		case "debug_step_into":
			if !app.GetDebugger().IsActive() {
				err := app.StartDebugging()
				if err != nil {
					sound.PlayError()
					app.SetStatusMessage(fmt.Sprintf("Debug start error: %v", err))
				} else {
					sound.PlayBreakpoint()
				}
			} else {
				if err := app.DebugStepInto(); err != nil {
					sound.PlayError()
					app.SetStatusMessage("Debug error: " + err.Error())
				} else {
					sound.PlayBreakpoint()
				}
			}
		case "debug_stop":
			app.StopDebugging()
		case "debug_watches":
			watch := app.GetWatchWindow()
			watch.Visible = !watch.Visible
		case "debug_toggle_engine":
			newEng := app.GetDebugger().ToggleEngine()
			label := "Internal (F77)"
			if newEng == debugger.BackendGdbLldb {
				label = "Native (LLDB/GDB)"
			}
			app.GetMenuBar().SetDebugEngineLabel(label)
			app.SetStatusMessage(fmt.Sprintf("Debugger Engine set to: %s", label))
			sound.PlayBell()
		case "debug_toggle_bp":
			currLine := editor.CursorY + 1
			app.ToggleBreakpoint(currLine)
			sound.PlayBell()
		case "options_toggle_linenums":
			editor.ToggleLineNumbers()
		case "options_toggle_guides":
			en := editor.ToggleColumnGuides()
			app.GetMenuBar().SetColumnGuidesEnabled(en)
		case "options_toggle_sound":
			en := sound.Toggle()
			app.GetMenuBar().SetSoundEnabled(en)
		case "options_compiler_info":
			info, found := compiler.FindFortranCompiler()
			msg := "Open Watcom 2.0 FORTRAN 77 not found."
			if found {
				msg = fmt.Sprintf("Found compiler: %s (%s)", info.Path, info.Kind)
			}
			app.SetStatusMessage(msg)
		case "search_find":
			initQ := editor.GetWordUnderCursor()
			if initQ == "" {
				initQ = editor.LastFindQuery
			}
			findDlg.ShowWithTitle("Find", initQ, func(query string, caseSensitive bool) {
				found := editor.FindNext(query, caseSensitive)
				if found {
					sound.PlayBell()
					app.SetStatusMessage(fmt.Sprintf("Found %q", query))
				} else {
					sound.PlayError()
					app.SetStatusMessage(fmt.Sprintf("Search string not found: %q", query))
				}
			})
		case "search_project":
			initQ := editor.GetWordUnderCursor()
			if initQ == "" {
				initQ = editor.LastFindQuery
			}
			findDlg.ShowWithTitle("Find in Project", initQ, func(query string, caseSensitive bool) {
				matches := compiler.SearchInProject(editor.FilePath, query, caseSensitive)
				if len(matches) > 0 {
					sound.PlayBell()
					rootDir := compiler.GetSearchRootDir(editor.FilePath)
					searchResDlg.Show(matches, rootDir, func(match compiler.SearchMatch) {
						editor.PushNavLocation()
						if match.File != "" && match.File != editor.FilePath {
							if err := editor.LoadFile(match.File); err != nil {
								sound.PlayError()
								app.SetStatusMessage("Failed to open " + filepath.Base(match.File) + ": " + err.Error())
								return
							}
						}
						editor.GotoLine(match.Line, match.Column)
						app.SetStatusMessage(fmt.Sprintf("Jumped to %s:%d", filepath.Base(match.File), match.Line))
					})
				} else {
					sound.PlayError()
					app.SetStatusMessage(fmt.Sprintf("No matches found for %q in project", query))
				}
			})
		case "search_definition":
			sym := editor.GetWordUnderCursor()
			if sym == "" {
				sym = editor.LastFindQuery
			}
			if sym != "" {
				file, line, col, ok := compiler.FindDefinitionInProject(editor.FilePath, sym)
				if ok {
					sound.PlayBell()
					editor.PushNavLocation()
					if file != "" && file != editor.FilePath {
						if err := editor.LoadFile(file); err != nil {
							sound.PlayError()
							app.SetStatusMessage("Failed to open " + filepath.Base(file) + ": " + err.Error())
							return
						}
					}
					editor.GotoLine(line, col)
					app.SetStatusMessage(fmt.Sprintf("Jumped to definition of %q (%s:%d)", sym, filepath.Base(file), line))
				} else {
					sound.PlayError()
					app.SetStatusMessage(fmt.Sprintf("Definition not found for %q", sym))
				}
			} else {
				app.SetStatusMessage("No symbol under cursor (press F12 on function/subroutine name)")
			}
		case "search_prev_pos":
			if editor.NavigateBack() {
				sound.PlayBell()
				app.SetStatusMessage(fmt.Sprintf("Navigated back to %s:%d", editor.FileName, editor.CursorY+1))
			} else {
				app.SetStatusMessage("Navigation history: at oldest location")
			}
		case "search_next_pos":
			if editor.NavigateForward() {
				sound.PlayBell()
				app.SetStatusMessage(fmt.Sprintf("Navigated forward to %s:%d", editor.FileName, editor.CursorY+1))
			} else {
				app.SetStatusMessage("Navigation history: at newest location")
			}
		case "edit_undo":
			if editor.Undo() {
				sound.PlayBell()
				app.SetStatusMessage("Undo performed")
			} else {
				app.SetStatusMessage("Already at oldest change")
			}
		case "edit_redo":
			if editor.Redo() {
				sound.PlayBell()
				app.SetStatusMessage("Redo performed")
			} else {
				app.SetStatusMessage("Already at newest change")
			}
		case "search_again":
			if editor.LastFindQuery != "" {
				found := editor.FindNext(editor.LastFindQuery, editor.LastCaseSensitive)
				if found {
					sound.PlayBell()
				} else {
					sound.PlayError()
				}
			} else {
				dispatchAction("search_find")
			}
		case "search_goto":
			gotoDlg.Show(editor.CursorY+1, len(editor.Lines), func(targetLine int) {
				editor.GotoLine(targetLine, 1)
			})
		case "edit_copy":
			txt := editor.GetSelectedText()
			if editor.CopySelection() {
				sound.PlayBell()
				app.SetStatusMessage(fmt.Sprintf("Copied %d characters to clipboard", len([]rune(txt))))
			} else {
				app.SetStatusMessage("No text selected to copy (use Shift+Arrows)")
			}
		case "edit_cut":
			txt := editor.GetSelectedText()
			if editor.CutSelection() {
				sound.PlayBell()
				app.SetStatusMessage(fmt.Sprintf("Cut %d characters to clipboard", len([]rune(txt))))
			} else {
				app.SetStatusMessage("No text selected to cut (use Shift+Arrows)")
			}
		case "edit_paste":
			clip := ui.GetClipboard()
			if clip != "" {
				editor.PasteText(clip)
				sound.PlayBell()
				app.SetStatusMessage(fmt.Sprintf("Pasted %d characters from clipboard", len([]rune(clip))))
			} else {
				app.SetStatusMessage("Clipboard is empty")
			}
		case "edit_clear":
			if editor.DeleteSelection() {
				app.SetStatusMessage("Selection deleted")
			} else {
				app.SetStatusMessage("No text selected to clear")
			}
		case "edit_select_all":
			editor.SelectAll()
			app.SetStatusMessage("All text selected")
		case "help_about", "help_f77":
			aboutDlg.Show()
		}
	}

	app.SetActionHandler(dispatchAction)

	var lastEscTime time.Time

	// Main event loop
	for {
		app.Redraw()

		ev := screen.PollEvent()
		switch tev := ev.(type) {
		case *tcell.EventResize:
			screen.Sync()

		case *tcell.EventKey:
			key := tev.Key()
			mod := tev.Modifiers()
			ch := tev.Rune()

			// 1. User Screen handles any key to return to IDE
			if userScreen.Active {
				if key == tcell.KeyUp {
					userScreen.ScrollUp()
				} else if key == tcell.KeyDown {
					_, h := screen.Size()
					userScreen.ScrollDown(h)
				} else {
					userScreen.Hide()
				}
				continue
			}

			// 2. Modals handling
			if compileDlg.Visible {
				if key == tcell.KeyEnter {
					compileDlg.Hide()
					if compileDlg.Result != nil && len(compileDlg.Result.Errors) > 0 {
						errListDlg.Show(compileDlg.Result.Errors, func(errItem compiler.CompileError) {
							if errItem.File != "" && errItem.File != editor.FilePath {
								if err := editor.LoadFile(errItem.File); err != nil {
									sound.PlayError()
									app.SetStatusMessage("Failed to open " + filepath.Base(errItem.File) + ": " + err.Error())
									return
								}
							}
							editor.GotoLine(errItem.Line, errItem.Column)
						})
					}
				} else if key == tcell.KeyEscape || key == tcell.KeyRune {
					compileDlg.Hide()
				}
				continue
			}

			if errListDlg.Visible {
				switch key {
				case tcell.KeyUp:
					errListDlg.MoveUp()
				case tcell.KeyDown:
					errListDlg.MoveDown()
				case tcell.KeyEnter:
					errListDlg.SelectCurrent()
				case tcell.KeyEscape:
					errListDlg.Hide()
				}
				continue
			}

			if openDlg.Visible {
				switch key {
				case tcell.KeyUp:
					openDlg.MoveUp()
				case tcell.KeyDown:
					openDlg.MoveDown()
				case tcell.KeyEnter:
					openDlg.HandleEnter()
				case tcell.KeyEscape:
					openDlg.Hide()
				}
				continue
			}

			if saveDlg.Visible {
				switch key {
				case tcell.KeyEnter:
					saveDlg.Confirm()
				case tcell.KeyEscape:
					saveDlg.Hide()
				case tcell.KeyBackspace, tcell.KeyBackspace2:
					saveDlg.Backspace()
				case tcell.KeyRune:
					saveDlg.InsertRune(ch)
				}
				continue
			}

			if aboutDlg.Visible {
				if key == tcell.KeyEnter || key == tcell.KeyEscape || key == tcell.KeyRune {
					aboutDlg.Hide()
				}
				continue
			}

			if gotoDlg.Visible {
				switch key {
				case tcell.KeyEnter:
					gotoDlg.Confirm()
				case tcell.KeyEscape:
					gotoDlg.Hide()
				case tcell.KeyBackspace, tcell.KeyBackspace2:
					gotoDlg.Backspace()
				case tcell.KeyRune:
					gotoDlg.InsertRune(ch)
				}
				continue
			}

			if findDlg.Visible {
				switch key {
				case tcell.KeyEnter:
					findDlg.Confirm()
				case tcell.KeyEscape:
					findDlg.Hide()
				case tcell.KeyTab:
					findDlg.NextField()
				case tcell.KeyBacktab:
					findDlg.PrevField()
				case tcell.KeyBackspace, tcell.KeyBackspace2:
					findDlg.Backspace()
				case tcell.KeyRune:
					findDlg.InsertRune(ch)
				}
				continue
			}

			if searchResDlg.Visible {
				switch key {
				case tcell.KeyUp:
					searchResDlg.MoveUp()
				case tcell.KeyDown:
					searchResDlg.MoveDown()
				case tcell.KeyEnter:
					searchResDlg.SelectCurrent()
				case tcell.KeyEscape:
					searchResDlg.Hide()
				}
				continue
			}

			// 3. Global Shortcuts (Turbo C / Turbo Pascal Standard + macOS Option Key Workarounds)
			now := time.Now()
			isEscPrefix := (!lastEscTime.IsZero() && now.Sub(lastEscTime) < 400*time.Millisecond)

			if key == tcell.KeyEscape {
				if app.IsMenuActive() {
					app.MenuClose()
					lastEscTime = time.Time{}
					continue
				}
				if editor.SelectActive {
					editor.ClearSelection()
					lastEscTime = time.Time{}
					continue
				}
				lastEscTime = now
				continue
			}
			lastEscTime = time.Time{}

			isAlt := ((mod & (tcell.ModAlt | tcell.ModMeta)) != 0) || isEscPrefix
			menuIdx := matchMenuHotKey(ch)

			if isAlt {
				if key == tcell.KeyF9 {
					// Alt+F9: Compile
					dispatchAction("compile_compile")
					continue
				} else if key == tcell.KeyF5 {
					// Alt+F5: User Screen
					dispatchAction("run_userscreen")
					continue
				} else if key == tcell.KeyF3 {
					// Alt+F3: Find in Project
					dispatchAction("search_project")
					continue
				} else if key == tcell.KeyLeft {
					// Alt+Left: Previous Location
					dispatchAction("search_prev_pos")
					continue
				} else if key == tcell.KeyRight {
					// Alt+Right: Next Location
					dispatchAction("search_next_pos")
					continue
				} else if key == tcell.KeyBackspace || key == tcell.KeyBackspace2 {
					// Alt+Backspace: Undo (Classic Turbo Vision convention)
					dispatchAction("edit_undo")
					continue
				} else if menuIdx >= 0 {
					// Alt+F, Alt+E, Alt+S, Alt+R, Alt+C, Alt+D, Alt+O, Alt+W, Alt+H: Open corresponding menu directly!
					app.OpenMenuAt(menuIdx)
					continue
				} else if ch == 'x' || ch == 'X' {
					// Alt+X: Exit
					dispatchAction("app_exit")
					return
				} else if ch == 'l' || ch == 'L' {
					// Alt+L: Toggle Line Numbers
					dispatchAction("options_toggle_linenums")
					continue
				} else if ch == 'g' || ch == 'G' {
					// Alt+G: Go to Line
					dispatchAction("search_goto")
					continue
				} else if ch == 'q' || ch == 'Q' {
					// Alt+Q: Stop Debugger
					dispatchAction("debug_stop")
					continue
				}
			}

			if mod&tcell.ModCtrl != 0 {
				if key == tcell.KeyCtrlZ {
					// Ctrl+Z: Undo
					dispatchAction("edit_undo")
					continue
				} else if key == tcell.KeyCtrlY {
					// Ctrl+Y: Redo
					dispatchAction("edit_redo")
					continue
				} else if key == tcell.KeyCtrlUnderscore || (ch == '-' && mod&tcell.ModCtrl != 0) {
					if mod&tcell.ModShift != 0 {
						// Ctrl+Shift+-: Next Location
						dispatchAction("search_next_pos")
					} else {
						// Ctrl+-: Previous Location
						dispatchAction("search_prev_pos")
					}
					continue
				} else if key == tcell.KeyCtrlC {
					dispatchAction("edit_copy")
					continue
				} else if key == tcell.KeyCtrlX {
					dispatchAction("edit_cut")
					continue
				} else if key == tcell.KeyCtrlV {
					dispatchAction("edit_paste")
					continue
				} else if key == tcell.KeyCtrlA {
					dispatchAction("edit_select_all")
					continue
				} else if key == tcell.KeyF9 {
					dispatchAction("run_run")
					continue
				} else if key == tcell.KeyF2 {
					dispatchAction("debug_stop")
					continue
				} else if key == tcell.KeyCtrlF {
					dispatchAction("search_find")
					continue
				} else if key == tcell.KeyCtrlL {
					dispatchAction("search_again")
					continue
				} else if key == tcell.KeyCtrlG {
					dispatchAction("search_goto")
					continue
				} else if key == tcell.KeyInsert {
					dispatchAction("edit_copy")
					continue
				} else if key == tcell.KeyDelete {
					dispatchAction("edit_clear")
					continue
				}
			}

			// Function Keys (F1..F12)
			switch key {
			case tcell.KeyF1:
				dispatchAction("help_about")
				continue
			case tcell.KeyF2:
				dispatchAction("file_save")
				continue
			case tcell.KeyF3:
				dispatchAction("file_open")
				continue
			case tcell.KeyF4:
				dispatchAction("debug_toggle_bp")
				continue
			case tcell.KeyF5:
				dispatchAction("debug_continue")
				continue
			case tcell.KeyF6:
				dispatchAction("options_toggle_linenums")
				continue
			case tcell.KeyF7:
				dispatchAction("debug_step_into")
				continue
			case tcell.KeyF8:
				dispatchAction("debug_step_over")
				continue
			case tcell.KeyF9:
				dispatchAction("compile_make")
				continue
			case tcell.KeyF10:
				app.ToggleMenu()
				continue
			case tcell.KeyF12:
				if mod&tcell.ModShift != 0 {
					// Shift+F12: Previous Location
					dispatchAction("search_prev_pos")
				} else {
					dispatchAction("search_definition")
				}
				continue
			}

			// 4. Menu Navigation
			if app.IsMenuActive() {
				if ch != 0 {
					// Direct hotkey to switch top-level menu (F, E, S, R, C, D, O, W, H)
					if idx := matchMenuHotKey(ch); idx >= 0 {
						app.OpenMenuAt(idx)
						continue
					}
					// Hotkey for items within current dropdown (e.g. N for New, O for Open, X for Exit)
					act := app.MenuTriggerHotKey(ch)
					if act != "" {
						app.MenuClose()
						dispatchAction(act)
						continue
					}
				}

				switch key {
				case tcell.KeyLeft:
					app.MenuMoveLeft()
				case tcell.KeyRight:
					app.MenuMoveRight()
				case tcell.KeyUp:
					app.MenuMoveUp()
				case tcell.KeyDown:
					app.MenuMoveDown()
				case tcell.KeyEnter:
					act := app.MenuSelect()
					if act != "" {
						dispatchAction(act)
					}
				case tcell.KeyEscape:
					app.MenuClose()
				}
				continue
			}

			// 5. Editor Navigation & Typing
			// Shift + Arrow block selection
			if mod&tcell.ModShift != 0 {
				switch key {
				case tcell.KeyLeft:
					editor.StartSelection()
					editor.MoveLeft()
					editor.UpdateSelection()
				case tcell.KeyRight:
					editor.StartSelection()
					editor.MoveRight()
					editor.UpdateSelection()
				case tcell.KeyUp:
					editor.StartSelection()
					editor.MoveUp()
					editor.UpdateSelection()
				case tcell.KeyDown:
					editor.StartSelection()
					editor.MoveDown()
					editor.UpdateSelection()
				}
				_, h := screen.Size()
				editor.AdjustScroll(app.GetEditor().CursorX+1, h-4)
				continue
			}

			switch key {
			case tcell.KeyLeft:
				if editor.SelectActive {
					editor.ClearSelection()
				}
				editor.MoveLeft()
			case tcell.KeyRight:
				if editor.SelectActive {
					editor.ClearSelection()
				}
				editor.MoveRight()
			case tcell.KeyUp:
				if editor.SelectActive {
					editor.ClearSelection()
				}
				editor.MoveUp()
			case tcell.KeyDown:
				if editor.SelectActive {
					editor.ClearSelection()
				}
				editor.MoveDown()
			case tcell.KeyHome:
				editor.MoveHome()
			case tcell.KeyEnd:
				editor.MoveEnd()
			case tcell.KeyPgUp:
				_, h := screen.Size()
				editor.PageUp(h - 4)
			case tcell.KeyPgDn:
				_, h := screen.Size()
				editor.PageDown(h - 4)
			case tcell.KeyEnter:
				editor.InsertNewLine()
			case tcell.KeyTab:
				editor.InsertTab()
			case tcell.KeyBackspace, tcell.KeyBackspace2:
				editor.Backspace()
			case tcell.KeyDelete:
				editor.Delete()
			case tcell.KeyEscape:
				editor.ClearSelection()
				editor.ClearHighlight()
			case tcell.KeyRune:
				editor.InsertRune(ch)
			}

			_, h := screen.Size()
			editor.AdjustScroll(app.GetEditor().CursorX+1, h-4)
		}
	}
}
