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

type ToolsView struct {
	window       fyne.Window
	app          fyne.App
	logger       *ActivityLogger
	container    *fyne.Container
	onRefreshAll func()

	// Status widgets
	gameStatusLabel        *widget.Label
	winDataLabel           *widget.Label
	xdeltaLabel            *widget.Label
	steamcmdLabel          *widget.Label
	butterscotchLabel      *widget.Label
	usernameEntry          *widget.Entry
	downloadTermBtn        *widget.Button
	injectWinBtn           *widget.Button
	butterscotchArchSelect *widget.Select
}

func NewToolsView(win fyne.Window, a fyne.App, logger *ActivityLogger, onRefreshAll func()) *ToolsView {
	tv := &ToolsView{
		window:       win,
		app:          a,
		logger:       logger,
		onRefreshAll: onRefreshAll,
	}

	tv.buildUI()
	tv.Refresh()
	return tv
}

func (tv *ToolsView) Container() *fyne.Container {
	return tv.container
}

func (tv *ToolsView) buildUI() {
	// 1. System Status Card
	tv.gameStatusLabel = widget.NewLabel("Checking...")
	tv.winDataLabel = widget.NewLabel("Checking...")
	tv.xdeltaLabel = widget.NewLabel("Checking...")
	tv.steamcmdLabel = widget.NewLabel("Checking...")
	tv.butterscotchLabel = widget.NewLabel("Checking...")

	statusRefreshBtn := widget.NewButtonWithIcon("Refresh Diagnostics", theme.ViewRefreshIcon(), func() {
		tv.Refresh()
	})

	onboardingBtn := widget.NewButtonWithIcon("Open Onboarding / Setup Guide", theme.HelpIcon(), func() {
		ShowOnboardingWizard(tv.window, tv.app, tv.logger, tv.onRefreshAll)
	})

	statusForm := widget.NewForm(
		widget.NewFormItem("Undertale Installation:", tv.gameStatusLabel),
		widget.NewFormItem("Windows data.win:", tv.winDataLabel),
		widget.NewFormItem("xdelta3 Tool:", tv.xdeltaLabel),
		widget.NewFormItem("SteamCMD Tool:", tv.steamcmdLabel),
		widget.NewFormItem("Butterscotch Runner:", tv.butterscotchLabel),
	)

	statusCard := widget.NewCard(
		"System Status & Diagnostics",
		"Dependencies and game detection status",
		container.NewVBox(statusForm, container.NewHBox(statusRefreshBtn, onboardingBtn)),
	)

	// 2. Butterscotch Runner Downloader Card
	tv.butterscotchArchSelect = widget.NewSelect([]string{
		"Auto-detect (Recommended)",
		"Apple Silicon (arm64)",
		"Intel (x86_64)",
	}, nil)
	tv.butterscotchArchSelect.SetSelected("Auto-detect (Recommended)")

	downloadButterscotchBtn := widget.NewButtonWithIcon("Download Butterscotch Runtime", theme.DownloadIcon(), func() {
		tv.downloadButterscotch()
	})
	downloadButterscotchBtn.Importance = widget.HighImportance

	openButterscotchBtn := widget.NewButtonWithIcon("Open Runtime Folder", theme.FolderIcon(), func() {
		_ = help.OpenFolder(help.DefaultButterscotchDir)
	})

	butterscotchCard := widget.NewCard(
		"Download Experimental Butterscotch Runner",
		"Open-source native GameMaker runner for macOS (Apple Silicon & Intel)",
		container.NewVBox(
			widget.NewLabel("Downloads and extracts the latest Butterscotch native runner from nightly.link.\nUseful for running GameMaker mods and standalone runner executables on macOS:"),
			widget.NewForm(
				widget.NewFormItem("Architecture:", tv.butterscotchArchSelect),
			),
			container.NewHBox(downloadButterscotchBtn, openButterscotchBtn),
		),
	)

	// 3. Windows Downloader Card
	tv.usernameEntry = widget.NewEntry()
	tv.usernameEntry.SetPlaceHolder("Steam username (optional for public / saved login)")

	tv.downloadTermBtn = widget.NewButtonWithIcon("Download via SteamCMD (Opens Terminal)", theme.ComputerIcon(), func() {
		tv.downloadInTerminal()
	})
	tv.downloadTermBtn.Importance = widget.HighImportance

	steamcmdCard := widget.NewCard(
		"Download Undertale for Windows",
		"Uses SteamCMD to fetch the Windows depot (needed for Windows mods / data.win)",
		container.NewVBox(
			widget.NewLabel("SteamCMD runs in Terminal to support interactive Steam password and Steam Guard 2FA prompts:"),
			widget.NewForm(
				widget.NewFormItem("Steam Username:", tv.usernameEntry),
			),
			tv.downloadTermBtn,
		),
	)

	// 4. Windows Data Injector Card
	tv.injectWinBtn = widget.NewButtonWithIcon("Inject data.win into Undertale", theme.MediaReplayIcon(), func() {
		tv.injectWindowsData()
	})

	injectCard := widget.NewCard(
		"Windows Data Injection",
		"Replace macOS game.ios with Windows data.win for Windows-based mod compatibility",
		container.NewVBox(
			widget.NewLabel("Copies ~/UMMC/windows/data.win into Undertale's game.ios and sets the winpatchdetect marker."),
			tv.injectWinBtn,
		),
	)

	// 5. Folder Shortcuts & Quick Launch Card
	openGameBtn := widget.NewButtonWithIcon("Game Folder", theme.FolderIcon(), func() {
		_ = help.OpenFolder(help.DefaultSteamUndertaleDir)
	})
	openModsBtn := widget.NewButtonWithIcon("Mods Folder", theme.FolderIcon(), func() {
		_ = help.OpenFolder(help.DefaultModsDir)
	})
	openBackupsBtn := widget.NewButtonWithIcon("Backups Folder", theme.FolderIcon(), func() {
		_ = help.OpenFolder(help.DefaultBackupsDir)
	})
	openUmmcBtn := widget.NewButtonWithIcon("UMMC Home", theme.HomeIcon(), func() {
		_ = help.OpenFolder("~/UMMC")
	})

	launchGameBtn := widget.NewButtonWithIcon("Launch Undertale Game", theme.MediaPlayIcon(), func() {
		go func() {
			if err := help.LaunchUndertale(tv.logger.Log); err != nil {
				dialog.ShowError(err, tv.window)
			}
		}()
	})
	launchGameBtn.Importance = widget.HighImportance

	shortcutsCard := widget.NewCard(
		"Quick Shortcuts & Launch",
		"Open directories in Finder or start Undertale",
		container.NewVBox(
			launchGameBtn,
			container.NewHBox(openGameBtn, openModsBtn, openBackupsBtn, openUmmcBtn),
		),
	)

	// 6. Complete Reset / Nuke Card
	nukeBtn := widget.NewButtonWithIcon("Nuke All Data (Complete Reset)", theme.DeleteIcon(), func() {
		tv.showNukeDialog()
	})
	nukeBtn.Importance = widget.DangerImportance

	dangerCard := widget.NewCard(
		"Danger Zone - Complete Reset",
		"Wipe all mods, backups, downloaded Windows data, database, and game files",
		container.NewVBox(
			widget.NewLabel("Reset UMMC to a fresh, clean state. You can also optionally delete modified game files."),
			nukeBtn,
		),
	)

	content := container.NewVBox(
		statusCard,
		shortcutsCard,
		butterscotchCard,
		steamcmdCard,
		injectCard,
		dangerCard,
	)

	tv.container = container.NewBorder(nil, nil, nil, nil, container.NewVScroll(content))
}

