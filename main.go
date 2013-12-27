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
	RealMain()
}

func RealMain() {
	WritePid("watchdog-sim")

	w := watchdog.New()
	p := process.NewProcess("Simulator", "/Applications/Xcode51-DP.app/Contents/Developer/Platforms/iPhoneSimulator.platform/Developer/Applications/iPhone Simulator.app/Contents/MacOS/iPhone Simulator")

	fmt.Println("Adding process")
	w.Add(p)

	fmt.Println("Start process")
	p.Run()
	p.Start()

	fmt.Printf("Process started pid=%d\n", p.PID())

	p.Wait()

	fmt.Printf("Process exited status=%d\n", p.LastExitStatus)
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
