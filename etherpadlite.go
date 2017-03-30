// Copyright 2017 Fabian Wenzelmann
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// etherpadlite provides an interface for Etherpad-Lite's HTTP API written in Go.
// The API documentation can be found at <https://github.com/ether/etherpad-lite/wiki/HTTP-API>.
// To use it create an instance of etherpadlite.EtherpadLite and call the
// API methods on it, for example CreateAuthorIfNotExistsFor(authorMapper, name, nil).
// If a parameter is optional, like name is in createAuthorIfNotExistsFor,
// simply set the value to nil.
// All methods return a Response and an error (!= if something went wrong).
// The last argument of all methods is always a context ctx. If set to a not nil
// context the requests are created with the given context and will stop working
// once ctx gets cancelled. If you don't want to use any context stuff just
// set it to nil and it should be fine.
//
// I didn't document the methods since they're documented very well on the
// etherpad-lite wiki.
package etherpadlite

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// EtherpadLite is a struct that is used to connect to the etherpadlite API.
type EtherpadLite struct {
	// APIVersion is the api version to use, at the moment it must be set to 1.
	APIVersion int

	// BaseParams are the parameter that are required in each request, i.e.
	// will be sent in each request.
	// The entry BaseParams["apikey"] = YOUR-API-KEY
	// should always be present.
	// NewEtherpadLite will take care of this.
	BaseParams map[string]interface{}

	// BaseURL is the URL pointing to the API of your pad, i.e.
	// http://pad.domain/api
	// It defaults to http://localhost:9001/api in NewEtherpadLite.
	BaseURL string

	// Client is used to send the GET requests to the API.
	// Set the values as required, especially timeout is important.
	// That's why there is a special Timeout method to set the timeout.
	// Timeout defaults to 20 seconds.
	Client *http.Client
}

// NewEtherpadLite creates a new EtherpadLite instance given the
// mandatory apiKey.
// Create a new instance with this method and then configure it if you must.
func NewEtherpadLite(apiKey string) *EtherpadLite {
	baseParams := make(map[string]interface{})
	baseParams["apikey"] = apiKey
	client := &http.Client{}
	client.Timeout = time.Duration(20 * time.Second)
	return &EtherpadLite{APIVersion: 1, BaseParams: baseParams,
		BaseURL: "http://localhost:9001/api", Client: client}
}

// Timeout sets Client.Timeout. Since this is something people often want to
// change I've added this wrapper.
// The default timeout is 20s.
func (pad *EtherpadLite) Timeout(timeout time.Duration) {
	pad.Client.Timeout = timeout
}

// Response is the response from the etherpad API.
// See https://github.com/ether/etherpad-lite/wiki/HTTP-API
type Response struct {
	Code    int
	Message string
	Data    map[string]interface{}
}

// sendRequest is the function doing most of the work by sending the real
// request. It will encode the BaseParams and params into URL queries and
// do the http GET.
// It decodes the JSON result and returns the decoded version.
// If ctx != nil the request gets cancelled when the context gets cancelled.
func (pad *EtherpadLite) sendRequest(path string, params map[string]interface{}, ctx context.Context) (*Response, error) {
	getURL, err := url.Parse(fmt.Sprintf("%s/%d/%s", pad.BaseURL, pad.APIVersion, path))
	if err != nil {
		return nil, err
	}
	parameters := url.Values{}
	for key, value := range pad.BaseParams {
		parameters.Add(key, fmt.Sprintf("%v", value))
	}
	for key, value := range params {
		parameters.Add(key, fmt.Sprintf("%v", value))
	}
	getURL.RawQuery = parameters.Encode()
	req, reqErr := http.NewRequest("GET", getURL.String(), nil)
	if ctx != nil {
		req = req.WithContext(ctx)
	}
	if reqErr != nil {
		return nil, reqErr
	}
	resp, getErr := pad.Client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if getErr != nil {
		return nil, getErr
	}
	allContent, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}
	var padResponse Response
	jsonErr := json.Unmarshal(allContent, &padResponse)
	if jsonErr != nil {
		return nil, jsonErr
	}
	return &padResponse, nil
}

// Groups

func (pad *EtherpadLite) CreateGroup(ctx context.Context) (*Response, error) {
	return pad.sendRequest("createGroup", nil, ctx)
}

func (pad *EtherpadLite) CreateGroupIfNotExistsFor(groupMapper interface{}, ctx context.Context) (*Response, error) {
	return pad.sendRequest("createGroupIfNotExistsFor", map[string]interface{}{"groupMapper": groupMapper}, ctx)
}

func (pad *EtherpadLite) DeleteGroup(groupID interface{}, ctx context.Context) (*Response, error) {
	return pad.sendRequest("deleteGroup", map[string]interface{}{"groupID": groupID}, ctx)
}

func (pad *EtherpadLite) ListPads(groupID interface{}, ctx context.Context) (*Response, error) {
	return pad.sendRequest("listPads", map[string]interface{}{"groupID": groupID}, ctx)
}

func (pad *EtherpadLite) CreateGroupPad(groupID, padName, text interface{}, ctx context.Context) (*Response, error) {
	params := map[string]interface{}{"groupID": groupID, "padName": padName}
	if text != nil {
		params["text"] = text
	}
	return pad.sendRequest("createGroupPad", params, ctx)
}

// Author

func (pad *EtherpadLite) CreateAuthor(name interface{}, ctx context.Context) (*Response, error) {
	params := make(map[string]interface{})
	if name != nil {
		params["name"] = name
	}
	return pad.sendRequest("createAuthor", params, ctx)
}

func (pad *EtherpadLite) CreateAuthorIfNotExistsFor(authorMapper, name interface{}, ctx context.Context) (*Response, error) {
	params := map[string]interface{}{"authorMapper": authorMapper}
	if name != nil {
		params["name"] = name
	}
	return pad.sendRequest("createAuthorIfNotExistsFor", params, ctx)
}

func (pad *EtherpadLite) ListPadsOfAuthor(authorID interface{}, ctx context.Context) (*Response, error) {
	return pad.sendRequest("listPadsOfAuthor", nil, ctx)
}

// Session
