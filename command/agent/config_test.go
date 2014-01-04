package agent

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func TestDecodeConfig(t *testing.T) {
	// Without a protocol
	input := `{"rpc_addr": "localhost:1789"}`
	config, err := DecodeConfigFromJSON(bytes.NewReader([]byte(input)))
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if config.RPCAddr != "localhost:1789" {
		t.Fatalf("bad: %#v", config)
	}
}

func TestReadConfigPaths_JSON_file(t *testing.T) {
	tf, err := ioutil.TempFile("", "watchdog.json")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	tf.Write([]byte(`{"rpc_addr":"localhost:1234"}`))
	tf.Close()
	defer os.Remove(tf.Name())

	config, err := ReadConfigPaths([]string{tf.Name()})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if config.RPCAddr != "localhost:1234" {
		t.Fatalf("bad: %#v", config)
	}
}

func TestReadConfigPaths_TOML_file(t *testing.T) {
	tf, err := ioutil.TempFile("", "watchdog.toml")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	tf.Write([]byte(`rpc_addr = "localhost:1234"`))
	tf.WriteString("\n")
	tf.Close()
	defer os.Remove(tf.Name())

	config, err := ReadConfigPaths([]string{tf.Name()})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if config.RPCAddr != "localhost:1234" {
		t.Fatalf("bad: %#v", config)
	}
}
