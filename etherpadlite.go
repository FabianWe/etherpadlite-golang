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
// The API documentation can be found at https://github.com/ether/etherpad-lite/wiki/HTTP-API.
// To use it create an instance of etherpadlite.EtherpadLite and call the
// API methods on it, for example CreatePad(nil, padID, text).
// If a parameter is optional, like text is in createPad,
// simply set the value to etherpadlite.OptionalParam.
// If there is a parameter with a default value, like copyPad(sourceID, destinationID[, force=false]),
// setting the parameter to OptionalParam will set the value to the default value.
//
// All methods return a Response and an error (!= nil if something went wrong).
// The first argument of all methods is always a Context ctx. If set to a non-nil
// context the method will return nil and an error != nil when the
// context gets cancelled.
// If you don't want to use any context stuff just set it to nil.
//
// Using one etherpadlite.EtherpadLite instance from multiple goroutines is
// safe (if you're not setting any values on it).
//
// I didn't document the methods since they're documented very well on the
// etherpad-lite wiki.
package etherpadlite

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// optionalParamType is an unexported type to identify an optional parameter
// we don't want to use.
type optionalParamType int

// OptionalParam is a constant used to identify an optional parameter we don't
// want to use.
const OptionalParam optionalParamType = 0

// EtherpadLite is a struct that is used to connect to the etherpadlite API.
type EtherpadLite struct {
	// APIVersion is the api version to use, at the moment it must be set to 1.
	APIVersion string

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
	// Timeout defaults to 20 seconds in NewEtherpadLite.
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
	return &EtherpadLite{APIVersion: "1.2.10", BaseParams: baseParams,
		BaseURL: "http://localhost:9001/api", Client: client}
}

// Timeout sets Client.Timeout. Since this is something people often want to
// change I've added this wrapper.
// The default timeout is 20s.
// Of course using a context WithTimeout is nicer.
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
// If ctx != nil the method will be cancelled once ctx gets cancelled.
func (pad *EtherpadLite) sendRequest(ctx context.Context, path string, params map[string]interface{}) (*Response, error) {
	type resType struct {
		response *Response
		err      error
	}
	if ctx == nil {
		return pad.sendRequestWithoutContext(ctx, path, params)
	}
	ch := make(chan *resType, 1)
	go func() {
		resp, resErr := pad.sendRequestWithoutContext(ctx, path, params)
		ch <- &resType{response: resp, err: resErr}
	}()
	select {
	case <-ctx.Done():
		return nil, errors.New("ctx cancelled")
	case res := <-ch:
		return res.response, res.err
	}
}

