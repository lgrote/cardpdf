package cardpdf

import (
	"bitbucket.org/zombiezen/gopdf/pdf"
	"image"
	"io"
)

const (
	Space         pdf.Unit = 0.00 * pdf.Cm
	CropSpace     pdf.Unit = 0.5 * pdf.Cm
	MarkLengh     pdf.Unit = 0.5 * pdf.Cm
	CardW         pdf.Unit = 2.5 * pdf.Inch
	CardH         pdf.Unit = 3.5 * pdf.Inch
	PicBorder     pdf.Unit = 0.165 * pdf.Cm
	Mm            pdf.Unit = 0.1 * pdf.Cm
	BorderWidth   pdf.Unit = 0.4 * pdf.Cm
	pdfColumns    int      = 3
	pdfRows       int      = 3
	pdfImgPerPage int      = pdfColumns * pdfRows
	MarginBottom  pdf.Unit = (pdf.A4Height - (pdf.Unit(pdfRows)*CardH + (pdf.Unit(pdfRows)-1)*Space)) / 2
	MarginLeft    pdf.Unit = (pdf.A4Width - (pdf.Unit(pdfColumns)*CardW + (pdf.Unit(pdfColumns)-1)*Space)) / 2
)

// retuns a new pdf writer. This writer writes Images to pdf file the dimensions are definte as constants
func NewPdfWriter(writer io.Writer) PdfWriter {
	return PdfWriter{
		doc:    pdf.New(),
		writer: writer,
	}
}

type PdfWriter struct {
	writer   io.Writer
	doc      *pdf.Document
	page     *pdf.Canvas
	imgCount int
	current  pdf.Point
}

// writes an Image  n-times to the pdf file. With crop lines and black border
func (this *PdfWriter) WriteImage(img image.Image, count int) {
	ref := this.addImage(img)
	for i := 0; i < count; i++ {
		this.setCurrentPoint()
		this.drawCropLines()
		this.drawImageReference(ref)
		this.drawBorder()
		this.imgCount++
	}
}

// based on the imageCount the lower left corner of the next image is computed
func (this *PdfWriter) setCurrentPoint() pdf.Point {
	switch {
	case this.imgCount%pdfImgPerPage == 0:
		this.newPage()
		this.current = pdf.Point{X: MarginLeft, Y: MarginBottom + ((pdf.Unit(pdfRows - 1)) * Space) + (pdf.Unit((pdfRows - 1)) * CardH)}
		return this.current

	case this.imgCount%pdfColumns == 0:
		this.current = pdf.Point{X: MarginLeft, Y: this.current.Y - Space - CardH}
		return this.current
	}
	this.current = pdf.Point{X: this.current.X + Space + CardW, Y: this.current.Y}
	return this.current
}

// adds an Image to the pdf and returns its reference. Note that the Image isn't yet drawn, it's just added.
func (this *PdfWriter) addImage(img image.Image) pdf.Reference {
	return this.doc.AddImage(img)
}

// draws an ImageReference in the pdf. The Image to be drawn needs to be added to the file first.
func (this *PdfWriter) drawImageReference(ref pdf.Reference) {
	this.page.DrawImageReference(ref, pdf.Rectangle{
		Min: pdf.Point{X: this.current.X + PicBorder, Y: this.current.Y + PicBorder},
		Max: pdf.Point{X: this.current.X + CardW - PicBorder, Y: this.current.Y + CardH - PicBorder},
	})
}

// draws a black border based on the current point
func (this *PdfWriter) drawBorder() {
	this.page.SetLineWidth(BorderWidth)
	path := new(pdf.Path)
	path.Move(pdf.Point{this.current.X + Mm, this.current.Y + Mm})
	path.Line(pdf.Point{this.current.X + Mm, this.current.Y + CardH - Mm})
	path.Line(pdf.Point{this.current.X + CardW - Mm, this.current.Y + CardH - Mm})
	path.Line(pdf.Point{this.current.X + CardW - Mm, this.current.Y + Mm})
	path.Line(pdf.Point{this.current.X + Mm - BorderWidth/2, this.current.Y + Mm})
	this.page.Stroke(path)
}

// draws the crop lines based on the current point (8 per card)
func (this *PdfWriter) drawCropLines() {
	this.page.SetLineWidth(0.1)

	drawCropLine := func(from, to pdf.Point) {
		path := new(pdf.Path)
		path.Move(from)
		path.Line(to)
		this.page.Stroke(path)
	}

	for i := 0; i < 2; i++ {
		val := pdf.Pt * pdf.Unit(float32(i))
		// left
		drawCropLine(pdf.Point{this.current.X, this.current.Y + CardH*val},
			pdf.Point{this.current.X - CropSpace, this.current.Y + CardH*val})

		// top
		drawCropLine(pdf.Point{this.current.X + CardW*val, this.current.Y + CardH},
			pdf.Point{this.current.X + CardW*val, this.current.Y + CardH + CropSpace})

		// right
		drawCropLine(pdf.Point{this.current.X + CardW, this.current.Y + CardH*val},
			pdf.Point{this.current.X + CardW + CropSpace, this.current.Y + CardH*val})

		// top
		drawCropLine(pdf.Point{this.current.X + CardW*val, this.current.Y},
			pdf.Point{this.current.X + CardW*val, this.current.Y - CropSpace})
	}
}

// closes the last page (if one exists) and creates a new empty one
func (this *PdfWriter) newPage() {
	if this.page != nil {
		this.page.Close()
	}
	this.page = this.doc.NewPage(pdf.A4Width, pdf.A4Height)
}

// the last page is closed. The objects are encoded to a pdf file and written to the given writer
func (this *PdfWriter) Close() error {

	if this.page != nil {
		this.page.Close()
	}

	return this.doc.Encode(this.writer)
}
