package main

import (
	"flag"
	"fmt"
	"github.com/lgrote/cardpdf"
	"image"
	_ "image/jpeg"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

var directoryPath string
var outputFile string

func init() {
	const (
		defaultDirectory  = "./"
		defaultOutputFile = "./output.pdf"
	)
	flag.StringVar(&directoryPath, "in", defaultDirectory, "the input directory")
	flag.StringVar(&outputFile, "out", defaultOutputFile, "the output file")
	flag.Parse()
}

func main() {
	fmt.Printf("read files in directory: %s\n", directoryPath)
	fmt.Printf("writing output to: %s\n", outputFile)

	outFile := getOutputFile()
	defer outFile.Close()
	pdfWriter := cardpdf.NewPdfWriter(outFile)

	files, err := ioutil.ReadDir(directoryPath)
	if err != nil {
		panic(err)
	}

	count := 0
	for _, info := range files {
		if strings.HasSuffix(info.Name(), ".jpg") || strings.HasSuffix(info.Name(), ".jpeg") {
			file, err := os.Open(directoryPath + string(os.PathSeparator) + info.Name())
			defer file.Close()
			if err != nil {
				panic(err)
			}
			img, _, err := image.Decode(file)
			pdfWriter.WriteImage(img, getCountFromFileName(info.Name()))
			count++
		}
	}

	if err := pdfWriter.Close(); err != nil {
		panic(err)
	}

	if count == 0 {
		outFile.Close()
		os.Remove(outputFile)
		fmt.Printf("there is no jpg in %s so no pdf has been generated", directoryPath)
	}
}

func getCountFromFileName(name string) int {
	res, err := strconv.ParseInt(name[0:1], 10, 32)
	if err != nil {
		return 1
	}
	return int(res)
}

func getOutputFile() *os.File {

	file, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}
	return file
}
