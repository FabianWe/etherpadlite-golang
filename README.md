# etherpadlite-golang
An interface for [Etherpad-Lite's HTTP API](https://github.com/ether/etherpad-lite/wiki/HTTP-API) written in Go.

## Installation
Run `go get FabianWe/etherpadlite-golang`
Read the code documentation on [GoDoc](https://godoc.org/github.com/FabianWe/etherpadlite-golang).
## Supported API Versions
Though I haven not tested each and every method I'm very confident that all versions including version 1.2.13 are supported. Feedback is very welcome!

## Usage
Here's a very simple example that should give you the idean:
```go
import (
	"fmt"
	"log"

	etherpadlite "github.com/FabianWe/etherpadlite-golang"
)

func main() {
	pad := etherpadlite.NewEtherpadLite("your-api-key")
	response, err := pad.CreatePad(nil, "foo", "Lorem ipsum dolor sit amet.")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(response.Message, response.Data)
}
```
All methods return two values: A [Response](https://godoc.org/github.com/FabianWe/etherpadlite-golang#Response) containing the parsed JSON response and an `error`. If `err != nil` something went really wrong, for example the connection to the host failed.

You can configure the [EtherpadLite](https://godoc.org/github.com/FabianWe/etherpadlite-golang#EtherpadLite), for example configure the [http.Client](https://golang.org/pkg/net/http/#Client).

An `EtherpadLite` instance has the following fields:

 - APIVersion: The HTTP API version. Defaults to 1.2.13. Note that this is a rather new version, if you have an older version of etherpad-lite you may have to adjust this!
 - BaseParams: A map that contains the parameters that are sent in every request. The API key gets added in `NewEtherpadLite`.
 - BaseURL: The URL pointing to the API of your pad, i.e. http://pad.domain/api. Defaults to http://localhost:9001/api in `NewEtherpadLite`.
 - Client: The client used to send the GET requests. The default Timeout for the Client is 20 seconds.

All functions take as first argument a [context.Context](https://golang.org/pkg/context/#Context).  If you pass `ctx != nil` the methods will get cancelled when `ctx` gets cancelled (i.e. return no Response and an error != nil). If you don't want to use this simply set it to `nil`.

If a method has an optional field, for example `text` in `CreatePad`, set the value to `etherpadlite.OptionalParam` if you don't want to use it. So to create a pad without text do:
```go
response, err := pad.CreatePad(ctx, "foo", etherpadlite.OptionalParam)
```
If a method has a default argument, such as `copyPad(sourceID, destinationID[, force=false])` setting the parameter to `OptionalParam` will set the value to its default.

Using one `EtherpadLite` instance from multiple goroutines is safe (if you're not setting any values on it).
## License
Copyright 2017 Fabian Wenzelmann

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
