package cardpdf

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"io/ioutil"
	"os"
	"testing"
)

func TestWriteTwoSeparateImages(t *testing.T) {

	outFileName := "WriteTwoSeparateImages.pdf"
	outFile := createTestOutpuFile(t, outFileName)

	pdfWriter := NewPdfWriter(outFile)

	pdfWriter.WriteImage(getTestImage(t, "ali.jpg"), 1)
	pdfWriter.WriteImage(getTestImage(t, "lotus.jpg"), 1)

	if err := pdfWriter.Close(); err != nil {
		t.Fatalf("error writing pdf: %v", err)
	}

	if err := outFile.Close(); err != nil {
		t.Fatalf("error closing pdf file: %v", err)
	}

	fileSize := statTestOutputFile(t, outFileName).Size()
	if fileSize < 730000 {
		t.Errorf("expected file size > %d but was %d\n", 730000, fileSize)
	}

	fmt.Printf("File: %s Size: %d\n", outFileName, fileSize)

	deleteTestOutpuFile(t, outFileName)
}

func TestWriteTwoImagesNTimes(t *testing.T) {
	outFileName := "WriteTwoImagesNTimes.pdf"
	outFile := createTestOutpuFile(t, outFileName)

	pdfWriter := NewPdfWriter(outFile)

	pdfWriter.WriteImage(getTestImage(t, "ali.jpg"), 50)
	pdfWriter.WriteImage(getTestImage(t, "lotus.jpg"), 50)

	if err := pdfWriter.Close(); err != nil {
		t.Fatalf("error writing pdf: %v", err)
	}

	if err := outFile.Close(); err != nil {
		t.Fatalf("error closing pdf file: %v", err)
	}

	fileSize := statTestOutputFile(t, outFileName).Size()
	if fileSize < 740000 || fileSize > 750000 {
		t.Errorf("expected file size > %d and < %d but was %d\n", 740000, 750000, fileSize)
	}

	fmt.Printf("File: %s Size: %d\n", outFileName, fileSize)

	deleteTestOutpuFile(t, outFileName)
}

func getTestImage(t *testing.T, name string) image.Image {
	file, err := os.Open("./testdata/" + name)
	defer file.Close()
	if err != nil {
		t.Fatalf("error opening file: %v", err)
	}

	img, _, err := image.Decode(file)
	if err != nil {
		t.Fatalf("error decoding image: %v", err)
	}
	return img
}

func createTestOutpuFile(t *testing.T, name string) *os.File {
	file, err := os.Create("./testdata/" + name)
	if err != nil {
		t.Fatalf("error creating testfile %v", err)
	}
	return file
}

func deleteTestOutpuFile(t *testing.T, name string) {
	err := os.Remove("./testdata/" + name)
	if err != nil {
		t.Fatalf("error removing testfile %v", err)
	}
}

func statTestOutputFile(t *testing.T, name string) os.FileInfo {
	info, err := os.Stat("./testdata/" + name)
	if err != nil {
		t.Fatalf("error getting stats for testfile %v", err)
	}
	return info
}

func ExampleCardBorderPadding() {

	pdfWriter := NewPdfWriter(ioutil.Discard)

	fmt.Printf("BorderPadding with border: %v\n", pdfWriter.cardBorderPadding())
	pdfWriter.Border = false
	fmt.Printf("BorderPadding without border: %v\n", pdfWriter.cardBorderPadding())

	// Output:
	//
	// BorderPadding with border: 4.67775
	// BorderPadding without border: 0.00000
}
