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
// API methods on it, for example CreatePad(nil, padID, text).
// If a parameter is optional, like text is in createPad,
// simply set the value to etherpadlite.OptionalParam.
//
// All methods return a Response and an error (!= nil if something went wrong).
// The first argument of all methods is always a Context ctx. If set to a non-nil
// context the requests are created with the given context and are cancelled
// once ctx gets cancelled. If you don't want to use any context stuff just
// set it to nil.
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

type optionalParamType int

const OptionalParam optionalParamType = 0

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
func (pad *EtherpadLite) sendRequest(ctx context.Context, path string, params map[string]interface{}) (*Response, error) {
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
	fmt.Println(getURL)
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
	return pad.sendRequest(ctx, "createGroup", nil)
}

func (pad *EtherpadLite) CreateGroupIfNotExistsFor(ctx context.Context, groupMapper interface{}) (*Response, error) {
	return pad.sendRequest(ctx, "createGroupIfNotExistsFor", map[string]interface{}{"groupMapper": groupMapper})
}

func (pad *EtherpadLite) DeleteGroup(ctx context.Context, groupID interface{}) (*Response, error) {
	return pad.sendRequest(ctx, "deleteGroup", map[string]interface{}{"groupID": groupID})
}

func (pad *EtherpadLite) ListPads(ctx context.Context, groupID interface{}) (*Response, error) {
	return pad.sendRequest(ctx, "listPads", map[string]interface{}{"groupID": groupID})
}

func (pad *EtherpadLite) CreateGroupPad(ctx context.Context, groupID, padName, text interface{}) (*Response, error) {
	params := map[string]interface{}{"groupID": groupID, "padName": padName}
	if text != OptionalParam {
		params["text"] = text
	}
	return pad.sendRequest(ctx, "createGroupPad", params)
}

// Author

func (pad *EtherpadLite) CreateAuthor(ctx context.Context, name interface{}) (*Response, error) {
	params := make(map[string]interface{})
	if name != OptionalParam {
		params["name"] = name
	}
	return pad.sendRequest(ctx, "createAuthor", params)
}

func (pad *EtherpadLite) CreateAuthorIfNotExistsFor(ctx context.Context, authorMapper, name interface{}) (*Response, error) {
	params := map[string]interface{}{"authorMapper": authorMapper}
	if name != OptionalParam {
		params["name"] = name
	}
	return pad.sendRequest(ctx, "createAuthorIfNotExistsFor", params)
}

func (pad *EtherpadLite) ListPadsOfAuthor(ctx context.Context, authorID interface{}) (*Response, error) {
	return pad.sendRequest(ctx, "listPadsOfAuthor", nil)
}

// Session

func (pad *EtherpadLite) CreateSession(ctx context.Context, groupID, authorID, validUntil interface{}) (*Response, error) {
	return pad.sendRequest(ctx,
		"createSession",
		map[string]interface{}{"groupID": groupID, "authorID": authorID, "validUntil": validUntil})
}

func (pad *EtherpadLite) DeleteSession(ctx context.Context, sessionID interface{}) (*Response, error) {
	return pad.sendRequest(ctx, "deleteSession", map[string]interface{}{"sessionID": sessionID})
}

func (pad *EtherpadLite) GetSessionInfo(ctx context.Context, sessionID interface{}) (*Response, error) {
	return pad.sendRequest(ctx, "getSessionInfo", map[string]interface{}{"sessionID": sessionID})
}

func (pad *EtherpadLite) ListSessionsOfGroup(ctx context.Context, groupID interface{}) (*Response, error) {
	return pad.sendRequest(ctx, "listSessionsOfGroup", map[string]interface{}{"groupID": groupID})
}

func (pad *EtherpadLite) ListSessionsOfAuthor(ctx context.Context, authorID interface{}) (*Response, error) {
	return pad.sendRequest(ctx, "listSessionsOfAuthor", map[string]interface{}{"authorID": authorID})
}

// Pad Content

func (pad *EtherpadLite) GetText(ctx context.Context, padID, rev interface{}) (*Response, error) {
	params := map[string]interface{}{"padID": padID}
	if rev != OptionalParam {
		params["rev"] = rev
	}
	return pad.sendRequest(ctx, "getText", params)
}

func (pad *EtherpadLite) SetText(ctx context.Context, padID, text interface{}) (*Response, error) {
	return pad.sendRequest(ctx, "setText", map[string]interface{}{"padID": padID, "text": text})
}

func (pad *EtherpadLite) GetHTML(ctx context.Context, padID, rev interface{}) (*Response, error) {
	params := map[string]interface{}{"padID": padID}
	if rev != OptionalParam {
		params["rev"] = rev
	}
	return pad.sendRequest(ctx, "getHTML", params)
}

// Pad

func (pad *EtherpadLite) CreatePad(ctx context.Context, padID, text interface{}) (*Response, error) {
	params := map[string]interface{}{"padID": padID}
	if text != OptionalParam {
		params["text"] = text
		fmt.Println("Setting text!")
	} else {
		fmt.Println("Not setting :()")
	}
	return pad.sendRequest(ctx, "createPad", params)
}

func (pad *EtherpadLite) GetRevisionsCount(ctx context.Context, padID interface{}) (*Response, error) {
	return pad.sendRequest(ctx, "getRevisionsCount", map[string]interface{}{"padID": padID})
}

func (pad *EtherpadLite) PadUsersCount(ctx context.Context, padID interface{}) (*Response, error) {
	return pad.sendRequest(ctx, "padUsersCount", map[string]interface{}{"padID": padID})
}

func (pad *EtherpadLite) DeletePad(ctx context.Context, padID interface{}) (*Response, error) {
	return pad.sendRequest(ctx, "deletePad", map[string]interface{}{"padID": padID})
}

func (pad *EtherpadLite) GetReadOnlyID(ctx context.Context, padID interface{}) (*Response, error) {
	return pad.sendRequest(ctx, "getReadOnlyID", map[string]interface{}{"padID": padID})
}

func (pad *EtherpadLite) SetPublicStatus(ctx context.Context, padID, publicStatus interface{}) (*Response, error) {
	return pad.sendRequest(ctx, "setPublicStatus", map[string]interface{}{"padID": padID, "publicStatus": publicStatus})
}

func (pad *EtherpadLite) GetPublicStatus(ctx context.Context, padID interface{}) (*Response, error) {
	return pad.sendRequest(ctx, "getPublicStatus", map[string]interface{}{"padID": padID})
}

func (pad *EtherpadLite) SetPassword(ctx context.Context, padID, password interface{}) (*Response, error) {
	return pad.sendRequest(ctx, "setPassword", map[string]interface{}{"padID": padID, "password": password})
}

func (pad *EtherpadLite) IsPasswordProtected(ctx context.Context, padID interface{}) (*Response, error) {
	return pad.sendRequest(ctx, "isPasswordProtected", map[string]interface{}{"padID": padID})
}

func (pad *EtherpadLite) ListAuthorsOfPad(ctx context.Context, padID interface{}) (*Response, error) {
	return pad.sendRequest(ctx, "listAuthorsOfPad", map[string]interface{}{"padID": padID})
}

func (pad *EtherpadLite) GetLastEdited(ctx context.Context, padID interface{}) (*Response, error) {
	return pad.sendRequest(ctx, "getLastEdited", map[string]interface{}{"padID": padID})
}
