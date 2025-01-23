package main_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	extism "github.com/extism/go-sdk"
	"github.com/stretchr/testify/require"
)

var input []byte = []byte(`{
    "options": {
	},
    "href": "oci://ghcr.io/stefanprodan/charts/podinfo:6.7.1"
}`)

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
		EnableWasi: true,
	}
	plugin, err := extism.NewPlugin(ctx, manifest, config, []extism.HostFunction{})

	if err != nil {
		fmt.Printf("Failed to initialize plugin: %v\n", err)
		os.Exit(1)
	}

	exit, out, err := plugin.Call("pluginhelmgetter", input)
	if err != nil {
		fmt.Println(err)
		os.Exit(int(exit))
	}

	fmt.Printf("%s", out)

	response := string(out)
	fmt.Println(response)
}
