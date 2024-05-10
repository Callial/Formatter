package main

import (
	"Formatter/internal"
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func main() {
	a := app.NewWithID("com.novellaforge.formatter")
	w := a.NewWindow("Formatter")
	go NewMainContent(w)
	w.ShowAndRun()
}

func NewMainContent(w fyne.Window) {
	w.SetMainMenu(NewMainMenu(w))
	formatAllButton := NewFormatAllButton(w)
	formatVideoButton := NewFormatVideoButton(w)
	formatAudioButton := NewFormatAudioButton(w)
	vbox := container.NewVBox(formatAllButton, formatVideoButton, formatAudioButton)
	w.SetContent(vbox)
}

func NewMainMenu(w fyne.Window) *fyne.MainMenu {
	return fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("Settings", func() {
				dialog.ShowCustom("Settings", "Close",
					widget.NewForm(
						widget.NewFormItem("Binary Location", widget.NewButton(
							"Chose Location", func() {
								userDir, err := os.UserConfigDir()
								if err != nil {
									dialog.ShowError(err, w)
									return
								}
								defaultDir := userDir + "/NovellaForge/Formatter/bin"
								curLocation := fyne.CurrentApp().Preferences().StringWithFallback("binaryLocation", defaultDir)
								dirDialog := dialog.NewFolderOpen(func(dir fyne.ListableURI, err error) {
									if err == nil && dir != nil {
										fyne.CurrentApp().Preferences().SetString("binaryLocation", dir.Path())
									}
								}, w)
								lister, err := storage.ListerForURI(storage.NewFileURI(curLocation))
								if err != nil {
									dialog.ShowError(err, w)
									return
								}
								dirDialog.SetLocation(lister)
								dirDialog.Show()
							})),
					), w)
			}),
		),
		fyne.NewMenu("Help",
			fyne.NewMenuItem("About", func() {
				dialog.ShowInformation("About", "Formatter is a tool to format audio and video files.", w)
			}),
			fyne.NewMenuItem("Documentation", func() {
				dialog.ShowInformation("Documentation", "Documentation is not available yet.", w)
			}),
		),
	)

}

func NewFormatAudioButton(w fyne.Window) fyne.CanvasObject {
	return widget.NewButton("Format Audio", func() {
		err := CheckBinaries(w)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
	})
}

func NewFormatVideoButton(w fyne.Window) fyne.CanvasObject {

}

func NewFormatAllButton(w fyne.Window) fyne.CanvasObject {

}

func CheckBinaries(w fyne.Window) error {
	userConfig, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	defaultLocation := userConfig + "/NovellaForge/Formatter/bin"
	binaryLocation := fyne.CurrentApp().Preferences().StringWithFallback("binaryLocation", defaultLocation)
	FfmpegPath := filepath.Join(binaryLocation, "ffmpeg")
	ProbePath := filepath.Join(binaryLocation, "ffprobe")

	_, err = exec.LookPath(FfmpegPath)
	if err != nil {
		err = UnpackBinaries(binaryLocation, w)
		if err != nil {
			return err
		}
	}
	_, err = exec.LookPath(ProbePath)
	if err != nil {
		err = UnpackBinaries(binaryLocation, w)
		if err != nil {
			return err
		}
	}

	cmd := exec.Command(FfmpegPath, "-version")
	_, err = cmd.CombinedOutput()
	if err != nil {
		return errors.New("ffmpeg binary is invalid")
	}
	cmd = exec.Command(ProbePath, "-version")
	_, err = cmd.CombinedOutput()
	if err != nil {
		return errors.New("ffprobe binary is invalid")
	}
	return nil
}

func UnpackBinaries(location string, w fyne.Window) error {
	switch runtime.GOOS {
	case "windows":
		return UnpackWindowsBinaries(location, w)
	default:
		return errors.New("unsupported OS")
	}
}

func UnpackWindowsBinaries(location string, w fyne.Window) error {
	// Check if the binaries are present in the embeddedBinaries
	ffmpeg, err := internal.EmbeddedBinaries.Open(filepath.Join("win", "ffmpeg.exe"))
	defer func() {
		if cerr := ffmpeg.Close(); cerr != nil {
			log.Println("Error closing internal ffmpeg:", cerr)
		}
	}()
	if err != nil {
		return err
	}
	ffprobe, err := internal.EmbeddedBinaries.Open(filepath.Join("win", "ffprobe.exe"))
	defer func() {
		if cerr := ffprobe.Close(); cerr != nil {
			log.Println("Error closing internal ffprobe:", cerr)
		}
	}()
	if err != nil {
		return err
	}
	//Copy the binaries to the specified location
	ffmpegPath := filepath.Join(location, "ffmpeg.exe")
	ffprobePath := filepath.Join(location, "ffprobe.exe")
	ffmpegFile, err := os.Create(ffmpegPath)
	defer func() {
		if cerr := ffmpegFile.Close(); cerr != nil {
			log.Println("Error closing external ffmpeg:", cerr)
		}
	}()
	if err != nil {
		return err
	}
	err = ffmpegFile.Chmod(os.ModePerm)
	if err != nil {
		return err
	}
	ffprobeFile, err := os.Create(ffprobePath)
	defer func() {
		if cerr := ffprobeFile.Close(); cerr != nil {
			log.Println("Error closing external ffprobe:", cerr)
		}
	}()
	if err != nil {
		return err
	}
	err = ffprobeFile.Chmod(os.ModePerm)
	if err != nil {
		return err
	}

	_, err = io.Copy(ffmpegFile, ffmpeg)
	if err != nil {
		return err
	}
	_, err = io.Copy(ffprobeFile, ffprobe)
	if err != nil {
		return err
	}
	return nil
}
