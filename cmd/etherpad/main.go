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

package main

import (
	"fmt"
	"log"

	etherpadlite "github.com/FabianWe/etherpadlite-golang"
)

func main() {
	apiKey := "7e7d913826c58a5ded3c5bc2c426918650c92829ae6d30ca091441ea052a58a3"
	pad := etherpadlite.NewEtherpadLite(apiKey)
	response, err := pad.CreatePad(nil, "foo", etherpadlite.OptionalParam)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(response.Message, response.Data)
	fmt.Println(pad.DeletePad(nil, "foo"))
}
