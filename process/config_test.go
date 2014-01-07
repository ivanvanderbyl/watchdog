package process

import (
	"io/ioutil"
	"os"
	"testing"
)

var testConfigJSON string = `{
  "name": "my_app",
  "disabled": false,
  "user_name": "deploy",
  "group_name": "admin",
  "program": "/usr/local/bin/node",
  "program_arguments": [
    "app.js",
    "--port=8000"
  ],
  "keep_alive": true,
  "run_at_load": true,
  "working_directory": "/home/deploy/srv",
  "environment_variables": {
    "HOSTNAME": "myapp.example.com"
  },
  "kill_signal": "SIGQUIT",
  "kill_timeout": "30s",
  "throttle_interval": "15s",
  "pid_file": "/var/run/{{name}}.pid",
  "outlets": {
    "l2met": {
      "url": "https://user:pass@myl2met.herokuapp.com/collect"
    },
    "file": {
      "path": "/var/log/{{name}}.log"
    },
    "logentries": {
      "token": "BLAH"
    }
  }
}`

func setupTestConfig(t *testing.T) *os.File {
	tf, err := ioutil.TempFile("", "watchdog.json")
	defer tf.Close()
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	tf.Write([]byte(testConfigJSON))
	return tf
}

func TestConfigLoadFromJSON(t *testing.T) {
	tf := setupTestConfig(t)
	defer os.Remove(tf.Name())

	config, err := LoadConfigFile(tf.Name())
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if expected, actual := "my_app", config.Name; actual != expected {
		expect(t, expected, actual)
	}

	if expected, actual := false, config.Disabled; actual != expected {
		t.Errorf("Expected process disabled=%v, got %v", actual, expected)
	}

	if expected, actual := true, config.KeepAlive; actual != expected {
		t.Errorf("Expected keep alive=%v, got %v", actual, expected)
	}

	if expected, actual := "SIGQUIT", config.KillSignal; actual != expected {
		t.Errorf("Expected kill signal=%v, got %v", actual, expected)
	}
}
