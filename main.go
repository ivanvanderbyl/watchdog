package main

import (
	"fmt"
	"github.com/appio/watchdog/process"
	"github.com/appio/watchdog/watchdog"
	"os"
	"path/filepath"
	"strconv"
)

func main() {
	WritePid("watchdog")

	w := watchdog.New()
	p := process.NewProcess("Simulator", "/Applications/Xcode51-DP.app/Contents/Developer/Platforms/iPhoneSimulator.platform/Developer/Applications/iPhone Simulator.app/Contents/MacOS/iPhone Simulator")

	w.Add(p)
	p.Run()

	done := make(chan bool)

	go func() {
		for {
			select {
			case event := <-p.Events:
				fmt.Printf("EVENT: %s\n", event.String())
				switch event {
				case process.StartEvent:
					fmt.Printf("Process started pid=%d\n", p.PID())
				case process.StopEvent:
					fmt.Printf("Process exited status=%d\n", p.LastExitStatus)
					done <- true
				}
			}
		}
	}()

	p.Start()
	<-done
}

func WritePid(name string) {
	pidPath := "run"
	pidfile := filepath.Join(pidPath, fmt.Sprintf("%s.pid", name))

	err := os.MkdirAll(pidPath, 0750)
	if err != nil {
		panic(err)
	}

	file, err := os.Create(pidfile)
	defer file.Close()

	if err != nil {
		panic(err)
	}

	pid := os.Getpid()
	file.WriteString(strconv.FormatInt(int64(pid), 10))
}
