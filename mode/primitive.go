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
	"bufio"
	"errors"
	"fmt"
	"github.com/cockroachdb/cockroach/client"
	"github.com/cockroachdb/cockroach/proto"
	"github.com/twinj/uuid"
	"github.com/ugorji/go/codec"
	"os"
	"time"
)

// Default chunk size for storage of primitives,
// 255kb limit for values in mongo gridfs appears to offer
// a reasonable tradeoff between performance and memory
// usage
const CHUNK_SIZE = 255000

// Primitive is analgous to the mongodb gridfs File type,
// where it specifies the details of the blob that is stored/sliced
// into the datamode.Primitive subspace. The Primitive is stored in the
// datamode.Primitive.Meta subspace
type Primitive struct {
	Id       string `json:"id"` // UUID of the primitive
	Name     string `json:"name"`
	Length   int    `json:"length"`              // number of bytes written to database
	CSize    int    `json:"chunkSize,omitempty"` // size of chunks in this primitive
	Chunks   int    `json:"chunks,omitempty"`    // total number of chunks written to database
	Created  string `json:"created,omitempty"`   // date file was created/uploaded
	Md5      string `json:"md5,omitempty"`       // md5 hash of file for comparison checking
	MimeType string `json:"mimeType,omitempty"`  // mime type
}

var pdb string
var metaDb string

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("%s took %s\n", name, elapsed)
}

// set up decoder
var mph codec.Handle = new(codec.MsgpackHandle)

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

// init creates and opens the connection to the FDB cluster
// need something here to specifiy which cluster and also provide
// authentication
func init() {
	// init called after variable initialization when file is loaded
	// here we change uuid format to to 'Clean' which gets rid of
	// default curly braces and dashes, cutting uuid length to 32 chars
	uuid.SwitchFormat(uuid.Clean, false)

	pdb = "primitive:"
	metaDb = pdb + "meta:"
}

// Make a new instance of Primitive, using the bytes read from the reader
// The only args required at this level of make, is the number of bytes expected
// to be on the reader, making this effectively a 'framed' read type of protocol
// header is optional, i.e. if present, will write RES_MSG onto readWriter
func (p *Primitive) Make(reader *bufio.Reader) error {
	defer timeTrack(time.Now(), "primtive.Make")
	// Create a new uuid for this primitive
	var id = uuid.NewV4().String()
	var numBytes, chunks int
	if p.Length == 0 {
		return MISSING_ARG
	}
	e := kvClient.RunTransaction(&client.TransactionOptions{Isolation: proto.SNAPSHOT}, func(txn *client.KV) error {
		buf := make([]byte, CHUNK_SIZE)
		for {

			// call read on readWriter until buffer is filled, or EOF
			var readOffset int
			for {
				// get the first chunk of bytes, increment bytes read
				n1, err := reader.Read(buf[readOffset:])
				readOffset += n1
				//fmt.Printf("Primtive.Make: read:%d offset%d error:%q\n", n1, readOffset, err)
				if err == EOF || readOffset+n1 == CHUNK_SIZE {
					//fmt.Printf("Primtive.Make: filled buffer:%d byteserror:%q\n", readOffset, err)
					break
				} else if numBytes+readOffset >= p.Length {
					break
				} else if err != nil && err != EOF {
					//fmt.Printf("\nPrimitive.Make err not nil or EOF:%q", err)
					return err
				}
			}

			if readOffset > 0 {
				// process bytes returned from Read before error
				numBytes = numBytes + readOffset
				// create key for this chunk
				ks := []byte(fmt.Sprintf("%s%s:%10d", pdb, id, chunks))
				key := proto.Key(ks)
				chunks = chunks + 1

				//fmt.Println("Primitive.Make: check amount read:", key, numBytes, readOffset, p.Length)
				// check if n1 < CHUNK_SIZE, if so slice off blank end
				putResp := &proto.PutResponse{}
				if readOffset < CHUNK_SIZE {
					//fmt.Println("Primitive.Make: readOffset < than chunk size:", numBytes, readOffset, chunks)
					sbuf := buf[:readOffset]
					if err := kvClient.Call(proto.Put, proto.PutArgs(key, sbuf), putResp); err != nil {
						return err
					}
					break
				} else {
					if err := kvClient.Call(proto.Put, proto.PutArgs(key, buf), putResp); err != nil {
						return err
					}
				}

			} else if readOffset == 0 {
				//fmt.Printf("\nPrimitive.Make n1 -s zero, so break")
				break
			}
		}
		//fmt.Println("Primitive.Make FINISHED READING BYTES:", numBytes, " expected:", p.Length)
		// check to see if bytes read equals number expected, stored in the original p.Size
		if p.Length != numBytes {
			//fmt.Printf("\nPrimitive.Make p.Length:%d not equal to numBytes:%d\n", p.Length, numBytes)
			return errors.New(fmt.Sprintf("bytes read %d doesn't match expected %d", numBytes, p.Length))
		}
		p.Id = id
		p.Chunks = chunks
		p.CSize = CHUNK_SIZE
		//fmt.Printf("\nPrimitive.Make assigning length: %d id: %s\n", numBytes, id)
		//fmt.Printf("\nPrimitive.Make length and id assigned\n")
		p.SetMeta()
		return nil
	})

	return e
}

