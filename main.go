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
	monitoring      = true
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
		for range exitItem.ClickedCh {
			systray.Quit()
			os.Exit(0)
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
		lbl := label
		itm := item
		go func() {
			for {
				<-itm.ClickedCh
				currentDuration = timerDurations[lbl]
				fmt.Println("Timer set to:", lbl)
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
	stopMonitoring()
}

func suspendSystem() {
	err := exec.Command("systemctl", "suspend").Run()
	if err != nil {
		fmt.Println("Error suspending system:", err)
	}
}
