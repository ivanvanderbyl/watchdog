package process

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	// "fmt"
	// "github.com/BurntSushi/toml"
	"github.com/mitchellh/mapstructure"
	"io"
	// "os"
	// "path/filepath"
	// "sort"
	// "strings"
)

// ProcessConfig provides methods for loading and parsing configuration files
// into a struct which can be loaded as a process.
//
// All exported fields in this struct are supported in a process configuration
// file.
type ProcessConfig struct {
	// Name is a required key that uniquely identifies the process to Watchdog.
	Name string `mapstructure:"name"`

	// Disables is an optional key that instructs Watchdog on whether to load and
	// run this process. If this is set to false, watchdog will stop the process
	// and not respond to commands to start it. This configuration file will still
	// be monitored for changes in the event it later gets set to true, which will
	// in turn launch the process if it is configured to do so.
	Disabled bool `mapstructure:"disabled"`

	// UserName is an optional key specifies the user to run the job as.
	// This key is only applicable when watchdog is running as root.
	UserName string `mapstructure:"user_name"`

	// GroupName is an optional key specifies the group to run the process as.
	// This key is only applicable when watchdog is running as root.
	//
	// If UserName is set and GroupName is not, the the group will be set to
	// the default group of the user
	GroupName string `mapstructure:"group_name"`

	// Program key maps to the first argument of exec.Command().  If this key is
	// missing, then the first element of the array of strings provided to the
	// ProgramArguments will be used instead.
	//
	// This key is required in the absence of the ProgramArguments key.
	Program string `mapstructure:"program"`

	// ProgramArguments key maps to the second and subsequent arguments of
	// exec.Command. This key is required in the absence of the Program key.
	ProgramArguments []string `mapstructure:"program_arguments"`

	// KeepAlive key is used to control whether Watchdog should restart this
	// process in the event that is exits early. If this key is set to false,
	// Watchdog will not restart the process unless you manually tell it to do so.
	KeepAlive bool `mapstructure:"keep_alive"`

	// RunAtLoad key is used to control whether your process is launched once
	// at the time the config is loaded. The default is true.
	RunAtLoad bool `mapstructure:"run_at_load"`

	// WorkingDirectory is an optional key that is used to specify a directory
	// to chdir(2) to before running the process.
	WorkingDirectory string `mapstructure:"working_directory"`

	// EnvironmentVariables key is used to specify additional environmental
	// variables to be set before running the process.
	EnvironmentVariables map[string]string `mapstructure:"environment_variables"`

	// KillSignal is used to specify which os.Signal to send to the process to
	// instruct it to exit gracefully. Default is SIGKILL.
	KillSignal string `mapstructure:"kill_signal"`

	// KillTimeout is used to specify the amount of time to wait for the process
	// to safely exit after sending KillSignal before sending a SIGTERM. The
	// default is 10s.
	KillTimeout string `mapstructure:"kill_timeout"`

	// ThrottleInterval specifies the amount of time to wait before respawning the
	// process after it exits, if it is set to KeepAlive. The default value is
	// 10s.
	ThrottleInterval string `mapstructure:"throttle_interval"`

	// PidFile specifies where to write a pid file for this process when it is
	// spawned. An empty value disables this function.
	PidFile string

	// Outlets specifies a an outlet type by key and value of configuration to
	// be passed to that outlet when being constructed.
	Outlets map[string]map[string]string
}

// LoadConfigFile loads a process configuration from a file on disk.
func LoadConfigFile(path string) (*ProcessConfig, error) {
	return decodeConfigFile(path)
}

// DecodeConfigFromProcfile will construct a minimal process configuration from
// a Foreman Procfile, starting it immediately and keeping it alive.
func DecodeConfigFromProcfile(r io.Reader) (*ProcessConfig, error) {
	return new(ProcessConfig), nil
}

// DecodeConfigFromJSON loads a ProcessConfig from a JSON file
func DecodeConfigFromJSON(r io.Reader) (*ProcessConfig, error) {
	var raw interface{}
	dec := json.NewDecoder(r)
	if err := dec.Decode(&raw); err != nil {
		return nil, err
	}

	// Decode
	var md mapstructure.Metadata
	var result ProcessConfig
	msdec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata: &md,
		Result:   &result,
	})
	if err != nil {
		return nil, err
	}

	if err := msdec.Decode(raw); err != nil {
		return nil, err
	}

	return &result, nil
}

// DecodeConfigFromTOML loads a ProcessConfig from a TOML file
func DecodeConfigFromTOML(r io.Reader) (*ProcessConfig, error) {
	return new(ProcessConfig), nil
}

func decodeConfigFile(path string) (*ProcessConfig, error) {
	var fi os.FileInfo
	var err error

	fi, err = os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("could't stat '%s': %s", path, err)
	}

	result := new(ProcessConfig)

	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return nil, fmt.Errorf("error reading '%s': %s", path, err)
	}

	// If it isn't a JSON file, ignore it
	if strings.HasSuffix(fi.Name(), ".json") {
		result, err = DecodeConfigFromJSON(f)

		if err != nil {
			return nil, fmt.Errorf("error decoding '%s': %s", path, err)
		}
	} else if strings.HasSuffix(fi.Name(), ".toml") {
		result, err = DecodeConfigFromTOML(f)

		if err != nil {
			return nil, fmt.Errorf("error decoding '%s': %s", path, err)
		}
	} else {

		// Try decoding with json then toml, else fail
		result, err = DecodeConfigFromJSON(f)
		if err != nil {
			f.Close()
			f, err = os.Open(path)
			defer f.Close()
			if err != nil {
				return nil, fmt.Errorf("error reading '%s': %s", path, err)
			}

			result, err = DecodeConfigFromTOML(f)
			if err != nil {
				return nil, fmt.Errorf("error decoding '%s': unknown format", path)
			}
		}
		f.Close()
	}

	return result, nil
}