func (tv *ToolsView) Refresh() {
	status := help.CheckSystemStatus()

	if status.UndertaleInstalled {
		tv.gameStatusLabel.SetText(fmt.Sprintf("Found: %s", status.UndertaleAppPath))
	} else {
		tv.gameStatusLabel.SetText("Not Found (Check Steam Undertale install)")
	}

	if status.WindowsDataExists {
		tv.winDataLabel.SetText(fmt.Sprintf("Found: %s", status.WindowsDataPath))
		tv.injectWinBtn.Enable()
	} else {
		tv.winDataLabel.SetText("Missing (Use downloader below)")
		tv.injectWinBtn.Disable()
	}

	if status.XdeltaAvailable {
		tv.xdeltaLabel.SetText(fmt.Sprintf("Installed: %s", status.XdeltaPath))
	} else {
		tv.xdeltaLabel.SetText("Not Found (Install via: brew install xdelta)")
	}

	if status.SteamcmdAvailable {
		tv.steamcmdLabel.SetText(fmt.Sprintf("Installed: %s", status.SteamcmdPath))
		if tv.downloadTermBtn != nil {
			tv.downloadTermBtn.Enable()
		}
	} else {
		tv.steamcmdLabel.SetText("Not Found (Install via: brew install steamcmd)")
		if tv.downloadTermBtn != nil {
			tv.downloadTermBtn.Disable()
		}
	}

	if status.ButterscotchAvailable {
		tv.butterscotchLabel.SetText(fmt.Sprintf("Installed: %s", status.ButterscotchPath))
	} else {
		tv.butterscotchLabel.SetText("Not Installed (Download below)")
	}
}

