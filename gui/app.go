package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func init() {
	enJson := []byte(`{
		"Advanced": "Advanced",
		"Cancel": "Cancel",
		"Confirm": "Confirm",
		"Copy": "Copy",
		"Create Folder": "Create Folder",
		"Cut": "Cut",
		"Enter filename": "Enter filename",
		"Error": "Error",
		"Favourites": "Favourites",
		"File": "File",
		"Folder": "Folder",
		"New Folder": "New Folder",
		"No": "No",
		"OK": "OK",
		"Open": "Open",
		"Paste": "Paste",
		"Quit": "Quit",
		"Redo": "Redo",
		"Save": "Save",
		"Select all": "Select all",
		"Show Hidden Files": "Show Hidden Files",
		"Undo": "Undo",
		"Yes": "Yes"
	}`)
	_ = lang.AddTranslationsForLocale(enJson, fyne.Locale("nl"))
	_ = lang.AddTranslationsForLocale(enJson, fyne.Locale("el"))
}

// StartApp initializes and runs the Fyne graphical interface.
func StartApp() {
	a := app.NewWithID("com.siemvk.UMMC")
	w := a.NewWindow("UMMC - Undertale ModManager Macos")
	w.Resize(fyne.NewSize(880, 600))

	// Status label
	statusLabel := widget.NewLabel("Ready")
	logger := NewActivityLogger(statusLabel)
	logger.Log("UMMC GUI initialized.")

	// Views
	modsView := NewModsView(w, logger)
	savesView := NewSavesView(w, logger)
	backupsView := NewBackupsView(w, logger)

	var toolsView *ToolsView
	refreshAll := func() {
		modsView.Refresh()
		savesView.Refresh()
		backupsView.Refresh()
		if toolsView != nil {
			toolsView.Refresh()
		}
	}

	toolsView = NewToolsView(w, a, logger, refreshAll)

	// Log tab content
	clearLogBtn := widget.NewButtonWithIcon("Clear Logs", theme.DeleteIcon(), func() {
		logger.Clear()
	})
	logTabContent := container.NewBorder(
		container.NewHBox(widget.NewLabelWithStyle("Activity Log", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), clearLogBtn),
		nil,
		nil,
		nil,
		logger.Widget(),
	)

	// Main App Tabs
	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("Mods", theme.FolderIcon(), modsView.Container()),
		container.NewTabItemWithIcon("Saves", theme.DocumentSaveIcon(), savesView.Container()),
		container.NewTabItemWithIcon("Backups", theme.StorageIcon(), backupsView.Container()),
		container.NewTabItemWithIcon("Tools & Windows", theme.SettingsIcon(), toolsView.Container()),
		container.NewTabItemWithIcon("Logs", theme.DocumentIcon(), logTabContent),
	)

	tabs.OnSelected = func(tab *container.TabItem) {
		switch tab.Text {
		case "Mods":
			modsView.Refresh()
		case "Saves":
			savesView.Refresh()
		case "Backups":
			backupsView.Refresh()
		case "Tools & Windows":
			toolsView.Refresh()
		}
	}

	// Bottom bar with status text
	statusBar := container.NewBorder(
		nil,
		nil,
		widget.NewLabelWithStyle("Status:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		nil,
		statusLabel,
	)

	mainLayout := container.NewBorder(
		nil,
		container.NewVBox(widget.NewSeparator(), statusBar),
		nil,
		nil,
		tabs,
	)

	w.SetContent(mainLayout)

	// Show onboarding on first launch
	if !a.Preferences().BoolWithFallback("onboarding_completed", false) {
		ShowOnboardingWizard(w, a, logger, func() {
			refreshAll()
			toolsView.Refresh()
		})
	}

	w.ShowAndRun()
}
