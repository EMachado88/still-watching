// Still Watching - Eye inactivity monitor

package main

import (
	"fmt"
	"os"
	"os/exec"
	"sync/atomic"
	"time"

	systray "github.com/getlantern/systray"
	"gocv.io/x/gocv"
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
	eyesDetected    atomic.Value
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
	// Cleanup code if needed
}

func startMonitoring() {
	fmt.Println("[Monitoring started for", currentDuration, "]")
	eyesDetected.Store(true)

	classifier := gocv.NewCascadeClassifier()
	defer classifier.Close()
	if !classifier.Load("./assets/haarcascade_eye.xml") {
		fmt.Println("Error loading cascade file")
		return
	}

	webcam, err := gocv.OpenVideoCapture(0)
	if err != nil {
		fmt.Println("Error opening webcam:", err)
		return
	}
	defer webcam.Close()

	img := gocv.NewMat()
	defer img.Close()

	inactivityTimer := time.NewTimer(currentDuration)
	defer inactivityTimer.Stop()

	checkInterval := time.Second
detectionLoop:
	for monitoring {
		if ok := webcam.Read(&img); !ok || img.Empty() {
			continue
		}
		eyes := classifier.DetectMultiScale(img)
		if len(eyes) == 0 {
			fmt.Println("No eyes detected")
			if inactivityTimer.Stop() {
				inactivityTimer.Reset(currentDuration)
			}
			select {
			case <-inactivityTimer.C:
				fmt.Println("No eyes for", currentDuration, ", suspending...")
				suspendSystem()
				break detectionLoop
			default:
			}
		} else {
			fmt.Println("Eyes detected")
			inactivityTimer.Reset(currentDuration)
		}
		time.Sleep(checkInterval)
	}

	fmt.Println("Monitoring stopped.")
}

func stopMonitoring() {
	fmt.Println("[Monitoring stopped]")
	monitoring = false
}

func suspendSystem() {
	err := exec.Command("systemctl", "suspend").Run()
	if err != nil {
		fmt.Println("Error suspending system:", err)
	}
}