// sendRequestWithoutContext does the actual stuff, it gets the request
// and ignores the ctx. sendRequest listens on done and calls this method in
// a different goroutine.
func (pad *EtherpadLite) sendRequestWithoutContext(ctx context.Context, path string, params map[string]interface{}) (*Response, error) {
	getURL, err := url.Parse(fmt.Sprintf("%s/%s/%s", pad.BaseURL, pad.APIVersion, path))
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
	if reqErr != nil {
		return nil, reqErr
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}
	resp, doErr := pad.Client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if doErr != nil {
		return nil, doErr
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

func (pad *EtherpadLite) ListAllGroups(ctx context.Context) (*Response, error) {
	return pad.sendRequest(ctx, "listAllGroups", nil)
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
	return pad.sendRequest(ctx, "listPadsOfAuthor", map[string]interface{}{"authorID": authorID})
}

func (pad *EtherpadLite) GetAuthorName(ctx context.Context, authorID interface{}) (*Response, error) {
	return pad.sendRequest(ctx, "getAuthorName", map[string]interface{}{"authorID": authorID})
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

func (pad *EtherpadLite) SetHTML(ctx context.Context, padID, html interface{}) (*Response, error) {
	return pad.sendRequest(ctx, "setHTML", map[string]interface{}{"padID": padID, "html": html})
}

func (pad *EtherpadLite) GetAttributePool(ctx context.Context, padID interface{}) (*Response, error) {
	return pad.sendRequest(ctx, "getAttributePool", map[string]interface{}{"padID": padID})
}

func (pad *EtherpadLite) GetRevisionChangeset(ctx context.Context, padID, rev interface{}) (*Response, error) {
	params := map[string]interface{}{"padID": padID}
	if rev != OptionalParam {
		params["rev"] = rev
	}
	return pad.sendRequest(ctx, "getRevisionChangeset", params)
}

func (pad *EtherpadLite) CreateDiffHTML(ctx context.Context, padID, startRev, endRev interface{}) (*Response, error) {
	return pad.sendRequest(ctx,
		"createDiffHTML",
		map[string]interface{}{
			"padID":    padID,
			"startRev": startRev,
			"endRev":   endRev,
		})
}

// Chat

func (pad *EtherpadLite) GetChatHistory(ctx context.Context, padID, start, end interface{}) (*Response, error) {
	params := map[string]interface{}{"padID": padID}
	// actually here both start and end must be != OptionalParam, not just one
	// of them. But we let the user read the docs, errors here are only for things
	// that really go wrong
	if start != OptionalParam {
		params["start"] = start
	}
	if end != OptionalParam {
		params["end"] = end
	}
	return pad.sendRequest(ctx, "getChatHistory", params)
}

func (pad *EtherpadLite) GetChatHead(ctx context.Context, padID interface{}) (*Response, error) {
	return pad.sendRequest(ctx, "getChatHead", map[string]interface{}{"padID": padID})
}

// Pad

func (pad *EtherpadLite) CreatePad(ctx context.Context, padID, text interface{}) (*Response, error) {
	params := map[string]interface{}{"padID": padID}
	if text != OptionalParam {
		params["text"] = text
	}
	return pad.sendRequest(ctx, "createPad", params)
}

func (pad *EtherpadLite) GetRevisionsCount(ctx context.Context, padID interface{}) (*Response, error) {
	return pad.sendRequest(ctx, "getRevisionsCount", map[string]interface{}{"padID": padID})
}

func (pad *EtherpadLite) GetSavedRevisionsCount(ctx context.Context, padID interface{}) (*Response, error) {
	return pad.sendRequest(ctx, "getSavedRevisionsCount", map[string]interface{}{"padID": padID})
}

func (pad *EtherpadLite) ListSavedRevisions(ctx context.Context, padID interface{}) (*Response, error) {
	return pad.sendRequest(ctx, "listSavedRevisions", map[string]interface{}{"padID": padID})
}

func (pad *EtherpadLite) SaveRevision(ctx context.Context, padID, rev interface{}) (*Response, error) {
	params := map[string]interface{}{"padID": padID}
	if rev != OptionalParam {
		params["rev"] = rev
	}
	return pad.sendRequest(ctx, "saveRevision", params)
}

func (pad *EtherpadLite) PadUsersCount(ctx context.Context, padID interface{}) (*Response, error) {
	return pad.sendRequest(ctx, "padUsersCount", map[string]interface{}{"padID": padID})
}

func (pad *EtherpadLite) PadUsers(ctx context.Context, padID interface{}) (*Response, error) {
	return pad.sendRequest(ctx, "padUsers", map[string]interface{}{"padID": padID})
}

func (pad *EtherpadLite) DeletePad(ctx context.Context, padID interface{}) (*Response, error) {
	return pad.sendRequest(ctx, "deletePad", map[string]interface{}{"padID": padID})
}

func (pad *EtherpadLite) CopyPad(ctx context.Context, sourceID, destinationID, force interface{}) (*Response, error) {
	params := map[string]interface{}{"sourceID": sourceID, "destinationID": destinationID}
	if force == OptionalParam {
		params["force"] = false
	} else {
		params["force"] = force
	}
	return pad.sendRequest(ctx, "copyPad", params)
}

func (pad *EtherpadLite) MovePad(ctx context.Context, sourceID, destinationID, force interface{}) (*Response, error) {
	params := map[string]interface{}{"sourceID": sourceID, "destinationID": destinationID}
	if force == OptionalParam {
		params["force"] = false
	} else {
		params["force"] = force
	}
	return pad.sendRequest(ctx, "movePad", params)
}

func (pad *EtherpadLite) GetReadOnlyID(ctx context.Context, padID interface{}) (*Response, error) {
	return pad.sendRequest(ctx, "getReadOnlyID", map[string]interface{}{"padID": padID})
}

func (pad *EtherpadLite) GetPadID(ctx context.Context, readOnlyID interface{}) (*Response, error) {
	return pad.sendRequest(ctx, "getPadID", map[string]interface{}{"readOnlyID": readOnlyID})
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

func (pad *EtherpadLite) SendClientsMessage(ctx context.Context, padID, msg interface{}) (*Response, error) {
	return pad.sendRequest(ctx, "sendClientsMessage", map[string]interface{}{"padID": padID, "msg": msg})
}

func (pad *EtherpadLite) CheckToken(ctx context.Context) (*Response, error) {
	return pad.sendRequest(ctx, "checkToken", nil)
}

// Pads

func (pad *EtherpadLite) ListAllPads(ctx context.Context) (*Response, error) {
	return pad.sendRequest(ctx, "listAllPads", nil)
}
