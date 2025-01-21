/*
Copyright The Helm Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/extism/go-pdk"
	"github.com/gjenkins8/ocigetterplugin/registry"
)

type GetterOptions struct {
	URL *string `json:"url"`
	//CertFile              *string       `json:"cert_file"`
	//KeyFile               *string       `json:"key_file"`
	CAFile                *string       `json:"ca_file"`
	UnTar                 *bool         `json:"untar"`
	InsecureSkipVerifyTLS *bool         `json:"insecure_skip_verify_tls"`
	PlainHTTP             bool          `json:"plain_http"`
	AcceptHeader          *string       `json:"accept_header"`
	Username              *string       `json:"username"`
	Password              *string       `json:"password"`
	PassCredentialsAll    *bool         `json:"pass_credentials_all"`
	UserAgent             *string       `json:"user_agent"`
	Version               *string       `json:"version"`
	Timeout               time.Duration `json:"timeout"`
	//Transport             *http.Transport
}

// OCIGetter is the default HTTP(/S) backend handler
type OCIGetter struct {
	registryClient *registry.Client
	opts           GetterOptions
	transport      *http.Transport
	once           sync.Once
}

// Get performs a Get from repo.Getter and returns the body.
func (g *OCIGetter) Get(href string) (*bytes.Buffer, error) {
	return g.get(href)
}

func (g *OCIGetter) get(href string) (*bytes.Buffer, error) {
	client := g.registryClient
	// if the user has already provided a configured registry client, use it,
	// this is particularly true when user has his own way of handling the client credentials.
	if client == nil {
		c, err := g.newRegistryClient()
		if err != nil {
			return nil, err
		}
		client = c
	}

	ref := strings.TrimPrefix(href, fmt.Sprintf("%s://", registry.OCIScheme))

	if version := g.opts.Version; version != nil && !strings.Contains(path.Base(ref), ":") {
		ref = fmt.Sprintf("%s:%s", ref, version)
	}
	var pullOpts []registry.PullOption
	requestingProv := strings.HasSuffix(ref, ".prov")
	if requestingProv {
		ref = strings.TrimSuffix(ref, ".prov")
		pullOpts = append(pullOpts,
			registry.PullOptWithChart(false),
			registry.PullOptWithProv(true))
	}

	result, err := client.Pull(ref, pullOpts...)
	if err != nil {
		return nil, err
	}

	if requestingProv {
		return bytes.NewBuffer(result.Prov.Data), nil
	}
	return bytes.NewBuffer(result.Chart.Data), nil
}

// NewOCIGetter constructs a valid http/https client as a Getter
func NewOCIGetter(options GetterOptions) (*OCIGetter, error) {
	var client OCIGetter

	client.opts = options
	return &client, nil
}

func (g *OCIGetter) newRegistryClient() (*registry.Client, error) {

	if g.transport != nil {
		client, err := registry.NewClient(
			registry.ClientOptHTTPClient(&http.Client{
				Transport: g.transport,
				Timeout:   g.opts.Timeout,
			}),
		)
		if err != nil {
			return nil, err
		}
		return client, nil
	}

	g.once.Do(func() {
		g.transport = &http.Transport{
			// From https://github.com/google/go-containerregistry/blob/31786c6cbb82d6ec4fb8eb79cd9387905130534e/pkg/v1/remote/options.go#L87
			//DisableCompression: true,
			//DialContext: (&net.Dialer{
			//	// By default we wrap the transport in retries, so reduce the
			//	// default dial timeout to 5s to avoid 5x 30s of connection
			//	// timeouts when doing the "ping" on certain http registries.
			//	Timeout:   5 * time.Second,
			//	KeepAlive: 30 * time.Second,
			//}).DialContext,
			//ForceAttemptHTTP2:     true,
			//MaxIdleConns:          100,
			//IdleConnTimeout:       90 * time.Second,
			//TLSHandshakeTimeout:   10 * time.Second,
			//ExpectContinueTimeout: 1 * time.Second,
			//Proxy:                 http.ProxyFromEnvironment,
		}
	})

	//if (g.opts.CertFile != "" && g.opts.KeyFile != nil) || g.opts.CAFile != nil || g.opts.InsecureSkipVerifyTLS {
	//	tlsConf, err := tlsutil.NewClientTLS(g.opts.CertFile, g.opts.KeyFile, g.opts.CAFile, g.opts.InsecureSkipVerifyTLS)
	//	if err != nil {
	//		return nil, fmt.Errorf("can't create TLS config for client: %w", err)
	//	}

	//	sni, err := ExtractHostname(g.opts.URL)
	//	if err != nil {
	//		return nil, err
	//	}
	//	tlsConf.ServerName = sni

	//	//g.transport.TLSClientConfig = tlsConf
	//}

	opts := []registry.ClientOption{registry.ClientOptHTTPClient(&http.Client{
		Transport: g.transport,
		Timeout:   g.opts.Timeout,
	})}
	if g.opts.PlainHTTP {
		opts = append(opts, registry.ClientOptPlainHTTP())
	}

	client, err := registry.NewClient(opts...)

	if err != nil {
		return nil, err
	}

	return client, nil
}

type GetterPluginInput struct {
	Options GetterOptions `json:"options"`
	HRef    string        `json:"href"`
}

//go:export greet
func Greet() uint32 {
	fmt.Printf("greet\n")
	fmt.Printf("greet\n")
	fmt.Printf("greet\n")

	fmt.Printf("env[HOME]: %q\n", os.Getenv("HOME"))

	input := GetterPluginInput{}
	if err := pdk.InputJSON(&input); err != nil {
		fmt.Printf("failed to decode input: %q\n", err)
		return 1
	}

	fmt.Printf("input: %+v\n", input)

	getter, err := NewOCIGetter(input.Options)
	if err != nil {
		fmt.Printf("new oci getter failed: %q\n", err)
		return 2
	}

	if _, err := getter.Get(input.HRef); err != nil {
		fmt.Printf("get failed: %q\n", err)
		return 2

	}
	return 0
}

func main() {
	fmt.Printf("main\n")
}
