package ui

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"golang.org/x/term"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

const (
	ansiReset  = "\033[0m"
	ansiCyan   = "\033[36m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiBlue   = "\033[34m"
	ansiRed    = "\033[31m"
	ansiClear  = "\033[K"
)

type Config struct {
	Out       io.Writer
	Err       io.Writer
	IsTTY     bool
	IsQuiet   bool
	IsVerbose bool
	NoColor   bool
}

type spinner struct {
	stopChannel chan struct{}
	doneChannel chan struct{}
}

type UI struct {
	config     Config
	stateMutex sync.Mutex
	spinner    *spinner
}

func New(config Config) *UI {
	if config.Out == nil {
		config.Out = os.Stdout
	}
	if config.Err == nil {
		config.Err = os.Stderr
	}
	return &UI{config: config}
}

func IsTerminal(writer io.Writer) bool {
	if file, isFile := writer.(*os.File); isFile {
		return term.IsTerminal(int(file.Fd()))
	}
	return false
}

func HasNoColor() bool {
	return os.Getenv("NO_COLOR") != ""
}

func IsCI() bool {
	return os.Getenv("CI") != ""
}

func (uiInstance *UI) StartSpinner(message string) {
	if uiInstance.config.IsQuiet || !uiInstance.config.IsTTY {
		return
	}

	uiInstance.stateMutex.Lock()
	defer uiInstance.stateMutex.Unlock()

	if uiInstance.spinner != nil {
		return
	}

	stopChannel := make(chan struct{})
	doneChannel := make(chan struct{})
	uiInstance.spinner = &spinner{
		stopChannel: stopChannel,
		doneChannel: doneChannel,
	}

	if !uiInstance.config.NoColor {
		fmt.Fprintf(uiInstance.config.Err, "\r\033[36m%s\033[0m %s", spinnerFrames[0], message)
	} else {
		fmt.Fprintf(uiInstance.config.Err, "\r%s %s", spinnerFrames[0], message)
	}

	go func() {
		defer close(doneChannel)
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()
		frameIndex := 0

		for {
			select {
			case <-stopChannel:
				return
			case <-ticker.C:
				frameIndex = (frameIndex + 1) % len(spinnerFrames)
				uiInstance.stateMutex.Lock()
				if !uiInstance.config.NoColor {
					fmt.Fprintf(uiInstance.config.Err, "\r\033[36m%s\033[0m %s", spinnerFrames[frameIndex], message)
				} else {
					fmt.Fprintf(uiInstance.config.Err, "\r%s %s", spinnerFrames[frameIndex], message)
				}
				uiInstance.stateMutex.Unlock()
			}
		}
	}()
}

func (uiInstance *UI) StopSpinner() {
	uiInstance.stateMutex.Lock()
	if uiInstance.spinner == nil {
		uiInstance.stateMutex.Unlock()
		return
	}

	activeSpinner := uiInstance.spinner
	uiInstance.spinner = nil
	close(activeSpinner.stopChannel)
	uiInstance.stateMutex.Unlock()

	<-activeSpinner.doneChannel

	if !uiInstance.config.NoColor {
		fmt.Fprint(uiInstance.config.Err, "\r\033[K")
	} else {
		fmt.Fprint(uiInstance.config.Err, "\r")
	}
}

func (uiInstance *UI) Success(format string, arguments ...any) {
	if uiInstance.config.IsQuiet {
		return
	}
	uiInstance.writeStatus("✔", ansiGreen, format, arguments...)
}

func (uiInstance *UI) Info(format string, arguments ...any) {
	if uiInstance.config.IsQuiet {
		return
	}
	uiInstance.writeStatus("ℹ", ansiBlue, format, arguments...)
}

func (uiInstance *UI) Warn(format string, arguments ...any) {
	if uiInstance.config.IsQuiet {
		return
	}
	uiInstance.writeStatus("!", ansiYellow, format, arguments...)
}

func (uiInstance *UI) Error(format string, arguments ...any) {
	uiInstance.writeStatus("✖", ansiRed, format, arguments...)
}

func (uiInstance *UI) Verbose(format string, arguments ...any) {
	if !uiInstance.config.IsVerbose || uiInstance.config.IsQuiet {
		return
	}
	uiInstance.StopSpinner()

	message := fmt.Sprintf(format, arguments...)
	fmt.Fprintf(uiInstance.config.Err, "  %s\n", message)
}

func (uiInstance *UI) writeStatus(glyph string, colorCode string, format string, arguments ...any) {
	uiInstance.StopSpinner()

	message := fmt.Sprintf(format, arguments...)
	if uiInstance.config.IsTTY && !uiInstance.config.NoColor {
		fmt.Fprintf(uiInstance.config.Err, "%s%s%s %s\n", colorCode, glyph, ansiReset, message)
	} else {
		fmt.Fprintf(uiInstance.config.Err, "%s %s\n", glyph, message)
	}
}