// Find an instance of Primitive, using the id arg provided in the args map
// If it's found return id and number of bytes read in reply,
// otherwise return an error "Primitive Not Found"
func (p *Primitive) Find() error {
	err := p.Meta() // p is now filled out
	if err != nil {
		return err
	}
	return nil
}

// Read the keys and values one row at a time, writing the value onto the stream
// Requires retrieving Meta first, to know how many chunks are there and to be
// able to generate the correct keys
func (p *Primitive) Stream(writer *bufio.Writer) error {
	defer timeTrack(time.Now(), "primtive.Stream")
	var b int
	err := p.Meta() // p is now filled out
	if err != nil {
		return err
	}
	for i := 0; i < p.Chunks; i++ {
		getResp := &proto.GetResponse{}
		key := proto.Key(fmt.Sprintf("%s%s:%10d", pdb, p.Id, i))
		if err := kvClient.Call(proto.Get, proto.GetArgs(key), getResp); err != nil {
			return err
		}
		p, e := writer.Write(getResp.Value.Bytes)
		if e != nil {
			return e
		}
		b = b + p
	}
	return nil
}

// Destroy the bytes associated with the id arg provided in the args map
// If the file is found, return id and number of bytes destroyed in reply
// otherwise return an error "Primitive Not Found"
func (p *Primitive) Destroy() error {

	err := p.Meta() // p is now filled out
	if err != nil {
		return err
	}
	for i := 0; i < p.Chunks; i++ {
		delReq := &proto.DeleteRequest{}
		delReq.Key = proto.Key(fmt.Sprintf("%s%s:%10d", pdb, p.Id, i))
		delResp := &proto.DeleteResponse{}
		err := kvClient.Call(proto.Delete, delReq, delResp)
		if err != nil {
			return err
		}
	}
	return p.DestroyMeta()
}

func (p *Primitive) Meta() error {
	if p.Id == "" || len(p.Id) != 32 {
		return errors.New(fmt.Sprintf("Invalid primtive id:%s", p.Id))
	}

	key := proto.Key(fmt.Sprintf("%s%s", metaDb, p.Id))
	getResp := &proto.GetResponse{}
	if err := kvClient.Call(proto.Get, proto.GetArgs(key), getResp); err != nil {
		return err
	}
	if getResp.Value == nil {
		return NOT_FOUND
	}
	var dec *codec.Decoder = codec.NewDecoderBytes(getResp.Value.Bytes, mph)
	return dec.Decode(p)
}

func (p *Primitive) SetMeta() error {
	if p.Id == "" || len(p.Id) != 32 {
		return errors.New(fmt.Sprintf("Invalid primtive id:%s", p.Id))
	}
	// 1. encode primitive to an array of bytes
	var buf []byte
	var enc *codec.Encoder = codec.NewEncoderBytes(&buf, mph) // mph is the msgpack codec
	err := enc.Encode(p)                                      // p is now encoded in buf
	if err != nil {
		fmt.Println("error encoding primitive:", p, err)
		return err
	}
	// 2. set value of key (primitive.Id)
	key := proto.Key(fmt.Sprintf("%s%s", metaDb, p.Id))
	putResp := &proto.PutResponse{}
	err = kvClient.Call(proto.Put, proto.PutArgs(key, buf), putResp)
	if err != nil {
		fmt.Println("SetMeta:", key, buf)
	}

	return err
}

func (p *Primitive) DestroyMeta() error {
	if p.Id == "" || len(p.Id) != 32 {
		return errors.New(fmt.Sprintf("Invalid primtive id:%s", p.Id))
	}
	delReq := &proto.DeleteRequest{}
	delReq.Key = proto.Key(fmt.Sprintf("%s%s", metaDb, p.Id))
	delResp := &proto.DeleteResponse{}

	fmt.Println("DestroyMeta:", delReq.Key)

	err := kvClient.Call(proto.Delete, delReq, delResp)
	if err != nil {
		p.Id = ""
	}
	return err

}
