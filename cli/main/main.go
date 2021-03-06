package main

import (
	"flag"
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"os"
	"runtime/pprof"
	"strconv"
	"strings"

	"github.com/lgrote/cardpdf"
)

var directoryPath string
var outputFile string
var memprofile = flag.String("memprofile", "", "write memory profile to this file")

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

	files := getMatchingFiles(directoryPath, func(fInfo os.FileInfo) bool {
		return strings.HasSuffix(fInfo.Name(), ".jpg") || strings.HasSuffix(fInfo.Name(), ".jpeg")
	})

	if len(files) == 0 {
		fmt.Printf("there is no jpg in %s so no pdf has been generated", directoryPath)
		return
	}

	outFile := getOutputFile()
	defer outFile.Close()
	pdfWriter := cardpdf.NewPdfWriter()

	for _, info := range files {
		file, err := os.Open(directoryPath + string(os.PathSeparator) + info.Name())
		defer file.Close()
		if err != nil {
			panic(err)
		}
		//img, _, err := image.Decode(file)
		fmt.Println("write image")
		pdfWriter.WriteImage(file, info.Name(), getCountFromFileName(info.Name()))
	}

	if err := pdfWriter.Output(outFile); err != nil {
		panic(err)
	}
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			panic(err)
		}
		pprof.WriteHeapProfile(f)
		f.Close()
		return
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

func getMatchingFiles(dir string, matches func(os.FileInfo) bool) []os.FileInfo {

	matchingFiles := make([]os.FileInfo, 0)
	files, err := ioutil.ReadDir(directoryPath)
	if err != nil {
		panic(err)
	}
	for _, fInfo := range files {
		if matches(fInfo) {
			matchingFiles = append(matchingFiles, fInfo)
		}
	}
	return matchingFiles
}
