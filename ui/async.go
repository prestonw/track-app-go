package ui

import "fyne.io/fyne/v2"

// onMain runs UI work on Fyne's main goroutine (required from v2.7+ for background updates).
func onMain(fn func()) {
	if fn != nil {
		fyne.Do(fn)
	}
}