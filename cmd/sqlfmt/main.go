package main

import (
	"bytes"
	"flag"
	"go/printer"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/kanmu/go-sqlfmt"
	"golang.org/x/tools/imports"
)

var (
	srcFile    = flag.String("s", "", "the source file")
	outputFile = flag.String("o", "", "the output file")
)

const (
	tabWidth    = 8
	printerMode = printer.UseSpaces | printer.TabIndent
)

func main() {
	flag.Parse()

	if *srcFile == "" {
		log.Fatal("-s is required")
	}

	f, err := os.Open(*srcFile)
	if err != nil {
		log.Fatal(err)
	}

	sfmt, err := sqlfmt.NewSQLFormatter(f)
	if err != nil {
		log.Fatal(err)
	}

	if err := sfmt.Format(); err != nil {
		log.Println(err)
	}

	var buf bytes.Buffer
	cfg := printer.Config{Mode: printerMode, Tabwidth: tabWidth}
	err = cfg.Fprint(io.Writer(&buf), sfmt.Fset, sfmt.AstNode)
	if err != nil {
		log.Fatal(err)
	}
	sqlFormatted := buf.Bytes()

	importsFormatted, err := imports.Process(*srcFile, sqlFormatted, nil)
	if err != nil {
		log.Fatal(err)
	}

	if *outputFile == "" {
		if _, err := os.Stdout.Write(importsFormatted); err != nil {
			log.Fatal(err)
		}
		if os.Stdout.Sync(); err != nil {
			log.Fatal(err)
		}
	} else {
		if err = writeFile(*outputFile, importsFormatted); err != nil {
			log.Fatal(err)
		}
	}
}

// atomic write
func writeFile(filename string, bytes []byte) error {
	tmpFile, err := filepath.Abs(filename + ".")
	if err != nil {
		return err
	}
	f, err := ioutil.TempFile(filepath.Dir(tmpFile), filepath.Base(tmpFile))
	if err != nil {
		return err
	}
	if _, err := f.Write(bytes); err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err = os.Rename(f.Name(), filename); err != nil {
		return err
	}
	return nil
}
