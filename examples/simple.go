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

package main

import (
	//"crypto/md5"
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/roachclip-fs/mode"
	"html/template"
	//	"io"
	"flag"
	"log"
	"net/http"
	"os"
)

//Compile templates on start
var templates = template.Must(template.ParseFiles("tmpl/upload.html"))

//Display the named template
func display(w http.ResponseWriter, tmpl string, data interface{}) {
	templates.ExecuteTemplate(w, tmpl+".html", data)
}

// upload logic
func upload(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method)
	if r.Method == "GET" {
		display(w, "upload", nil)
	} else {
		//parse the multipart form in the request
		err := r.ParseMultipartForm(100000)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//get a ref to the parsed multipart form
		m := r.MultipartForm

		//get the *fileheaders
		files := m.File["myfiles"]
		p := new(mode.Primitive)
		for i, _ := range files {
			//for each fileheader, get a handle to the actual file
			file, err := files[i].Open()
			defer file.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			stat, err := file.(*os.File).Stat()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			//create destination file making sure the path is writeable.
			header := files[i].Header
			fmt.Printf("file:%d %T %#v\n", i, file, file)
			fmt.Printf("files[i]:type:\n%T \n%#v\n", header.Get("Content-Type"), header.Get("Content-Type"))
			//dst, err := os.Create("/tmp/" + files[i].Filename)

			p.Name = files[i].Filename
			p.MimeType = header.Get("Content-Type")
			p.Length = int(stat.Size())

			reader := bufio.NewReader(file)
			err = p.Make(reader)

			//defer dst.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			//copy the uploaded file to the destination file
			//if _, err := io.Copy(dst, file); err != nil {
			//	http.Error(w, err.Error(), http.StatusInternalServerError)
			//	return
			//}

		}
		js, err := json.Marshal(p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}

func download(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// check to see if id is in URL
		// otherwise return error
		id := r.FormValue("id")
		if id == "" || len(id) != 32 {
			w.Header().Set("Content-Type", "application/json")
			e := map[string]string{"success": "false", "error": "missing or invalid id in query"}
			js, err := json.Marshal(e)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(js)
		}
		writer := bufio.NewWriter(w)
		p := new(mode.Primitive)
		p.Id = id
		err := p.Stream(writer)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	} else {
		http.Error(w, "method not supported", http.StatusInternalServerError)
		return
	}
	return
}

func main() {
	hostname := flag.String("roachhost", "localhost", "a valid ip address")
	portnumber := flag.Int("roachport", 8080, "a valid port name")

	flag.Parse()

	fmt.Println("roachhost:", *hostname, " roachport:", *portnumber)

	mode.OpenRoach(*hostname, *portnumber)
	defer mode.CloseRoach()
	//http.HandleFunc("/", sayhelloName) // setting router rule
	//http.HandleFunc("/login", login)
	http.HandleFunc("/upload", upload)
	http.HandleFunc("/download", download)

	fmt.Println("Simple Server listening on http://localhost:9090/upload")
	fmt.Println("Simple Server download uri is http://localhost:9090/download?id=<id>")

	err := http.ListenAndServe(":9090", nil) // setting listening port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
