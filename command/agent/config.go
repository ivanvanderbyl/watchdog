package agent

import (
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/mitchellh/mapstructure"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DefaultConfig contains the defaults for configurations.
var DefaultConfig = &Config{
	LogLevel: "INFO",
	RPCAddr:  "127.0.0.1:7373",
}

type dirEnts []os.FileInfo

// Config is the configuration that can be set for an Agent. Some of these
// configurations are exposed as command-line flags to `serf agent`, whereas
// many of the more advanced configurations can only be set by creating
// a configuration file.
type Config struct {

	// ConfigDir is the directory to load process configurations from. This
	// directory will be watched for changes.
	ConfigDir string `mapstructure:"config_dir"`

	// LogLevel is the level of the logs to output.
	// This can be updated during a reload.
	LogLevel string `mapstructure:"log_level"`

	// RPCAddr is the address and port to listen on for the agent's RPC
	// interface.
	RPCAddr string `mapstructure:"rpc_addr"`

	// LogEntriesToken is used to authenticate the logging from the agent to LE.
	// This is only used for sending agent/watchdog logs, not supervised process
	// logs.
	LogEntriesToken string `mapstructure:"log_entries_token"`

	// loggingOutlet is the outlet to send logs to
	loggingOutlet io.Writer
}

// DecodeConfigFromJSON decodes a JSON config
func DecodeConfigFromJSON(r io.Reader) (*Config, error) {
	var raw interface{}
	dec := json.NewDecoder(r)
	if err := dec.Decode(&raw); err != nil {
		return nil, err
	}

	// Decode
	var md mapstructure.Metadata
	var result Config
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

// DecodeConfigFromTOML decodes a TOML config
func DecodeConfigFromTOML(r io.Reader) (*Config, error) {
	var raw interface{}

	_, err := toml.DecodeReader(r, &raw)
	if err != nil {
		return nil, err
	}

	// Decode
	var md mapstructure.Metadata
	var result Config
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

// MergeConfig merges two configurations together to make a single new
// configuration.
func MergeConfig(a, b *Config) *Config {
	var result Config = *a

	// Copy the strings if they're set
	if b.ConfigDir != "" {
		result.ConfigDir = b.ConfigDir
	}

	if b.LogEntriesToken != "" {
		result.LogEntriesToken = b.LogEntriesToken
	}

	if b.LogLevel != "" {
		result.LogLevel = b.LogLevel
	}

	if b.RPCAddr != "" {
		result.RPCAddr = b.RPCAddr
	}

	return &result
}

// ReadConfigPaths reads the paths in the given order to load configurations.
// The paths can be to files or directories. If the path is a directory,
// we read one directory deep and read any files ending in ".json" as
// configuration files.
func ReadConfigPaths(paths []string) (*Config, error) {
	result := new(Config)
	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("error reading '%s': %s", path, err)
		}

		fi, err := f.Stat()
		if err != nil {
			f.Close()
			return nil, fmt.Errorf("error reading '%s': %s", path, err)
		}

		if !fi.IsDir() {
			config, err := decodeConfigFile(path, fi)

			if err != nil {
				return nil, err
			}

			result = MergeConfig(result, config)
			continue
		}

		contents, err := f.Readdir(-1)
		f.Close()
		if err != nil {
			return nil, fmt.Errorf("error reading '%s': %s", path, err)
		}

		// Sort the contents, ensures lexical order
		sort.Sort(dirEnts(contents))

		for _, fi := range contents {
			// Don't recursively read contents
			if fi.IsDir() {
				continue
			}

			path := filepath.Join(path, fi.Name())

			config, err := decodeConfigFile(path, fi)
			if err != nil {
				return nil, err
			}

			result = MergeConfig(result, config)
		}
	}

	return result, nil
}

func decodeConfigFile(path string, fi os.FileInfo) (*Config, error) {
	result := new(Config)
	var err error

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

// Implement the sort interface for dirEnts
func (d dirEnts) Len() int {
	return len(d)
}

func (d dirEnts) Less(i, j int) bool {
	return d[i].Name() < d[j].Name()
}

func (d dirEnts) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}
