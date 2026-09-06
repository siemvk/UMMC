package gui

import (
	"UMMC/help"
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type SavesView struct {
	window       fyne.Window
	logger       *ActivityLogger
	container    *fyne.Container
	savesList    *widget.List
	saves        []help.SaveRecord
	filtered     []help.SaveRecord
	selectedID   int
	activeInfo   *help.ActiveSaveInfo
	filterSelect *widget.Select
	activeStatus *widget.Label

	// Details widgets
	splashImage     *canvas.Image
	splashContainer *fyne.Container
	nameLabel       *widget.Label
	modLabel        *widget.Label
	statusLabel     *widget.Label
	dateLabel       *widget.Label
	pathLabel       *widget.Label
	loadBtn         *widget.Button
	copyBtn         *widget.Button
	deleteBtn       *widget.Button
}

func NewSavesView(win fyne.Window, logger *ActivityLogger) *SavesView {
	sv := &SavesView{
		window:     win,
		logger:     logger,
		selectedID: -1,
	}

	sv.buildUI()
	sv.Refresh()
	return sv
}

func (sv *SavesView) Container() *fyne.Container {
	return sv.container
}

func (sv *SavesView) buildUI() {
	sv.savesList = widget.NewList(
		func() int {
			return len(sv.filtered)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.DocumentSaveIcon()),
				widget.NewLabel("Save Name (Mod)"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < 0 || id >= len(sv.filtered) {
				return
			}
			s := sv.filtered[id]
			box := obj.(*fyne.Container)
			icon := box.Objects[0].(*widget.Icon)
			label := box.Objects[1].(*widget.Label)

			isActive := sv.activeInfo != nil && sv.activeInfo.Name == s.Name && sv.activeInfo.ModName == s.ModName
			if isActive {
				icon.SetResource(theme.ConfirmIcon())
				label.SetText(fmt.Sprintf("%s  [Mod: %s]  ★ ACTIVE", s.Name, s.ModName))
			} else {
				icon.SetResource(theme.DocumentSaveIcon())
				label.SetText(fmt.Sprintf("%s  [Mod: %s]", s.Name, s.ModName))
			}
		},
	)

	sv.savesList.OnSelected = func(id widget.ListItemID) {
		if id >= 0 && id < len(sv.filtered) {
			sv.selectedID = sv.filtered[id].ID
			sv.showSaveDetails(sv.filtered[id])
		}
	}

	// Details Pane - Splash Image
	sv.splashImage = canvas.NewImageFromFile("")
	sv.splashImage.FillMode = canvas.ImageFillContain
	sv.splashImage.ScaleMode = canvas.ImageScaleSmooth
	sv.splashImage.SetMinSize(fyne.NewSize(120, 120))
	sv.splashContainer = container.NewPadded(sv.splashImage)
	sv.splashContainer.Hide()

	sv.nameLabel = widget.NewLabelWithStyle("-", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	sv.modLabel = widget.NewLabel("-")
	sv.statusLabel = widget.NewLabel("-")
	sv.dateLabel = widget.NewLabel("-")
	sv.pathLabel = widget.NewLabel("-")

	sv.loadBtn = widget.NewButtonWithIcon("Load This Save", theme.MediaPlayIcon(), func() {
		sv.loadSelectedSave()
	})
	sv.loadBtn.Importance = widget.HighImportance
	sv.loadBtn.Disable()

	sv.copyBtn = widget.NewButtonWithIcon("Copy / Change Mod", theme.ContentCopyIcon(), func() {
		sv.showCopySaveDialog()
	})
	sv.copyBtn.Disable()

	sv.deleteBtn = widget.NewButtonWithIcon("Delete Save", theme.DeleteIcon(), func() {
		sv.deleteSelectedSave()
	})
	sv.deleteBtn.Importance = widget.DangerImportance
	sv.deleteBtn.Disable()

	detailsForm := widget.NewForm(
		widget.NewFormItem("Save Name:", sv.nameLabel),
		widget.NewFormItem("Associated Mod:", sv.modLabel),
		widget.NewFormItem("Current Status:", sv.statusLabel),
		widget.NewFormItem("Last Saved:", sv.dateLabel),
		widget.NewFormItem("Save Path:", sv.pathLabel),
	)

	actionButtons := container.NewGridWithColumns(3,
		sv.loadBtn,
		sv.copyBtn,
		sv.deleteBtn,
	)

	bottomInfo := container.NewVBox(
		detailsForm,
		widget.NewSeparator(),
		actionButtons,
	)

	detailsCard := widget.NewCard("Save Profile Details", "", container.NewBorder(
		nil,
		bottomInfo,
		nil,
		nil,
		sv.splashContainer,
	))

	// Header & Actions - Single Row
	saveCurrentBtn := widget.NewButtonWithIcon("Save Current Game", theme.DocumentSaveIcon(), func() {
		sv.showSaveCurrentDialog()
	})
	saveCurrentBtn.Importance = widget.HighImportance

	refreshBtn := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		sv.Refresh()
	})

	sv.activeStatus = widget.NewLabel("Active Save: Detecting...")

	sv.filterSelect = widget.NewSelect([]string{"All Mods"}, func(selected string) {
		sv.applyFilter(selected)
	})
	sv.filterSelect.SetSelected("All Mods")

	topHeader := container.NewBorder(
		nil,
		nil,
		widget.NewLabelWithStyle("Undertale Saves", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewHBox(saveCurrentBtn, refreshBtn),
	)

	filterBar := container.NewBorder(
		nil,
		nil,
		widget.NewLabel("Filter by Mod:"),
		nil,
		sv.filterSelect,
	)

	leftPane := container.NewBorder(
		container.NewVBox(topHeader, sv.activeStatus, filterBar, widget.NewSeparator()),
		nil,
		nil,
		nil,
		sv.savesList,
	)

	split := container.NewHSplit(leftPane, detailsCard)
	split.SetOffset(0.55)

	sv.container = container.NewBorder(nil, nil, nil, nil, split)
}

func (sv *SavesView) getAvailableModOptions() []string {
	modSet := map[string]bool{
		"Vanilla": true,
	}

	if mods, err := help.GetMods(); err == nil {
		for _, m := range mods {
			modSet[m.Name] = true
		}
	}

	for _, s := range sv.saves {
		if s.ModName != "" {
			modSet[s.ModName] = true
		}
	}

	options := []string{"Vanilla"}
	for m := range modSet {
		if m != "Vanilla" {
			options = append(options, m)
		}
	}
	return options
}

func (sv *SavesView) Refresh() {
	activeInfo, _ := help.GetActiveSaveInfo()
	sv.activeInfo = activeInfo

	if activeInfo != nil && activeInfo.Name != "" {
		sv.activeStatus.SetText(fmt.Sprintf("Active Save in Undertale: '%s' (Mod: %s)", activeInfo.Name, activeInfo.ModName))
	} else {
		if help.HasActiveSaveFiles(help.GetUndertaleSaveDir()) {
			sv.activeStatus.SetText("Active Save in Undertale: Untracked save files present")
		} else {
			sv.activeStatus.SetText("Active Save in Undertale: No active save files")
		}
	}

	saves, err := help.GetSaves("")
	if err != nil {
		sv.logger.Log(fmt.Sprintf("Failed to load saves: %v", err))
		return
	}
	sv.saves = saves

	// Update filter options
	currentFilter := "All Mods"
	if sv.filterSelect != nil && sv.filterSelect.Selected != "" {
		currentFilter = sv.filterSelect.Selected
	}

	options := []string{"All Mods"}
	options = append(options, sv.getAvailableModOptions()...)
	if sv.filterSelect != nil {
		sv.filterSelect.Options = options
		sv.filterSelect.SetSelected(currentFilter)
	}

	sv.applyFilter(currentFilter)
}

func (sv *SavesView) applyFilter(modFilter string) {
	if modFilter == "All Mods" || modFilter == "" {
		sv.filtered = sv.saves
	} else {
		sv.filtered = nil
		for _, s := range sv.saves {
			if strings.EqualFold(s.ModName, modFilter) {
				sv.filtered = append(sv.filtered, s)
			}
		}
	}

	sv.savesList.Refresh()

	if len(sv.filtered) == 0 {
		sv.selectedID = -1
		sv.splashContainer.Hide()
		sv.nameLabel.SetText("No saves found")
		sv.modLabel.SetText("-")
		sv.statusLabel.SetText("-")
		sv.dateLabel.SetText("-")
		sv.pathLabel.SetText("-")
		sv.loadBtn.Disable()
		sv.copyBtn.Disable()
		sv.deleteBtn.Disable()
	} else if sv.selectedID != -1 {
		for i, s := range sv.filtered {
			if s.ID == sv.selectedID {
				sv.savesList.Select(i)
				return
			}
		}
		sv.savesList.Select(0)
	} else {
		sv.savesList.Select(0)
	}
}

func (sv *SavesView) showSaveDetails(s help.SaveRecord) {
	sv.nameLabel.SetText(s.Name)
	sv.modLabel.SetText(s.ModName)

	isActive := sv.activeInfo != nil && sv.activeInfo.Name == s.Name && sv.activeInfo.ModName == s.ModName
	if isActive {
		sv.statusLabel.SetText("★ Currently Active in Undertale")
	} else {
		sv.statusLabel.SetText("Inactive")
	}

	sv.dateLabel.SetText(s.CreatedAt.Local().Format("2006-01-02 15:04:05"))
	sv.pathLabel.SetText(s.Path)

	// Display splash image for the mod this save belongs to
	splashPath := help.GetSaveSplashImage(s.Name, s.ModName)
	if splashPath != "" {
		sv.splashImage.File = splashPath
		sv.splashImage.Refresh()
		sv.splashContainer.Show()
	} else {
		sv.splashContainer.Hide()
	}

	sv.loadBtn.Enable()
	sv.copyBtn.Enable()
	sv.deleteBtn.Enable()
}

func (sv *SavesView) getSelectedSave() *help.SaveRecord {
	for _, s := range sv.filtered {
		if s.ID == sv.selectedID {
			return &s
		}
	}
	return nil
}

func (sv *SavesView) loadSelectedSave() {
	s := sv.getSelectedSave()
	if s == nil {
		return
	}

	activeDir := help.GetUndertaleSaveDir()
	hasFiles := help.HasActiveSaveFiles(activeDir)
	activeInfo, _ := help.GetActiveSaveInfo()

	doLoad := func() {
		sv.logger.Log(fmt.Sprintf("Loading save '%s' (mod: %s)...", s.Name, s.ModName))
		go func() {
			if err := help.LoadSave(s.Name, s.ModName, sv.logger.Log); err != nil {
				dialog.ShowError(err, sv.window)
				return
			}
			sv.Refresh()
			dialog.ShowInformation("Save Loaded", fmt.Sprintf("Save '%s' (mod: %s) is now loaded and active in Undertale!", s.Name, s.ModName), sv.window)
		}()
	}

	if hasFiles && (activeInfo == nil || activeInfo.Name == "") {
		// Untracked active save! Prompt user to save it before loading
		nameEntry := widget.NewEntry()
		nameEntry.SetText("Previous Run")

		modOptions := sv.getAvailableModOptions()
		modSelect := widget.NewSelect(modOptions, nil)
		modSelect.SetSelected("Vanilla")

		form := widget.NewForm(
			widget.NewFormItem("Save Profile Name:", nameEntry),
			widget.NewFormItem("Associated Mod:", modSelect),
		)

		d := dialog.NewCustomConfirm(
			"Save Active Game Progress",
			"Save & Load Target",
			"Cancel",
			container.NewVBox(
				widget.NewLabel("Active save files were detected in Undertale without a save profile name.\nPlease name your current save to preserve your progress before loading:"),
				form,
			),
			func(confirm bool) {
				if !confirm {
					return
				}
				saveName := strings.TrimSpace(nameEntry.Text)
				if saveName == "" {
					saveName = "untracked_save"
				}
				modName := modSelect.Selected
				if modName == "" {
					modName = "Vanilla"
				}
				go func() {
					if _, err := help.SaveCurrentGame(saveName, modName, true, sv.logger.Log); err != nil {
						dialog.ShowError(err, sv.window)
						return
					}
					doLoad()
				}()
			},
			sv.window,
		)
		d.Resize(fyne.NewSize(500, 240))
		d.Show()
		return
	}

	dialog.ShowConfirm(
		"Load Save",
		fmt.Sprintf("Load save '%s' (mod: %s)?\n\nYour current game progress will be automatically auto-saved before loading.", s.Name, s.ModName),
		func(confirm bool) {
			if !confirm {
				return
			}
			doLoad()
		},
		sv.window,
	)
}

func (sv *SavesView) showSaveCurrentDialog() {
	nameEntry := widget.NewEntry()
	defaultMod := "Vanilla"
	if sv.activeInfo != nil && sv.activeInfo.Name != "" {
		nameEntry.SetText(sv.activeInfo.Name)
		if sv.activeInfo.ModName != "" {
			defaultMod = sv.activeInfo.ModName
		}
	} else {
		nameEntry.SetPlaceHolder("e.g. Pacifist Run, Before Sans...")
	}

	modOptions := sv.getAvailableModOptions()
	modSelect := widget.NewSelect(modOptions, nil)
	modSelect.SetSelected(defaultMod)

	form := widget.NewForm(
		widget.NewFormItem("Save Name:", nameEntry),
		widget.NewFormItem("Associated Mod:", modSelect),
	)

	d := dialog.NewCustomConfirm(
		"Save Current Game",
		"Save Progress",
		"Cancel",
		container.NewVBox(
			widget.NewLabel("Create or update a save slot from your active Undertale game files:"),
			form,
		),
		func(confirm bool) {
			if !confirm {
				return
			}
			saveName := strings.TrimSpace(nameEntry.Text)
			if saveName == "" {
				dialog.ShowError(fmt.Errorf("save name cannot be empty"), sv.window)
				return
			}
			modName := modSelect.Selected
			if modName == "" {
				modName = "Vanilla"
			}

			go func() {
				rec, err := help.SaveCurrentGame(saveName, modName, true, sv.logger.Log)
				if err != nil {
					dialog.ShowError(err, sv.window)
					return
				}
				sv.selectedID = rec.ID
				sv.Refresh()
				dialog.ShowInformation("Game Saved", fmt.Sprintf("Current game successfully saved as '%s' (mod: %s)!", saveName, modName), sv.window)
			}()
		},
		sv.window,
	)

	d.Resize(fyne.NewSize(480, 220))
	d.Show()
}

func (sv *SavesView) showCopySaveDialog() {
	s := sv.getSelectedSave()
	if s == nil {
		return
	}

	nameEntry := widget.NewEntry()
	nameEntry.SetText(s.Name + "_copy")

	modOptions := sv.getAvailableModOptions()
	modSelect := widget.NewSelect(modOptions, nil)
	modSelect.SetSelected(s.ModName)

	form := widget.NewForm(
		widget.NewFormItem("Original Save:", widget.NewLabel(fmt.Sprintf("%s (%s)", s.Name, s.ModName))),
		widget.NewFormItem("New Save Name:", nameEntry),
		widget.NewFormItem("Target Mod:", modSelect),
	)

	d := dialog.NewCustomConfirm(
		"Duplicate / Transfer Save",
		"Make Copy",
		"Cancel",
		container.NewVBox(
			widget.NewLabel("Create a duplicate of this save and optionally change its associated mod:"),
			form,
		),
		func(confirm bool) {
			if !confirm {
				return
			}
			newName := strings.TrimSpace(nameEntry.Text)
			if newName == "" {
				dialog.ShowError(fmt.Errorf("new save name cannot be empty"), sv.window)
				return
			}
			newMod := modSelect.Selected
			if newMod == "" {
				newMod = s.ModName
			}

			go func() {
				rec, err := help.CopySaveAndChangeMod(s.Name, s.ModName, newName, newMod, sv.logger.Log)
				if err != nil {
					dialog.ShowError(err, sv.window)
					return
				}
				sv.selectedID = rec.ID
				sv.Refresh()
				dialog.ShowInformation("Save Copied", fmt.Sprintf("Save successfully copied to '%s' (mod: %s)!", rec.Name, rec.ModName), sv.window)
			}()
		},
		sv.window,
	)

	d.Resize(fyne.NewSize(500, 240))
	d.Show()
}

func (sv *SavesView) deleteSelectedSave() {
	s := sv.getSelectedSave()
	if s == nil {
		return
	}

	dialog.ShowConfirm(
		"Delete Save Slot",
		fmt.Sprintf("Are you sure you want to delete save '%s' (mod: %s)?\nThis will remove the save folder from disk.", s.Name, s.ModName),
		func(confirm bool) {
			if !confirm {
				return
			}
			go func() {
				if err := help.DeleteSave(s.Name, s.ModName, sv.logger.Log); err != nil {
					dialog.ShowError(err, sv.window)
					return
				}
				sv.selectedID = -1
				sv.Refresh()
			}()
		},
		sv.window,
	)
}
