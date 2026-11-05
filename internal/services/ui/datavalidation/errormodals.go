// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

// errormodals.go provides error modals

package datavalidation

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func showValidationErrorModal(app *tview.Application, container *tview.Flex, returnTo tview.Primitive, message string) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			app.SetRoot(container, true).SetFocus(returnTo)
		})

	modal.SetBackgroundColor(tcell.ColorDarkRed).
		SetBorderColor(tcell.ColorRed)
	modal.SetButtonBackgroundColor(tcell.ColorDarkRed).
		SetButtonTextColor(tcell.ColorWhite)

	app.SetRoot(modal, true).SetFocus(modal)
}
