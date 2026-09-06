package gui

import (
	"UMMC/help"
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type BackupsView struct {
	window      fyne.Window
	logger      *ActivityLogger
	container   *fyne.Container
	backupList  *widget.List
	backups     []help.BackupRecord
	selectedID  int

	// Details widgets
	versionLabel *widget.Label
	idLabel      *widget.Label
	dateLabel    *widget.Label
	pathLabel    *widget.Label
	restoreBtn   *widget.Button
	deleteBtn    *widget.Button
}

func NewBackupsView(win fyne.Window, logger *ActivityLogger) *BackupsView {
	bv := &BackupsView{
		window:     win,
		logger:     logger,
		selectedID: -1,
	}

	bv.buildUI()
	bv.Refresh()
	return bv
}

func (bv *BackupsView) Container() *fyne.Container {
	return bv.container
}

func (bv *BackupsView) buildUI() {
	bv.backupList = widget.NewList(
		func() int {
			return len(bv.backups)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.StorageIcon()),
				widget.NewLabel("Version 1.08 (Date)"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < 0 || id >= len(bv.backups) {
				return
			}
			b := bv.backups[id]
			box := obj.(*fyne.Container)
			label := box.Objects[1].(*widget.Label)
			timeStr := b.CreatedAt.Local().Format("2006-01-02 15:04")
			label.SetText(fmt.Sprintf("Version: %s   [%s]   (ID: %d)", b.Version, timeStr, b.ID))
		},
	)

	bv.backupList.OnSelected = func(id widget.ListItemID) {
		if id >= 0 && id < len(bv.backups) {
			bv.selectedID = bv.backups[id].ID
			bv.showBackupDetails(bv.backups[id])
		}
	}

	// Details Pane
	bv.versionLabel = widget.NewLabelWithStyle("-", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	bv.idLabel = widget.NewLabel("-")
	bv.dateLabel = widget.NewLabel("-")
	bv.pathLabel = widget.NewLabel("-")

	bv.restoreBtn = widget.NewButtonWithIcon("Restore This Backup", theme.HistoryIcon(), func() {
		bv.restoreSelectedBackup()
	})
	bv.restoreBtn.Importance = widget.HighImportance
	bv.restoreBtn.Disable()

	bv.deleteBtn = widget.NewButtonWithIcon("Delete Backup", theme.DeleteIcon(), func() {
		bv.deleteSelectedBackup()
	})
	bv.deleteBtn.Importance = widget.DangerImportance
	bv.deleteBtn.Disable()

	detailsForm := widget.NewForm(
		widget.NewFormItem("Version Name:", bv.versionLabel),
		widget.NewFormItem("Database ID:", bv.idLabel),
		widget.NewFormItem("Created At:", bv.dateLabel),
		widget.NewFormItem("Backup Path:", bv.pathLabel),
	)

	actionButtons := container.NewGridWithColumns(2,
		bv.restoreBtn,
		bv.deleteBtn,
	)

	detailsCard := widget.NewCard("Backup Details", "", container.NewBorder(
		nil,
		actionButtons,
		nil,
		nil,
		container.NewVScroll(detailsForm),
	))

	// Header & Actions - Single Row
	createBtn := widget.NewButtonWithIcon("Create Backup", theme.ContentAddIcon(), func() {
		bv.showCreateBackupDialog()
	})
	refreshBtn := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		bv.Refresh()
	})

	topHeader := container.NewBorder(
		nil,
		nil,
		widget.NewLabelWithStyle("Undertale Backups", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewHBox(createBtn, refreshBtn),
	)

	leftPane := container.NewBorder(
		container.NewVBox(topHeader, widget.NewSeparator()),
		nil,
		nil,
		nil,
		bv.backupList,
	)

	split := container.NewHSplit(leftPane, detailsCard)
	split.SetOffset(0.55)

	bv.container = container.NewBorder(nil, nil, nil, nil, split)
}

func (bv *BackupsView) Refresh() {
	backups, err := help.GetBackups()
	if err != nil {
		bv.logger.Log(fmt.Sprintf("Failed to load backups: %v", err))
		return
	}
	bv.backups = backups
	bv.backupList.Refresh()

	if len(bv.backups) == 0 {
		bv.selectedID = -1
		bv.versionLabel.SetText("No backups found")
		bv.idLabel.SetText("-")
		bv.dateLabel.SetText("-")
		bv.pathLabel.SetText("-")
		bv.restoreBtn.Disable()
		bv.deleteBtn.Disable()
	} else if bv.selectedID != -1 {
		for i, b := range bv.backups {
			if b.ID == bv.selectedID {
				bv.backupList.Select(i)
				return
			}
		}
		bv.backupList.Select(0)
	}
}

func (bv *BackupsView) showBackupDetails(b help.BackupRecord) {
	bv.versionLabel.SetText(b.Version)
	bv.idLabel.SetText(fmt.Sprintf("%d", b.ID))
	bv.dateLabel.SetText(b.CreatedAt.Local().Format("2006-01-02 15:04:05"))
	bv.pathLabel.SetText(b.BackupPath)

	bv.restoreBtn.Enable()
	bv.deleteBtn.Enable()
}

func (bv *BackupsView) getSelectedBackup() *help.BackupRecord {
	for _, b := range bv.backups {
		if b.ID == bv.selectedID {
			return &b
		}
	}
	return nil
}

func (bv *BackupsView) restoreSelectedBackup() {
	b := bv.getSelectedBackup()
	if b == nil {
		return
	}

	dialog.ShowConfirm(
		"Restore Backup",
		fmt.Sprintf("Restore Undertale backup version '%s'? This will overwrite the current game files in your Undertale installation.", b.Version),
		func(confirm bool) {
			if !confirm {
				return
			}
			bv.logger.Log(fmt.Sprintf("Restoring backup '%s'...", b.Version))
			go func() {
				if err := help.RestoreBackupAction(fmt.Sprintf("%d", b.ID), "", true, bv.logger.Log); err != nil {
					dialog.ShowError(err, bv.window)
					return
				}
				dialog.ShowInformation("Backup Restored", fmt.Sprintf("Undertale version '%s' was restored successfully!", b.Version), bv.window)
			}()
		},
		bv.window,
	)
}

func (bv *BackupsView) deleteSelectedBackup() {
	b := bv.getSelectedBackup()
	if b == nil {
		return
	}

	dialog.ShowConfirm(
		"Delete Backup",
		fmt.Sprintf("Are you sure you want to delete backup '%s'? This will remove the backup files from disk.", b.Version),
		func(confirm bool) {
			if !confirm {
				return
			}
			go func() {
				if err := help.RemoveBackupAction(fmt.Sprintf("%d", b.ID), bv.logger.Log); err != nil {
					dialog.ShowError(err, bv.window)
					return
				}
				bv.Refresh()
			}()
		},
		bv.window,
	)
}

func (bv *BackupsView) showCreateBackupDialog() {
	pathEntry := widget.NewEntry()
	pathEntry.SetText(help.FindUndertaleApp(""))

	versionEntry := widget.NewEntry()
	versionEntry.SetText("1.08")

	forceCheck := widget.NewCheck("Force overwrite existing backup if already present", nil)

	form := widget.NewForm(
		widget.NewFormItem("Undertale App Path:", pathEntry),
		widget.NewFormItem("Version Label:", versionEntry),
		widget.NewFormItem("", forceCheck),
	)

	d := dialog.NewCustomConfirm(
		"Create Undertale Backup",
		"Create Backup",
		"Cancel",
		container.NewVBox(
			widget.NewLabel("Back up your current Undertale installation:"),
			form,
		),
		func(confirm bool) {
			if !confirm {
				return
			}
			gamePath := strings.TrimSpace(pathEntry.Text)
			version := strings.TrimSpace(versionEntry.Text)
			if version == "" {
				version = "1.08"
			}

			go func() {
				err := help.CreateBackupAction(gamePath, version, forceCheck.Checked, bv.logger.Log)
				if err != nil {
					dialog.ShowError(err, bv.window)
					return
				}
				bv.Refresh()
				dialog.ShowInformation("Backup Created", fmt.Sprintf("Backup '%s' created successfully!", version), bv.window)
			}()
		},
		bv.window,
	)

	d.Resize(fyne.NewSize(520, 260))
	d.Show()
}
