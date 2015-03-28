# roachclip-fs

roachclip-fs is a File Storage system for cockroachdb written in go
intended to model the mongodb gridfs file storage layer. roachclip-fs
is a direct port of our (now defunct) FoundationFS File Storage layer
built on top of FoundationDB. The port from FroundationFS to cockroachdb
took about two days, including tests and climbing the learning curve. Overall,
a good experience. We will continue to keep this up to date as cockroachdb
moves from alpha to beta to production ready software.

Includes an example http server interface with support for basic upload and download.

For details on cockroachdb visit their website at cockroachdb.org.

## Installation

```bash
go get github.com/cloudmode/roachclip-fs
```

## Dependencies

roachclip-fs requires:

Go 1.3+ with CGO enabled


To install and run the cockroach package interface (https://github.com/cockroachdb/cockroach) on Mac OSX:

```bash
$ go get github.com/github.com/cockroachdb/cockroach
$ cd $GOPATH/src/github.com/cockroachdb/cockroach
$ make build
$ docker run -d -p 8080:8080 "cockroachdb/cockroach" \
    init -rpc="localhost:0" \
    -stores="ssd=$(mktemp -d /tmp/db)"

```
Note the cockroach http address and port that the cockroach server is listening on. You'll
need that when you start the example, or use curl tests below.

It will look something like this:

```bash
I0328 13:04:55.704463   26591 server.go:162] Starting HTTP server at 192.168.0.2:8080

```

roachclip-fs uses go/codec (using the msgpack codec) for storing values in FDB
such as the file meta data. (https://github.com/ugorji/go)

To install the go/codec package:

```bash
go get github.com/ugorji/go/codec
```

Other dependencies:

```bash
go get github.com/twinj/uuid
```

## Example

To run the example, use the host address and host port printed out when you started the cockroach server above:

```bash
cd examples
go run simple.go -roachhost 192.168.0.2 -roachport 8080```

The server will be listening on localhost:9090. Direct your browser to localhost:9090/upload, upload
a file and json with the id of the uploaded file is returned. Copy and paste the id of the uploaded file
into http://localhost:9090/download?id=<the id returned> and the file will be displayed in your browser

To test with curl:

```bash
curl -i -X POST -H "Content-Type: multipart/form-data" -F "uploadfile=@test.png" http://localhost:9090/upload
curl -o foo.png http://localhost:9090/download?id=<the id returned from curl POST>

```

## Test Suite

The test suite uses the standard go test runner along with convey, download here.

```bash
go get github.com/smartystreets/goconvey/convey
cd test
go test
```
The test script expects there to be at least one file in the test/data subdirectory which is used
for uploading and downloading tests. It will fail if there are no files in data. 

## License

The MIT License (MIT) - see LICENSE.md for more details