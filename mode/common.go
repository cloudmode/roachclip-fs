// Copyright 2015 CloudMoDe, LLC.
//
// The MIT License (MIT)

// Copyright (c) 2015 cloudmode

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//
//
// Author: Michael McFall (mike@cloudmo.de)

package mode

import (
	"errors"
	"fmt"
	"github.com/cockroachdb/cockroach/client"
	"github.com/cockroachdb/cockroach/rpc"
	"github.com/cockroachdb/cockroach/storage"
	"net/http"
)

var EOF = errors.New("EOF")
var NOT_FOUND = errors.New("Primitive Not Found")
var MISSING_ARG = errors.New("missing required arg")

var kvClient *client.KV

func OpenRoach(hostname string, port int) {
	// Key Value Client initialization.

	//serverAddress := "192.168.0.2:8080"
	serverAddress := fmt.Sprintf("%s:%d", hostname, port)

	sender := client.NewHTTPSender(serverAddress, &http.Transport{
		TLSClientConfig: rpc.LoadInsecureTLSConfig().Config(),
	})
	kvClient = client.NewKV(sender, nil)
	kvClient.User = storage.UserRoot

}

func CloseRoach() {
	kvClient.Close()
}
