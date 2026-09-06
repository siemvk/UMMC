package gui

import (
	"UMMC/help"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const VanillaModID = -1

var VanillaMod = help.ModRecord{
	ID:    VanillaModID,
	Name:  "Undertale (Original)",
	Maker: "Toby Fox",
	Base:  "Vanilla",
}

type ModsView struct {
	window     fyne.Window
	logger     *ActivityLogger
	container  *fyne.Container
	modList    *widget.List
	mods       []help.ModRecord
	selectedID int

	// Details widgets
	splashImage     *canvas.Image
	splashContainer *fyne.Container
	nameLabel       *widget.Label
	makerLabel      *widget.Label
	baseLabel       *widget.Label
	applyBtn        *widget.Button
	editBtn         *widget.Button
	deleteBtn       *widget.Button
}

func NewModsView(win fyne.Window, logger *ActivityLogger) *ModsView {
	mv := &ModsView{
		window:     win,
		logger:     logger,
		selectedID: VanillaModID,
	}

	mv.buildUI()
	mv.Refresh()
	return mv
}

func (mv *ModsView) Container() *fyne.Container {
	return mv.container
}

func (mv *ModsView) buildUI() {
	// Mod List Widget
	mv.modList = widget.NewList(
		func() int {
			return len(mv.mods)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.FolderIcon()),
				widget.NewLabel("Mod Name (by Author)"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < 0 || id >= len(mv.mods) {
				return
			}
			m := mv.mods[id]
			box := obj.(*fyne.Container)
			icon := box.Objects[0].(*widget.Icon)
			label := box.Objects[1].(*widget.Label)
			if m.ID == VanillaModID {
				icon.SetResource(theme.HomeIcon())
				label.SetText(fmt.Sprintf("%s  [%s | %s]", m.Name, m.Maker, m.Base))
			} else {
				icon.SetResource(theme.FolderIcon())
				label.SetText(fmt.Sprintf("%s  [Author: %s | Base: %s]", m.Name, m.Maker, m.Base))
			}
		},
	)

	mv.modList.OnSelected = func(id widget.ListItemID) {
		if id >= 0 && id < len(mv.mods) {
			mv.selectedID = mv.mods[id].ID
			mv.showModDetails(mv.mods[id])
		}
	}

	// Details Pane
	mv.splashImage = canvas.NewImageFromFile("")
	mv.splashImage.FillMode = canvas.ImageFillContain
	mv.splashImage.ScaleMode = canvas.ImageScaleSmooth
	mv.splashImage.SetMinSize(fyne.NewSize(120, 120))
	mv.splashContainer = container.NewPadded(mv.splashImage)
	mv.splashContainer.Hide()

	mv.nameLabel = widget.NewLabelWithStyle("-", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	mv.makerLabel = widget.NewLabel("-")
	mv.baseLabel = widget.NewLabel("-")

	mv.applyBtn = widget.NewButtonWithIcon("Apply Mod", theme.ConfirmIcon(), func() {
		mv.applySelectedMod()
	})
	mv.applyBtn.Importance = widget.HighImportance
	mv.applyBtn.Disable()

	mv.editBtn = widget.NewButtonWithIcon("Edit Metadata", theme.DocumentCreateIcon(), func() {
		mv.showEditModDialog()
	})
	mv.editBtn.Disable()

	mv.deleteBtn = widget.NewButtonWithIcon("Delete Mod", theme.DeleteIcon(), func() {
		mv.deleteSelectedMod()
	})
	mv.deleteBtn.Importance = widget.DangerImportance
	mv.deleteBtn.Disable()

	detailsForm := widget.NewForm(
		widget.NewFormItem("Mod Name:", mv.nameLabel),
		widget.NewFormItem("Author / Maker:", mv.makerLabel),
		widget.NewFormItem("Target Base:", mv.baseLabel),
	)

	// Single row for details action buttons
	actionButtons := container.NewGridWithColumns(3,
		mv.applyBtn,
		mv.editBtn,
		mv.deleteBtn,
	)

	bottomInfo := container.NewVBox(
		detailsForm,
		widget.NewSeparator(),
		actionButtons,
	)

	detailsCard := widget.NewCard("Mod Details", "", container.NewBorder(
		nil,
		bottomInfo,
		nil,
		nil,
		mv.splashContainer,
	))

	// Left Side List Header & Actions - Single Row
	addBtn := widget.NewButtonWithIcon("Add", theme.ContentAddIcon(), func() {
		mv.showAddModDialog()
	})
	quickPatchBtn := widget.NewButtonWithIcon("Quick Patch", theme.DocumentCreateIcon(), func() {
		mv.showQuickPatchDialog()
	})
	refreshBtn := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		mv.Refresh()
	})

	topHeader := container.NewBorder(
		nil,
		nil,
		widget.NewLabelWithStyle("Installed Mods", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewHBox(addBtn, quickPatchBtn, refreshBtn),
	)

	leftPane := container.NewBorder(
		container.NewVBox(topHeader, widget.NewSeparator()),
		nil,
		nil,
		nil,
		mv.modList,
	)

	// Split Container
	split := container.NewHSplit(leftPane, detailsCard)
	split.SetOffset(0.55)

	mv.container = container.NewBorder(nil, nil, nil, nil, split)
}

func (mv *ModsView) Refresh() {
	mods, err := help.GetMods()
	if err != nil {
		mv.logger.Log(fmt.Sprintf("Failed to load mods: %v", err))
		mods = nil
	}

	// Always include Vanilla Undertale at index 0
	mv.mods = append([]help.ModRecord{VanillaMod}, mods...)
	mv.modList.Refresh()

	// Maintain selection or default to Vanilla
	found := false
	for i, m := range mv.mods {
		if m.ID == mv.selectedID {
			mv.modList.Select(i)
			found = true
			break
		}
	}
	if !found {
		mv.modList.Select(0)
	}
}

func (mv *ModsView) showModDetails(m help.ModRecord) {
	mv.nameLabel.SetText(m.Name)
	mv.makerLabel.SetText(m.Maker)
	mv.baseLabel.SetText(m.Base)

	var splashPath string

	if m.ID == VanillaModID {
		splashPath = help.GetDefaultSplashImage()
		mv.applyBtn.SetText("Restore Vanilla")
		mv.applyBtn.SetIcon(theme.ConfirmIcon())
		mv.applyBtn.Enable()
		mv.editBtn.Disable()
		mv.deleteBtn.Disable()
	} else {
		splashPath = help.GetModSplashImage(m.Name)
		if m.UseButterscotch {
			mv.applyBtn.SetText("Load & Play")
			mv.applyBtn.SetIcon(theme.MediaPlayIcon())
		} else {
			mv.applyBtn.SetText("Apply Mod")
			mv.applyBtn.SetIcon(theme.ConfirmIcon())
		}
		mv.applyBtn.Enable()
		mv.editBtn.Enable()
		mv.deleteBtn.Enable()

		var tags []string
		if !strings.HasSuffix(m.Base, "-w") {
			tags = append(tags, "macOS Mod")
		}
		if m.InstallToAppRoot {
			tags = append(tags, "Root Mode")
		}
		if m.UseButterscotch {
			tags = append(tags, "Butterscotch")
		}
		if m.UseVanillaSaves {
			tags = append(tags, "Vanilla Saves")
		}
		if len(tags) > 0 {
			mv.baseLabel.SetText(fmt.Sprintf("%s  [%s]", m.Base, strings.Join(tags, " | ")))
		}
	}

	if splashPath != "" {
		mv.splashImage.File = splashPath
		mv.splashImage.Show()
		mv.splashContainer.Show()
		mv.splashImage.Refresh()
	} else {
		mv.splashImage.File = ""
		mv.splashImage.Hide()
		mv.splashContainer.Hide()
		mv.splashImage.Refresh()
	}
}

func (mv *ModsView) getSelectedMod() *help.ModRecord {
	for _, m := range mv.mods {
		if m.ID == mv.selectedID {
			return &m
		}
	}
	return nil
}

func (mv *ModsView) applySelectedMod() {
	m := mv.getSelectedMod()
	if m == nil {
		return
	}
	andLaunch := m.UseButterscotch
	mv.switchModWithSaveHandling(*m, andLaunch)
}

func (mv *ModsView) showEditModDialog() {
	m := mv.getSelectedMod()
	if m == nil || m.ID == VanillaModID {
		return
	}

	nameEntry := widget.NewEntry()
	nameEntry.SetText(m.Name)

	makerEntry := widget.NewEntry()
	makerEntry.SetText(m.Maker)

	cleanBase := strings.TrimSuffix(m.Base, "-w")
	baseEntry := widget.NewEntry()
	baseEntry.SetText(cleanBase)

	isMacCheck := widget.NewCheck("macOS Mod (Mac Mode: unchecked = Windows data mod)", nil)
	isMacCheck.SetChecked(!strings.HasSuffix(m.Base, "-w"))

	rootModeCheck := widget.NewCheck("Install to App Root (Root Mode: for custom macOS runners)", nil)
	rootModeCheck.SetChecked(m.InstallToAppRoot)

	var butterscotchCheck *widget.Check
	if help.IsButterscotchInstalled() {
		butterscotchCheck = widget.NewCheck("Use Butterscotch Runtime (experimental native GameMaker runner)", nil)
		butterscotchCheck.SetChecked(m.UseButterscotch)
	}

	vanillaSavesCheck := widget.NewCheck("Use Vanilla save files (shared with original game)", nil)
	vanillaSavesCheck.SetChecked(m.UseVanillaSaves)

	form := widget.NewForm(
		widget.NewFormItem("Mod Name:", nameEntry),
		widget.NewFormItem("Author / Maker:", makerEntry),
		widget.NewFormItem("Target Base:", baseEntry),
		widget.NewFormItem("", isMacCheck),
		widget.NewFormItem("", rootModeCheck),
	)
	if butterscotchCheck != nil {
		form.Append("", butterscotchCheck)
	}
	form.Append("", vanillaSavesCheck)

	d := dialog.NewCustomConfirm(
		"Edit Mod Metadata",
		"Save Changes",
		"Cancel",
		container.NewVBox(
			widget.NewLabel(fmt.Sprintf("Update metadata and configuration for mod '%s':", m.Name)),
			form,
		),
		func(confirm bool) {
			if !confirm {
				return
			}
			newName := strings.TrimSpace(nameEntry.Text)
			if newName == "" {
				dialog.ShowError(fmt.Errorf("mod name cannot be empty"), mv.window)
				return
			}
			newMaker := strings.TrimSpace(makerEntry.Text)
			newBase := strings.TrimSpace(baseEntry.Text)
			isMac := isMacCheck.Checked
			rootMode := rootModeCheck.Checked
			useButterscotch := m.UseButterscotch
			if butterscotchCheck != nil {
				useButterscotch = butterscotchCheck.Checked
			}
			useVanilla := vanillaSavesCheck.Checked

			mv.logger.Log(fmt.Sprintf("Updating metadata for mod '%s'...", m.Name))
			go func() {
				updated, err := help.UpdateModMetadataAction(m.ID, newName, newMaker, newBase, isMac, useVanilla, rootMode, useButterscotch, mv.logger.Log)
				if err != nil {
					dialog.ShowError(err, mv.window)
					return
				}
				mv.selectedID = updated.ID
				mv.Refresh()
				dialog.ShowInformation("Mod Updated", fmt.Sprintf("Mod '%s' metadata updated successfully!", updated.Name), mv.window)
			}()
		},
		mv.window,
	)

	d.Resize(fyne.NewSize(520, 390))
	d.Show()
}

func (mv *ModsView) switchModWithSaveHandling(m help.ModRecord, andLaunch bool) {
	targetModName := m.Name
	if m.ID == VanillaModID || m.UseVanillaSaves {
		targetModName = "Vanilla"
	}

	activeDir := help.GetUndertaleSaveDir()
	hasFiles := help.HasActiveSaveFiles(activeDir)
	activeInfo, _ := help.GetActiveSaveInfo()

	proceedToLoadModSave := func() {
		mv.prepareAndLoadTargetModSave(m, targetModName, andLaunch)
	}

	// 1. Save current game progress
	if hasFiles {
		if activeInfo != nil && activeInfo.Name != "" {
			// Save is already tracked - auto-save current progress
			mv.logger.Log(fmt.Sprintf("Auto-saving active game progress to '%s' (mod: %s)...", activeInfo.Name, activeInfo.ModName))
			if _, err := help.SaveCurrentGame(activeInfo.Name, activeInfo.ModName, true, mv.logger.Log); err != nil {
				mv.logger.Log(fmt.Sprintf("Warning: Failed to auto-save current game: %v", err))
			}
			proceedToLoadModSave()
		} else {
			// Untracked active save files! Ask user for save name before moving/switching
			nameEntry := widget.NewEntry()
			nameEntry.SetText("Previous Run")

			modOptions := []string{"Vanilla"}
			if mods, err := help.GetMods(); err == nil {
				for _, mod := range mods {
					modOptions = append(modOptions, mod.Name)
				}
			}
			modSelect := widget.NewSelect(modOptions, nil)
			modSelect.SetSelected("Vanilla")

			form := widget.NewForm(
				widget.NewFormItem("Save Profile Name:", nameEntry),
				widget.NewFormItem("Associated Mod:", modSelect),
			)

			d := dialog.NewCustomConfirm(
				"Save Active Game Progress",
				"Save & Continue",
				"Cancel",
				container.NewVBox(
					widget.NewLabel("Active Undertale save files were detected without a save profile name.\nPlease name your current save to preserve your progress before switching:"),
					form,
				),
				func(confirm bool) {
					if !confirm {
						mv.logger.Log("Mod switch cancelled.")
						return
					}
					saveName := strings.TrimSpace(nameEntry.Text)
					if saveName == "" {
						saveName = fmt.Sprintf("untracked_%s", time.Now().Format("20060102_150405"))
					}
					modName := modSelect.Selected
					if modName == "" {
						modName = "Vanilla"
					}
					if _, err := help.SaveCurrentGame(saveName, modName, true, mv.logger.Log); err != nil {
						dialog.ShowError(err, mv.window)
						return
					}
					proceedToLoadModSave()
				},
				mv.window,
			)
			d.Resize(fyne.NewSize(500, 240))
			d.Show()
		}
	} else {
		proceedToLoadModSave()
	}
}

func (mv *ModsView) prepareAndLoadTargetModSave(m help.ModRecord, targetModName string, andLaunch bool) {
	// If active save is already the target mod and tracked, proceed directly
	activeInfo, _ := help.GetActiveSaveInfo()
	if activeInfo != nil && strings.EqualFold(activeInfo.ModName, targetModName) && activeInfo.Name != "" {
		mv.finishApplyingAndLaunching(m, andLaunch)
		return
	}

	saves, err := help.GetSaves(targetModName)
	if err != nil {
		saves = nil
	}

	if len(saves) == 1 {
		// Exactly 1 save for target mod: load it directly
		mv.logger.Log(fmt.Sprintf("Loading save '%s' for mod '%s'...", saves[0].Name, targetModName))
		go func() {
			if err := help.LoadSave(saves[0].Name, targetModName, mv.logger.Log); err != nil {
				dialog.ShowError(err, mv.window)
				return
			}
			mv.finishApplyingAndLaunching(m, andLaunch)
		}()
		return
	}

	if len(saves) > 1 {
		// Multiple saves exist for target mod: let user select which one to load
		var options []string
		for _, s := range saves {
			options = append(options, s.Name)
		}
		saveSelect := widget.NewSelect(options, nil)
		saveSelect.SetSelected(saves[0].Name)

		form := widget.NewForm(
			widget.NewFormItem("Select Save Slot:", saveSelect),
		)

		d := dialog.NewCustomConfirm(
			fmt.Sprintf("Select Save for %s", targetModName),
			"Load Selected Save",
			"Cancel",
			container.NewVBox(
				widget.NewLabel(fmt.Sprintf("Multiple save profiles were found for '%s'. Which save would you like to load?", targetModName)),
				form,
			),
			func(confirm bool) {
				if !confirm {
					mv.logger.Log("Mod switch cancelled.")
					return
				}
				chosenSave := saveSelect.Selected
				if chosenSave == "" {
					chosenSave = saves[0].Name
				}
				go func() {
					if err := help.LoadSave(chosenSave, targetModName, mv.logger.Log); err != nil {
						dialog.ShowError(err, mv.window)
						return
					}
					mv.finishApplyingAndLaunching(m, andLaunch)
				}()
			},
			mv.window,
		)
		d.Resize(fyne.NewSize(480, 200))
		d.Show()
		return
	}

	// len(saves) == 0: No save data for the mod we are loading! Ask for it!
	nameEntry := widget.NewEntry()
	nameEntry.SetText("Save 1")

	allSaves, _ := help.GetSaves("")
	copyOptions := []string{"[Fresh Empty Save]"}
	for _, s := range allSaves {
		copyOptions = append(copyOptions, fmt.Sprintf("%s (%s)", s.Name, s.ModName))
	}
	copySelect := widget.NewSelect(copyOptions, nil)
	copySelect.SetSelected("[Fresh Empty Save]")

	form := widget.NewForm(
		widget.NewFormItem("New Save Name:", nameEntry),
		widget.NewFormItem("Base Save Data:", copySelect),
	)

	d := dialog.NewCustomConfirm(
		fmt.Sprintf("Create Save for %s", targetModName),
		"Create & Start",
		"Cancel",
		container.NewVBox(
			widget.NewLabel(fmt.Sprintf("No save data was found for mod '%s'.\nPlease specify how to create your save profile:", targetModName)),
			form,
		),
		func(confirm bool) {
			if !confirm {
				mv.logger.Log("Mod switch cancelled.")
				return
			}
			newSaveName := strings.TrimSpace(nameEntry.Text)
			if newSaveName == "" {
				newSaveName = "Save 1"
			}
			chosenBase := copySelect.Selected

			go func() {
				if chosenBase != "" && chosenBase != "[Fresh Empty Save]" {
					for _, s := range allSaves {
						label := fmt.Sprintf("%s (%s)", s.Name, s.ModName)
						if label == chosenBase {
							if _, err := help.CopySaveAndChangeMod(s.Name, s.ModName, newSaveName, targetModName, mv.logger.Log); err != nil {
								dialog.ShowError(err, mv.window)
								return
							}
							break
						}
					}
				} else {
					if _, err := help.CreateEmptySave(newSaveName, targetModName, mv.logger.Log); err != nil {
						dialog.ShowError(err, mv.window)
						return
					}
				}

				if err := help.LoadSave(newSaveName, targetModName, mv.logger.Log); err != nil {
					dialog.ShowError(err, mv.window)
					return
				}
				mv.finishApplyingAndLaunching(m, andLaunch)
			}()
		},
		mv.window,
	)
	d.Resize(fyne.NewSize(500, 240))
	d.Show()
}

func (mv *ModsView) finishApplyingAndLaunching(m help.ModRecord, andLaunch bool) {
	go func() {
		if m.ID == VanillaModID {
			mv.logger.Log("Restoring vanilla Undertale assets...")
			if err := help.RestoreBackupAction("1.08", "", true, mv.logger.Log); err != nil {
				dialog.ShowError(err, mv.window)
				return
			}
			mv.logger.Log("Undertale (Vanilla) ready.")
		} else {
			mv.logger.Log(fmt.Sprintf("Applying mod '%s' assets...", m.Name))
			if err := help.ApplyModAction(fmt.Sprintf("%d", m.ID), mv.logger.Log); err != nil {
				dialog.ShowError(err, mv.window)
				return
			}
			mv.logger.Log(fmt.Sprintf("Mod '%s' applied successfully!", m.Name))
		}

		if andLaunch {
			mv.logger.Log("Launching Undertale...")
			if err := help.LaunchUndertale(mv.logger.Log); err != nil {
				dialog.ShowError(err, mv.window)
				return
			}
			mv.logger.Log("Undertale is running!")
		} else {
			dialog.ShowInformation("Mod Ready", fmt.Sprintf("Mod '%s' and its save data are ready!", m.Name), mv.window)
		}
	}()
}

func (mv *ModsView) deleteSelectedMod() {
	m := mv.getSelectedMod()
	if m == nil || m.ID == VanillaModID {
		return
	}

	dialog.ShowConfirm(
		"Delete Mod",
		fmt.Sprintf("Are you sure you want to delete '%s'? This will remove the mod folder and database record.", m.Name),
		func(confirm bool) {
			if !confirm {
				return
			}
			go func() {
				if err := help.RemoveModAction(fmt.Sprintf("%d", m.ID), mv.logger.Log); err != nil {
					dialog.ShowError(err, mv.window)
					return
				}
				mv.Refresh()
			}()
		},
		mv.window,
	)
}

func (mv *ModsView) showAddModDialog() {
	folderEntry := widget.NewEntry()
	folderEntry.SetPlaceHolder("/path/to/mod/folder")

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Mod Name (e.g. Redux)")

	makerEntry := widget.NewEntry()
	makerEntry.SetPlaceHolder("Author Name (e.g. Toby)")

	baseEntry := widget.NewEntry()
	baseEntry.SetText("1.08")

	isMacCheck := widget.NewCheck("macOS Mod (Mac Mode: unchecked = Windows data mod)", nil)
	rootModeCheck := widget.NewCheck("Install to App Root (Root Mode: for custom macOS runners)", nil)

	var butterscotchCheck *widget.Check
	if help.IsButterscotchInstalled() {
		butterscotchCheck = widget.NewCheck("Use Butterscotch Runtime (experimental native GameMaker runner)", nil)
	}

	vanillaSavesCheck := widget.NewCheck("Use Vanilla save files (shared with original game)", nil)
	forceCheck := widget.NewCheck("Force overwrite existing mod folder", nil)

	updateFromFolder := func(dirPath string) {
		folderEntry.SetText(dirPath)
		if nameEntry.Text == "" {
			nameEntry.SetText(filepath.Base(dirPath))
		}
		cfg, err := help.ReadModConfig(dirPath)
		if err == nil && cfg != nil {
			if cfg.Metadata.Name != "" {
				nameEntry.SetText(cfg.Metadata.Name)
			}
			if cfg.Metadata.Author != "" {
				makerEntry.SetText(cfg.Metadata.Author)
			}
			if cfg.Metadata.GameVersion != "" {
				baseEntry.SetText(strings.TrimSuffix(cfg.Metadata.GameVersion, "-w"))
				if !strings.HasSuffix(cfg.Metadata.GameVersion, "-w") {
					isMacCheck.SetChecked(true)
				}
			}
			if cfg.Metadata.UseVanillaSaves {
				vanillaSavesCheck.SetChecked(true)
			}
			if cfg.Metadata.InstallToAppRoot {
				rootModeCheck.SetChecked(true)
			}
			if cfg.Metadata.UseButterscotch && butterscotchCheck != nil {
				butterscotchCheck.SetChecked(true)
			}
		}
	}

	pickFolderBtn := widget.NewButtonWithIcon("Browse...", theme.FolderOpenIcon(), func() {
		folderDialog := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			updateFromFolder(uri.Path())
		}, mv.window)
		folderDialog.Show()
	})

	form := widget.NewForm(
		widget.NewFormItem("Mod Folder:", container.NewBorder(nil, nil, nil, pickFolderBtn, folderEntry)),
		widget.NewFormItem("Mod Name:", nameEntry),
		widget.NewFormItem("Author / Maker:", makerEntry),
		widget.NewFormItem("Base Game Version:", baseEntry),
		widget.NewFormItem("", isMacCheck),
		widget.NewFormItem("", rootModeCheck),
	)
	if butterscotchCheck != nil {
		form.Append("", butterscotchCheck)
	}
	form.Append("", vanillaSavesCheck)
	form.Append("", forceCheck)

	d := dialog.NewCustomConfirm(
		"Add New Mod",
		"Install",
		"Cancel",
		container.NewVBox(
			widget.NewLabel("Select a mod folder containing .xdelta or game assets:"),
			form,
		),
		func(confirm bool) {
			if !confirm {
				return
			}
			folderPath := strings.TrimSpace(folderEntry.Text)
			if folderPath == "" {
				dialog.ShowError(fmt.Errorf("please choose a mod folder"), mv.window)
				return
			}

			name := strings.TrimSpace(nameEntry.Text)
			maker := strings.TrimSpace(makerEntry.Text)
			base := strings.TrimSpace(baseEntry.Text)
			isMac := isMacCheck.Checked
			rootMode := rootModeCheck.Checked
			useButterscotch := false
			if butterscotchCheck != nil {
				useButterscotch = butterscotchCheck.Checked
			}
			useVanilla := vanillaSavesCheck.Checked
			force := forceCheck.Checked

			go func() {
				_, err := help.AddModAction(folderPath, name, maker, base, isMac, useVanilla, rootMode, useButterscotch, force, mv.logger.Log)
				if err != nil {
					dialog.ShowError(err, mv.window)
					return
				}
				mv.Refresh()
			}()
		},
		mv.window,
	)

	d.Resize(fyne.NewSize(520, 450))
	d.Show()
}

