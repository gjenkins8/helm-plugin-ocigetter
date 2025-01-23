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
	"io"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	pdk "github.com/extism/go-pdk"
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
	roundTripper   http.RoundTripper
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
func NewOCIGetter(options GetterOptions, roundTripper http.RoundTripper) (*OCIGetter, error) {
	var client OCIGetter

	client.opts = options
	client.roundTripper = roundTripper
	return &client, nil
}

func (g *OCIGetter) newRegistryClient() (*registry.Client, error) {

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
		Transport: g.roundTripper,
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

type GetterPluginOutput struct {
	ChartData *bytes.Buffer `json:"chart_data"`
}

func runOciGetter(input GetterPluginInput, roundTripper http.RoundTripper) (*GetterPluginOutput, error) {

	getter, err := NewOCIGetter(input.Options, roundTripper)
	if err != nil {
		return nil, fmt.Errorf("new oci getter failed: %q\n", err)
	}

	chartData, err := getter.Get(input.HRef)
	if err != nil {
		return nil, fmt.Errorf("get failed: %q\n", err)
	}

	result := GetterPluginOutput{
		ChartData: chartData,
	}

	return &result, nil
}

type ExtismRoundTripper struct {
}

func (e *ExtismRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {

	getHTTPMethod := func(httpMethod string) (pdk.HTTPMethod, error) {
		switch httpMethod {
		case "GET":
			return pdk.MethodGet, nil
		case "HEAD":
			return pdk.MethodHead, nil
		}

		return pdk.HTTPMethod(0), fmt.Errorf("unknown method: %s", req.Method)
	}

	httpMethod, err := getHTTPMethod(req.Method)
	if err != nil {
		return nil, err
	}

	pdkReq := pdk.NewHTTPRequest(httpMethod, req.URL.String())
	for name, values := range req.Header {
		for _, value := range values {
			pdkReq.SetHeader(name, value)
		}
	}

	switch req.Method {
	case "PUT", "POST":
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %q", err)
		}
		pdkReq.SetBody(body)
	}

	pdkResp := pdkReq.Send()

	pdkRespBody := pdkResp.Body()
	resp := http.Response{
		Status:     "unknown",
		StatusCode: int(pdkResp.Status()),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header: func() http.Header {
			result := http.Header{}
			for name, value := range pdkResp.Headers() {
				result.Set(name, value)
			}

			return result
		}(),
		Body:          io.NopCloser(bytes.NewReader(pdkRespBody)),
		ContentLength: int64(len(pdkRespBody)),
		Request:       req,
	}
	return &resp, nil
}

//go:export pluginhelmgetter
func PluginHelmGetter() uint32 {
	pdk.Log(pdk.LogInfo, "PluginRunGetter")

	input := GetterPluginInput{}
	if err := pdk.InputJSON(&input); err != nil {
		pdk.SetError(fmt.Errorf("failed to decode input: %q", err))
		return 1
	}

	pdk.Log(pdk.LogDebug, fmt.Sprintf("input: %+v", input))
	result, err := runOciGetter(input, &ExtismRoundTripper{})
	if err != nil {
		pdk.SetError(fmt.Errorf("error fetching chart: %s %q", input.HRef, err))
		return 2
	}

	if err := pdk.OutputJSON(result); err != nil {
		pdk.SetError(fmt.Errorf("failed to encode output: %q", err))
		return 1
	}

	return 0
}

func main() {
	fmt.Printf("main\n")

	//input := GetterPluginInput
	//runOciGetter(
}
