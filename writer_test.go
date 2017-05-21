package cardpdf

import (
	"fmt"
	_ "image/jpeg"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

func TestWriteTwoSeparateImages(t *testing.T) {

	outFileName := "WriteTwoSeparateImages.pdf"
	outFile := createTestOutpuFile(t, outFileName)

	pdfWriter := NewPdfWriter()

	pdfWriter.WriteImage(getTestImage(t, "ali.jpg"), "ali.jpg", 1)
	pdfWriter.WriteImage(getTestImage(t, "lotus.jpg"), "lotus.jpg", 1)

	if err := pdfWriter.Output(outFile); err != nil {
		t.Fatalf("error writing pdf: %v", err)
	}

	if err := outFile.Close(); err != nil {
		t.Fatalf("error closing pdf file: %v", err)
	}

	fileSize := statTestOutputFile(t, outFileName).Size()
	if fileSize < 150000 {
		t.Errorf("expected file size > %d but was %d\n", 150000, fileSize)
	}

	fmt.Printf("File: %s Size: %d\n", outFileName, fileSize)

	deleteTestOutpuFile(t, outFileName)
}

func TestWriteTwoImagesNTimes(t *testing.T) {
	outFileName := "WriteTwoImagesNTimes.pdf"
	outFile := createTestOutpuFile(t, outFileName)

	pdfWriter := NewPdfWriter()

	pdfWriter.WriteImage(getTestImage(t, "ali.jpg"), "ali.jpg", 50)
	pdfWriter.WriteImage(getTestImage(t, "lotus.jpg"), "lotus.jpg", 50)

	if err := pdfWriter.Output(outFile); err != nil {
		t.Fatalf("error writing pdf: %v", err)
	}

	if err := outFile.Close(); err != nil {
		t.Fatalf("error closing pdf file: %v", err)
	}

	fileSize := statTestOutputFile(t, outFileName).Size()
	if fileSize < 200000 || fileSize > 201000 {
		t.Errorf("expected file size > %d and < %d but was %d\n", 200000, 201000, fileSize)
	}

	fmt.Printf("File: %s Size: %d\n", outFileName, fileSize)

	deleteTestOutpuFile(t, outFileName)
}

func TestImmediateClose(t *testing.T) {
	pdfWriter := NewPdfWriter()
	if err := pdfWriter.Output(ioutil.Discard); err != nil {
		t.Errorf("Expected no erro, but got: %s", err)
	}
}

func getTestImage(t *testing.T, name string) io.Reader {
	file, err := os.Open("./testdata/" + name)

	if err != nil {
		t.Fatalf("error opening file: %v", err)
	}
	return file
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

	pdfWriter := NewPdfWriter()

	fmt.Printf("BorderPadding with border: %v\n", pdfWriter.cardBorderPadding())
	pdfWriter.Border = false
	fmt.Printf("BorderPadding without border: %v\n", pdfWriter.cardBorderPadding())

	// Output:
	//
	// BorderPadding with border: 4.67775
	// BorderPadding without border: 0
}

func TestCardsPerPage(t *testing.T) {

	tests := []struct {
		name    string
		columns int
		rows    int
		want    int
	}{
		{"1 x 1", 1, 1, 1},
		{"0 x 1", 0, 1, 0},
		{"1 x 0", 1, 0, 0},
		{"3 x 3", 3, 3, 9},
		{"3 x 4", 3, 4, 12},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &PdfWriter{
				Columns: tt.columns,
				Rows:    tt.rows,
			}
			if got := w.cardsPerPage(); got != tt.want {
				t.Errorf("PdfWriter.cardsPerPage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMarginBottom(t *testing.T) {

	tests := []struct {
		name       string
		pageHeight Unit
		rows       int
		space      Unit
		want       Unit
	}{
		{"3 rows no space", A4Height, 3, 0, 42.997498},
		{"3 rows 1pt space", A4Height, 3, 1, 41.997498},
		{"2 rows 1pt space", 2*3.5*Inch + 1, 2, 1, 0},
		{"too small", 10, 2, 0, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &PdfWriter{
				Rows:       tt.rows,
				PageHeight: tt.pageHeight,
				Space:      tt.space,
			}
			if got := w.marginBottom(); got != tt.want {
				t.Errorf("PdfWriter.marginBottom() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMarginLeft(t *testing.T) {

	tests := []struct {
		name      string
		pageWidth Unit
		columns   int
		space     Unit
		want      Unit
	}{
		{"3 columns no space", A4Width, 3, 0, 27.675018},
		{"3 columns 1pt space", A4Width, 3, 1, 26.675018},
		{"2 columns 1pt space", 2*2.5*Inch + 1, 2, 1, 0},
		{"too small", 10, 2, 0, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &PdfWriter{
				Columns:   tt.columns,
				PageWidth: tt.pageWidth,
				Space:     tt.space,
			}
			if got := w.marginLeft(); got != tt.want {
				t.Errorf("PdfWriter.marginLeft() = %v, want %v", got, tt.want)
			}
		})
	}
}