func (mv *ModsView) showQuickPatchDialog() {
	fileEntry := widget.NewEntry()
	fileEntry.SetPlaceHolder("/path/to/patch.xdelta")

	winDataCheck := widget.NewCheck("Inject Windows data.win before patching", func(checked bool) {})
	winDataCheck.SetChecked(true)

	forceCheck := widget.NewCheck("Force patch (disable checksum verification)", nil)

	pickFileBtn := widget.NewButtonWithIcon("Browse...", theme.FileIcon(), func() {
		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			defer reader.Close()
			fileEntry.SetText(reader.URI().Path())
		}, mv.window)
		fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".xdelta", ".patch"}))
		fileDialog.Show()
	})

	form := widget.NewForm(
		widget.NewFormItem("Patch File:", container.NewBorder(nil, nil, nil, pickFileBtn, fileEntry)),
		widget.NewFormItem("", winDataCheck),
		widget.NewFormItem("", forceCheck),
	)

	d := dialog.NewCustomConfirm(
		"Quick Patch Undertale",
		"Apply Patch",
		"Cancel",
		container.NewVBox(
			widget.NewLabel("Apply an .xdelta patch directly to your Undertale installation:"),
			form,
		),
		func(confirm bool) {
			if !confirm {
				return
			}
			patchPath := strings.TrimSpace(fileEntry.Text)
			if patchPath == "" {
				dialog.ShowError(fmt.Errorf("please choose an .xdelta patch file"), mv.window)
				return
			}

			go func() {
				err := help.QuickPatchAction(patchPath, winDataCheck.Checked, forceCheck.Checked, mv.logger.Log)
				if err != nil {
					dialog.ShowError(err, mv.window)
					return
				}
				dialog.ShowInformation("Quick Patch", "Undertale was patched successfully!", mv.window)
			}()
		},
		mv.window,
	)

	d.Resize(fyne.NewSize(500, 240))
	d.Show()
}
