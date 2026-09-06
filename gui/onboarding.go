package gui

import (
	"UMMC/help"
	"fmt"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type OnboardingWizard struct {
	window      fyne.Window
	app         fyne.App
	logger      *ActivityLogger
	currentStep int
	totalSteps  int
	stepTitle   *widget.Label
	stepDesc    *widget.Label
	stepBody    *fyne.Container
	prevBtn     *widget.Button
	nextBtn     *widget.Button
	dialog      dialog.Dialog
	onComplete  func()
}

func ShowOnboardingWizard(win fyne.Window, a fyne.App, logger *ActivityLogger, onComplete func()) {
	wizard := &OnboardingWizard{
		window:      win,
		app:         a,
		logger:      logger,
		currentStep: 0,
		totalSteps:  5,
		onComplete:  onComplete,
	}

	wizard.buildUI()
}

func (ow *OnboardingWizard) buildUI() {
	ow.stepTitle = widget.NewLabelWithStyle("Step Title", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	ow.stepDesc = widget.NewLabel("Step Description")
	ow.stepDesc.Wrapping = fyne.TextWrapWord
	ow.stepBody = container.NewVBox()

	ow.prevBtn = widget.NewButtonWithIcon("Back", theme.NavigateBackIcon(), func() {
		if ow.currentStep > 0 {
			ow.currentStep--
			ow.renderStep()
		}
	})

	ow.nextBtn = widget.NewButtonWithIcon("Next", theme.NavigateNextIcon(), func() {
		if ow.currentStep < ow.totalSteps-1 {
			ow.currentStep++
			ow.renderStep()
		} else {
			ow.finish()
		}
	})
	ow.nextBtn.Importance = widget.HighImportance

	navBar := container.NewBorder(
		nil,
		nil,
		ow.prevBtn,
		ow.nextBtn,
		widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{}),
	)

	content := container.NewBorder(
		container.NewVBox(ow.stepTitle, ow.stepDesc, widget.NewSeparator()),
		container.NewVBox(widget.NewSeparator(), navBar),
		nil,
		nil,
		container.NewVScroll(ow.stepBody),
	)

	d := dialog.NewCustom(
		"UMMC Setup & Onboarding",
		"Close",
		content,
		ow.window,
	)

	d.Resize(fyne.NewSize(640, 480))
	ow.dialog = d
	ow.renderStep()
	d.Show()
}

func (ow *OnboardingWizard) renderStep() {
	ow.stepBody.Objects = nil

	if ow.currentStep == 0 {
		ow.prevBtn.Disable()
	} else {
		ow.prevBtn.Enable()
	}

	if ow.currentStep == ow.totalSteps-1 {
		ow.nextBtn.SetText("Finish Setup")
		ow.nextBtn.SetIcon(theme.ConfirmIcon())
	} else {
		ow.nextBtn.SetText("Next")
		ow.nextBtn.SetIcon(theme.NavigateNextIcon())
	}

	switch ow.currentStep {
	case 0:
		ow.stepTitle.SetText("Welcome to UMMC (Undertale Mac Mod Manager)")
		ow.stepDesc.SetText("This setup guide will help you configure Undertale, back up your original game files, and set up Windows mod compatibility.")

		ow.stepBody.Add(widget.NewLabel("What this setup wizard will do:"))
		ow.stepBody.Add(widget.NewLabel("1. Check installed tools (Undertale, xdelta3, SteamCMD)"))
		ow.stepBody.Add(widget.NewLabel("2. Create a vanilla Undertale backup (1.08)"))
		ow.stepBody.Add(widget.NewLabel("3. Set up Windows data (data.win) for Windows-based mods"))
		ow.stepBody.Add(widget.NewLabel("4. Get everything ready for 1-click modding!"))

	case 1:
		ow.stepTitle.SetText("Step 1: Dependency & Game Check")
		ow.stepDesc.SetText("Let's make sure Undertale and the required modding tools are found on your Mac.")

		status := help.CheckSystemStatus()

		var gameTxt string
		if status.UndertaleInstalled {
			gameTxt = fmt.Sprintf("Found at: %s", status.UndertaleAppPath)
		} else {
			gameTxt = "Not Found (Install Undertale via Steam)"
		}

		var xdeltaTxt string
		if status.XdeltaAvailable {
			xdeltaTxt = fmt.Sprintf("Installed: %s", status.XdeltaPath)
		} else {
			xdeltaTxt = "Missing (Run in Terminal: brew install xdelta)"
		}

		var steamcmdTxt string
		if status.SteamcmdAvailable {
			steamcmdTxt = fmt.Sprintf("Installed: %s", status.SteamcmdPath)
		} else {
			steamcmdTxt = "Optional for Windows depot (brew install steamcmd)"
		}

		form := widget.NewForm(
			widget.NewFormItem("Undertale (macOS):", widget.NewLabel(gameTxt)),
			widget.NewFormItem("xdelta3 Tool:", widget.NewLabel(xdeltaTxt)),
			widget.NewFormItem("SteamCMD Tool:", widget.NewLabel(steamcmdTxt)),
		)

		refreshBtn := widget.NewButtonWithIcon("Re-check Status", theme.ViewRefreshIcon(), func() {
			ow.renderStep()
		})

		ow.stepBody.Add(form)
		ow.stepBody.Add(refreshBtn)

	case 2:
		ow.stepTitle.SetText("Step 2: Create Vanilla Game Backup")
		ow.stepDesc.SetText("Before installing mods or Windows data, create a clean backup of your vanilla Undertale game so you can revert back at any time.")

		statusMsg := widget.NewLabel("Status: Ready to backup.")
		backupBtn := widget.NewButtonWithIcon("Create Vanilla Backup (Version 1.08)", theme.ContentAddIcon(), func() {
			statusMsg.SetText("Backing up Undertale...")
			go func() {
				err := help.CreateBackupAction("", "1.08", true, ow.logger.Log)
				if err != nil {
					statusMsg.SetText(fmt.Sprintf("Error: %v", err))
				} else {
					statusMsg.SetText("Success: Vanilla backup (1.08) created!")
					if ow.onComplete != nil {
						ow.onComplete()
					}
				}
			}()
		})
		backupBtn.Importance = widget.HighImportance

		ow.stepBody.Add(widget.NewLabel("Click below to create your initial backup:"))
		ow.stepBody.Add(backupBtn)
		ow.stepBody.Add(statusMsg)

	case 3:
		ow.stepTitle.SetText("Step 3: Windows Setup (data.win)")
		ow.stepDesc.SetText("Most Undertale mods are built for the Windows data.win file. Here you can download or import data.win and set up Windows mod support.")

		status := help.CheckSystemStatus()
		winStatusLabel := widget.NewLabel("Checking...")
		if status.WindowsDataExists {
			winStatusLabel.SetText(fmt.Sprintf("Found: %s", status.WindowsDataPath))
		} else {
			winStatusLabel.SetText("Not yet downloaded / imported")
		}

		statusMsg := widget.NewLabel("")

		// Downloader widgets
		userEntry := widget.NewEntry()
		userEntry.SetPlaceHolder("Steam username (optional if already logged in)")

		termBtn := widget.NewButtonWithIcon("Download via SteamCMD (Opens Terminal)", theme.ComputerIcon(), func() {
			username := strings.TrimSpace(userEntry.Text)
			if err := help.LaunchSteamCMDInTerminal(username); err != nil {
				statusMsg.SetText(fmt.Sprintf("Failed to launch Terminal: %v", err))
			} else {
				statusMsg.SetText("SteamCMD launched in Terminal. Log in there, then click 'Check Status' below.")
			}
		})
		termBtn.Importance = widget.HighImportance

		checkBtn := widget.NewButtonWithIcon("Check Status", theme.ViewRefreshIcon(), func() {
			ow.renderStep()
		})

		// File picker to import data.win manually
		browseBtn := widget.NewButtonWithIcon("Select Existing data.win...", theme.FolderOpenIcon(), func() {
			fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
				if err != nil || reader == nil {
					return
				}
				defer reader.Close()
				srcPath := reader.URI().Path()
				dstPath := help.ExpandPath(help.DefaultWindowsDataPath)
				_ = os.MkdirAll(help.ExpandPath("~/UMMC/windows/"), 0755)
				if copyErr := help.CopyFile(srcPath, dstPath, true); copyErr != nil {
					statusMsg.SetText(fmt.Sprintf("Import error: %v", copyErr))
				} else {
					statusMsg.SetText("Imported data.win successfully!")
					ow.renderStep()
				}
			}, ow.window)
			fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".win"}))
			fileDialog.Show()
		})

		// Inject & Create 1.08-w Backup
		injectBtn := widget.NewButtonWithIcon("Inject data.win & Create Windows Backup (1.08-w)", theme.ConfirmIcon(), func() {
			statusMsg.SetText("Injecting Windows data.win and creating backup 1.08-w...")
			go func() {
				if err := help.InjectWindowsDataAction("", "", ow.logger.Log); err != nil {
					statusMsg.SetText(fmt.Sprintf("Injection error: %v", err))
					return
				}
				if err := help.CreateBackupAction("", "1.08-w", true, ow.logger.Log); err != nil {
					statusMsg.SetText(fmt.Sprintf("Windows backup error: %v", err))
					return
				}
				statusMsg.SetText("Success: Windows data injected & 1.08-w backup created!")
				if ow.onComplete != nil {
					ow.onComplete()
				}
			}()
		})
		injectBtn.Importance = widget.HighImportance

		if !status.WindowsDataExists {
			injectBtn.Disable()
		}

		ow.stepBody.Add(widget.NewForm(
			widget.NewFormItem("Windows data.win Status:", winStatusLabel),
		))

		ow.stepBody.Add(widget.NewLabelWithStyle("Option A: Download Windows depot with SteamCMD", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
		ow.stepBody.Add(widget.NewForm(widget.NewFormItem("Steam Username:", userEntry)))
		ow.stepBody.Add(container.NewHBox(termBtn, checkBtn))

		ow.stepBody.Add(widget.NewSeparator())
		ow.stepBody.Add(widget.NewLabelWithStyle("Option B: Import existing data.win file", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
		ow.stepBody.Add(browseBtn)

		ow.stepBody.Add(widget.NewSeparator())
		ow.stepBody.Add(widget.NewLabelWithStyle("Step 3B: Configure Windows Base", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
		ow.stepBody.Add(injectBtn)
		ow.stepBody.Add(statusMsg)

	case 4:
		ow.stepTitle.SetText("Step 4: Setup Complete!")
		ow.stepDesc.SetText("UMMC is fully configured and ready.")

		ow.stepBody.Add(widget.NewLabel("You can now:"))
		ow.stepBody.Add(widget.NewLabel("• Go to the Mods tab to add, quick-patch, and play mods."))
		ow.stepBody.Add(widget.NewLabel("• Go to the Backups tab to switch between vanilla (1.08) and Windows base (1.08-w)."))
		ow.stepBody.Add(widget.NewLabel("• Go to Tools & Windows anytime to manage files or perform a complete reset."))
	}

	ow.stepBody.Refresh()
}

func (ow *OnboardingWizard) finish() {
	ow.app.Preferences().SetBool("onboarding_completed", true)
	if ow.dialog != nil {
		ow.dialog.Hide()
	}
	if ow.onComplete != nil {
		ow.onComplete()
	}
	ow.logger.Log("Setup guide completed! Ready to mod Undertale.")
}
