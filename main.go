package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

const tmpl = `
package {{.Package}}

var {{.FilesVar}} = map[string][]byte{
{{range $name, $bytes := .Files}}
	"{{$name}}": []byte{ {{range $bytes}}{{.}},{{end}} },
{{end}}
}
`

type tmplData struct {
	Package      string
	Files        map[string][]byte
	FilesVar     string
}

func main() {
	out := flag.String("out", "files.go", "output go `file`")
	pkg := flag.String("pkg", "main", "`package` name of the go file")
	include := flag.String("p", "./include", "dir path with files to embed")
	fileMap := flag.String("map", "files", "name of the generated files map")
	verbose := flag.Bool("verbose", false, "")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "By default, statics takes all of the files in your ./include folder and embeds them as byte arrays in a map called files in a separate .go file called files.go.\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\n")
		fmt.Fprintf(os.Stderr, "  statics [-p=./include] [-out=files.go] [-pkg=main] [-map=files]\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	inputPath := include


	f, err := os.Create(*out)
	if err != nil {
		printErr("creating file", err)
		return
	}

	files := map[string][]byte{}
	err = filepath.Walk(*inputPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walking: %s", err)
		}

		if info.IsDir() {
			return nil
		}

		if *verbose {
			fmt.Printf("%s ", path)
		}

		contents, err := ioutil.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading file: %s", err)
		}
		if *verbose {
			fmt.Printf("(%d bytes)\n", len(contents))
		}

		path = filepath.Base(path)
		files[path] = contents
		return nil
	})
	if err != nil {
		printErr("walking", err)
		return
	}

	t, err := template.New("").Parse(tmpl)
	if err != nil {
		printErr("parsing template", err)
		return
	}

	buf := bytes.Buffer{}
	err = t.Execute(&buf, &tmplData{Package: *pkg, Files: files, FilesVar: *fileMap})
	if err != nil {
		printErr("generating code", err)
		return
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		printErr("formatting code", err)
		return
	}

	_, _ = f.Write(formatted)
	err = f.Close()
	if err != nil {
		printErr("finalizing file", err)
		return
	}
}

func printErr(doing string, err error) {
	fmt.Fprintf(os.Stderr, "Error %s: %s\n", doing, err)
}
