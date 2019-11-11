package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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

func contains(term string, sl *[]string) bool {
	term = strings.ToLower(term)
	for _, item := range *sl {
		if term == item {
			return true
		}
	}
	return false
}

func parseExclude(ef string) *[]string {
	var arr []string
	ef = strings.ToLower(ef)
	for _, x := range strings.Split(ef, "|") {
		x = strings.TrimSpace(x)
		arr = append(arr, x)
	}
	return &arr
}

func main() {
	out := flag.String("out", "files.go", "output go `file`")
	pkg := flag.String("pkg", "main", "`package` name of the go file")
	include := flag.String("p", "./include", "dir path with files to embed")
	fileMap := flag.String("map", "files", "name of the generated files map")
	verbose := flag.Bool("v", false, "verbose")
	keepDirs := flag.Bool("k", false, "retain directory path in file names used as keys in file map.\ndirname/filename stays dirname/filename instead of just filename in the file map")
	excludeFiles := flag.String("x", "", "pipe-separated list of files in include path to exclude.\nFiles in include or subfolders of include with matching name will be excluded.\nSurround whole list with quotes like: \"file1 | file2 | file3\"")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "By default, statics takes all of the files in your ./include folder and embeds them as byte arrays in a map called files in a separate .go file called files.go.\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\n")
		fmt.Fprintf(os.Stderr, "  statics [-p=./include] [-out=files.go] [-pkg=main] [-map=files] [-k] [-x=\"file1 | file2 | file3\"] [-v]\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	inputPath := include
	excludes := parseExclude(*excludeFiles)

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

		if contains(filepath.Base(path), excludes) {
			if *verbose {
				fmt.Printf("skipping exluded file: %s\n", path)
			}
			return nil
		}

		path = filepath.ToSlash(path)
		path = strings.TrimPrefix(path, fmt.Sprintf("%s/", *inputPath)[2:])
		if !*keepDirs {
			path = filepath.Base(path)
		}
		if *verbose {
			fmt.Printf("Added key to files map: %s\n", path)
		}

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