func (tv *ToolsView) downloadButterscotch() {
	var arch string
	if tv.butterscotchArchSelect != nil {
		switch tv.butterscotchArchSelect.Selected {
		case "Apple Silicon (arm64)":
			arch = "arm64"
		case "Intel (x86_64)":
			arch = "x86_64"
		default:
			arch = "auto"
		}
	}

	tv.logger.Log(fmt.Sprintf("Starting download for Butterscotch runtime (%s)...", arch))
	go func() {
		dest, err := help.DownloadButterscotchRuntime(arch, "", "", tv.logger.Log)
		if err != nil {
			dialog.ShowError(err, tv.window)
			return
		}
		tv.Refresh()
		dialog.ShowInformation(
			"Download Complete",
			fmt.Sprintf("Butterscotch runner was successfully downloaded and extracted to:\n%s", dest),
			tv.window,
		)
	}()
}

func (tv *ToolsView) downloadInTerminal() {
	username := strings.TrimSpace(tv.usernameEntry.Text)
	tv.logger.Log("Opening interactive SteamCMD session in Terminal.app...")
	if err := help.LaunchSteamCMDInTerminal(username); err != nil {
		dialog.ShowError(err, tv.window)
	}
}

func (tv *ToolsView) injectWindowsData() {
	dialog.ShowConfirm(
		"Inject Windows data.win",
		"This will replace Contents/Resources/game.ios in Undertale with ~/UMMC/windows/data.win.\n\nMake sure you have backed up your vanilla Undertale first!",
		func(confirm bool) {
			if !confirm {
				return
			}
			go func() {
				if err := help.InjectWindowsDataAction("", "", tv.logger.Log); err != nil {
					dialog.ShowError(err, tv.window)
					return
				}
				dialog.ShowInformation("Injected", "Windows data.win successfully injected into Undertale!", tv.window)
			}()
		},
		tv.window,
	)
}

func (tv *ToolsView) showNukeDialog() {
	deleteGameFilesCheck := widget.NewCheck("Also delete Undertale game files (UNDERTALE.app) from Steam folder", nil)

	customContent := container.NewVBox(
		widget.NewLabelWithStyle("WARNING: This is a destructive action.", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("This will permanently delete:\n• All installed mods in ~/UMMC/mods\n• All backups in ~/UMMC/Backup\n• Downloaded Windows files in ~/UMMC/windows\n• The entire UMMC SQLite database"),
		deleteGameFilesCheck,
		widget.NewLabel("You will be starting from a completely clean slate."),
	)

	d := dialog.NewCustomConfirm(
		"Nuke All Data (Complete Reset)",
		"Nuke Everything",
		"Cancel",
		customContent,
		func(confirm bool) {
			if !confirm {
				return
			}
			deleteGame := deleteGameFilesCheck.Checked
			tv.logger.Log("Starting complete data nuke...")
			go func() {
				err := help.NukeAllDataAction(deleteGame, tv.logger.Log)
				if err != nil {
					dialog.ShowError(err, tv.window)
					return
				}
				if tv.onRefreshAll != nil {
					tv.onRefreshAll()
				}
				dialog.ShowInformation("Reset Complete", "All UMMC data and files have been completely reset.", tv.window)
			}()
		},
		tv.window,
	)

	d.Resize(fyne.NewSize(540, 280))
	d.Show()
}
