// Still Watching - Eye inactivity monitor

package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	systray "github.com/getlantern/systray"
)

var (
	timerDurations = map[string]time.Duration{
		"5 minutes":  5 * time.Minute,
		"10 minutes": 10 * time.Minute,
		"15 minutes": 15 * time.Minute,
	}

	currentDuration = timerDurations["5 minutes"]
	monitoring      = false
	timerItems      = make(map[string]*systray.MenuItem)
)

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(iconActive)
	systray.SetTitle("Still Watching")
	systray.SetTooltip("Eye Monitor")

	startStopItem := systray.AddMenuItem("Deactivate", "Toggle eye monitoring")
	timerMenu := systray.AddMenuItem("Set Timer", "Set time before suspend")
	timerLabels := []string{"5 minutes", "10 minutes", "15 minutes"} // define desired order
	for _, label := range timerLabels {
		timerItems[label] = timerMenu.AddSubMenuItemCheckbox(label, fmt.Sprintf("Set timer to %s", label), false)
	}
	updateTimerChecks()

	systray.AddSeparator()
	exitItem := systray.AddMenuItem("Quit", "Exit the program")

	go func() {
		for {
			select {
			case <-exitItem.ClickedCh:
				systray.Quit()
				os.Exit(0)
			}
		}
	}()

	go func() {
		for {
			<-startStopItem.ClickedCh
			monitoring = !monitoring
			if monitoring {
				startStopItem.SetTitle("Deactivate")
				systray.SetIcon(iconActive)
				go startMonitoring()
			} else {
				startStopItem.SetTitle("Activate")
				systray.SetIcon(iconInactive)
				stopMonitoring()
			}
		}
	}()

	for label, item := range timerItems {
		go func() {
			for {
				<-item.ClickedCh
				currentDuration = timerDurations[label]
				fmt.Println("Timer set to:", label)
				updateTimerChecks()
			}
		}()
	}
	startMonitoring()
}

func updateTimerChecks() {
	for label, item := range timerItems {
		if timerDurations[label] == currentDuration {
			item.Check()
		} else {
			item.Uncheck()
		}
	}
}

func onExit() {
	// Cleanup code if needed
}

func startMonitoring() {
	fmt.Println("[Monitoring started for", currentDuration, "]")
	// Placeholder: Webcam detection + overlay logic here
}

func stopMonitoring() {
	fmt.Println("[Monitoring stopped]")
	// Placeholder: stop webcam + clear overlays
}

func suspendSystem() {
	exec.Command("systemctl", "suspend").Run()
}
