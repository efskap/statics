**Statics**  
The purpose of Statics is to make it dead simple to embed external files into a Go binary.  
By default, it takes all of the files in your ./include folder and embeds them as byte arrays in map called files in a separate .go file called files.go.    

**Install**  
`go get github.com/jerblack/statics`  
`go install github.com/jerblack/statics`  

This will create a file called `statics[.exe]` in `GOPATH/bin`.

**Usage**  
Statics is intended to do the right thing by default so you can just run `statics` and assuming you have a folder called `./include ` with files inside, a new `./files.go` will be generated (or overwritten).  
However, the default behavior can be changed. Run `statics -h` to see the available options.  

```
Usage:

  statics [-p=./include] [-out=files.go] [-pkg=main] [-map=files]

Flags:

  -map string
        name of the generated files map (default "files")
  -out file
        output go file (default "files.go")
  -p string
        dir path with files to embed (default "./include")
  -pkg package
        package name of the go file (default "main")
  -verbose

```
Just be sure to re-run `statics` after modifying any of the files in your `./include` folder. My build script usually starts with something like `statics && go build`.

**Example**  
A typical generated `files.go` will look something like this.  
```go
package main

var files = map[string][]byte{

	"index.htm": []byte{108, 101, 116, ...},

	"favicon.png": []byte{137, 80, 78, ...},

	"code.js": []byte{60, 33, 68, 79, ...},

	"style.css": []byte{104, 116, 109, ...},
}

```

The example below shows how it can be used.  

```go
func main() {
    http.HandleFunc("/", serveFiles)
	log.Fatal(http.ListenAndServe(":80", nil))
}

func serveFiles(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path
	var fName string
	switch {
	case p == "/":
		fName = "index.html"
	case p == "/favicon.ico":
		fName = "favicon.png"
	case strings.HasPrefix(p, "/static"):
		fName = strings.TrimPrefix(p, "/static/")
	default:
		http.Error(w, http.StatusText(http.StatusNotFound),
			http.StatusNotFound)
		return
	}
	f, ok := files[fName]  // <- the files map is what is provided by statics
	if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound),
			http.StatusNotFound)
		return
	}
	ext := path.Ext(fName)
	mimes := map[string]string {
		".js": "application/javascript",
		".png": "images/png",
		".jpg": "images/jpeg",
		".css": "text/css",
		".html": "text/html",
	}
	mime, ok := mimes[ext]
	if !ok {
		mime = "application/octet-stream"
	}
	w.Header().Set("Content-Type", mime)
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(f)
    if err != nil {
        fmt.Println(err.Error())
    }
}

```

**Credit**  
Statics was distilled down from https://github.com/leighmcculloch/embedfiles.   
Look at that version if you want to retain the folder path of your files in the file name in the map.  
