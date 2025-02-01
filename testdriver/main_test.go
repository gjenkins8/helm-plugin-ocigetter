package main_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	extism "github.com/extism/go-sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var input []byte = []byte(`{
    "options": {
	},
    "href": "oci://ghcr.io/stefanprodan/charts/podinfo:6.7.1"
}`)

type GetterPluginOutput struct {
	ChartData []byte `json:"chart_data"`
}

func TestPullOCI(t *testing.T) {
	ociGetterPluginBytes, err := os.ReadFile("../ocigetterplugin.wasm")
	require.Nil(t, err)

	manifest := extism.Manifest{
		Wasm: []extism.Wasm{
			extism.WasmData{
				Data: ociGetterPluginBytes,
				Name: "ocigetterplugin",
			},
		},
		Memory: &extism.ManifestMemory{
			MaxPages:             65535,
			MaxHttpResponseBytes: 1024 * 1024 * 10,
			MaxVarBytes:          1024 * 1024 * 10,
		},
		Config:       map[string]string{},
		AllowedHosts: []string{"ghcr.io"},
		AllowedPaths: map[string]string{},
		Timeout:      0,
	}

	ctx := context.Background()
	config := extism.PluginConfig{
		EnableWasi:                true,
		EnableHttpResponseHeaders: true,
	}
	plugin, err := extism.NewPlugin(ctx, manifest, config, []extism.HostFunction{})
	plugin.SetLogger(func(logLevel extism.LogLevel, s string) {
		fmt.Println(s)
	})

	if err != nil {
		fmt.Printf("Failed to initialize plugin: %v\n", err)
		os.Exit(1)
	}

	exitCode, outputData, err := plugin.Call("_start", input)
	require.Nil(t, err)
	assert.Equal(t, uint32(0), exitCode)

	output := GetterPluginOutput{}
	if err := json.Unmarshal(outputData, &output); err != nil {
		assert.Nil(t, err)
	}

	h := sha256.New()
	h.Write(output.ChartData)
	assert.Equal(t, "23b693db774406415a07ebf5a0163932260701958d3eaf1b3a6c9945daf8dd81", hex.EncodeToString(h.Sum(nil)))
}
