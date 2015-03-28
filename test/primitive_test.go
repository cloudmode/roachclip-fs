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
	//"fmt"
	"github.com/roachclip-fs/mode"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"
)

func TestPrimitive(t *testing.T) {
	Convey("Set the meta data for a non-existent Primitive", t, func() {
		primitive := mode.Primitive{"e64a919ef57c4481bcd5fba43f8efb9c", "sample.jpg", 4, 5, 6, "one", "two", "image/jpg"}
		err := primitive.SetMeta()
		So(err, ShouldEqual, nil)
		//fmt.Println("Primitive.Meta after set:", primitive)
		So(primitive.MimeType, ShouldEqual, "image/jpg")
		Convey("Read the meta data for the same Primitive", func() {
			primitive := mode.Primitive{"e64a919ef57c4481bcd5fba43f8efb9c", "", 0, 0, 0, "", "", ""}
			err := primitive.Meta()
			So(err, ShouldEqual, nil)
			So(primitive.MimeType, ShouldEqual, "image/jpg")
			So(primitive.Name, ShouldEqual, "sample.jpg")
		})
		Convey("Read the meta data for the same Primitive", func() {
			primitive := mode.Primitive{"e64a919ef57c4481bcd5fba43f8efb9c", "", 0, 0, 0, "", "", ""}
			err := primitive.DestroyMeta()
			So(err, ShouldEqual, nil)
			Convey("Meta data should not exist", func() {
				primitive := mode.Primitive{"e64a919ef57c4481bcd5fba43f8efb9c", "", 0, 0, 0, "", "", ""}
				err := primitive.Meta()
				So(err, ShouldNotEqual, nil)
				//fmt.Println("Primitive.Meta:", primitive)
				So(primitive.Id, ShouldEqual, "e64a919ef57c4481bcd5fba43f8efb9c")
				So(primitive.Name, ShouldEqual, "")
			})
		})
	})
	dataDir := "./data/"
	var files []os.FileInfo
	candidates, errDir := ioutil.ReadDir(dataDir)
	for _, s := range candidates {
		if !strings.HasPrefix(s.Name(), ".") {
			files = append(files, s)
		}
	}

	Convey("There should be at least one file in test data directory", t, func() {
		So(errDir, ShouldEqual, nil)
		So(len(files), ShouldBeGreaterThanOrEqualTo, 1)
	})

	for _, s := range files {
		Convey("Primitive: write and read files of various sizes and types", t, func() {

			inputFile := "./data/" + s.Name()
			file, err := os.Open(inputFile) // For read access.
			if err != nil {
				log.Fatal(err)
			}
			stat, err := file.Stat()
			if err != nil {
				log.Fatal(err)
			}
			primitive := mode.Primitive{"", "", int(stat.Size()), 0, 0, "", "", ""}
			reader := bufio.NewReader(file)
			primitive.Length = int(stat.Size())
			err = primitive.Make(reader)
			if err != nil {
				log.Fatal(err)
			}
			Convey("Given a Reader, make a new primitive", func() {
				Convey("the value of size should be the same as the original file", func() {
					So(primitive.Length, ShouldEqual, stat.Size())
				})
			})
			Convey("Given a Writer, write primitive to a file", func() {
				outputFile := "./data/" + primitive.Id
				file, err := os.Create(outputFile) // For write access.
				if err != nil {
					log.Fatal(err)
				}
				defer file.Close()

				var readPrimitive mode.Primitive
				readPrimitive.Id = primitive.Id
				readFile := bufio.NewWriter(file)
				e := readPrimitive.Stream(readFile)
				file.Sync()

				So(e, ShouldEqual, nil)
				So(readPrimitive.Length, ShouldEqual, stat.Size())

			})
			Convey("Delete the primitive", func() {
				e := primitive.Destroy()
				So(e, ShouldEqual, nil)
			})
			Convey("Try to read the primitive that was just deleted", func() {
				// returns zero bytes read, meaning the primitive does not exist
				var readPrimitive mode.Primitive
				readPrimitive.Id = primitive.Id

				e := readPrimitive.Find()
				So(e, ShouldEqual, mode.NOT_FOUND)
			})
		})
	}
}
