package main

import (
	"fmt"
	"time"

	"gocv.io/x/gocv"
)

func startMonitoring() {
	fmt.Println("[Monitoring started for", currentDuration, "]")

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
	for monitoring {
		if ok := webcam.Read(&img); !ok || img.Empty() {
			continue
		}
		eyes := classifier.DetectMultiScale(img)

		if len(eyes) == 0 {
			fmt.Println("No eyes detected")
			select {
			case <-inactivityTimer.C:
				fmt.Println("No eyes for", currentDuration, ", suspending...")
				suspendSystem()
			default:
			}
		} else {
			fmt.Println("Eyes detected")
			inactivityTimer.Reset(currentDuration)
		}

		time.Sleep(checkInterval)
	}
}

func stopMonitoring() {
	fmt.Println("[Monitoring stopped]")
	monitoring = false
}
