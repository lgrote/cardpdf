package cardpdf

import (
	"bitbucket.org/zombiezen/gopdf/pdf"
	"image"
	"io"
)

const (
	pageWidth         pdf.Unit = pdf.A4Width
	pageHeight        pdf.Unit = pdf.A4Height
	space             pdf.Unit = 0.00 * pdf.Cm
	cropSpace         pdf.Unit = 0.5 * pdf.Cm
	markLength        pdf.Unit = 0.5 * pdf.Cm
	cardWidth         pdf.Unit = 2.5 * pdf.Inch
	cardHeight        pdf.Unit = 3.5 * pdf.Inch
	cardBorderPadding pdf.Unit = 0.165 * pdf.Cm
	mm                pdf.Unit = 0.1 * pdf.Cm
	borderWidth       pdf.Unit = 0.4 * pdf.Cm
	columns           int      = 3
	rows              int      = 3
	cardsPerPage      int      = columns * rows
	marginBottom      pdf.Unit = (pdf.A4Height - (pdf.Unit(rows)*cardHeight + (pdf.Unit(rows)-1)*space)) / 2
	marginLeft        pdf.Unit = (pdf.A4Width - (pdf.Unit(columns)*cardWidth + (pdf.Unit(columns)-1)*space)) / 2
)

// retuns a new pdf writer. This writer writes Images to pdf file the dimensions are definte as constants
func NewPdfWriter(writer io.Writer) PdfWriter {
	return PdfWriter{
		doc:       pdf.New(),
		writer:    writer,
		Border:    true,
		CropLines: true,
	}
}

type PdfWriter struct {
	Border    bool
	CropLines bool
	writer    io.Writer
	doc       *pdf.Document
	page      *pdf.Canvas
	imgCount  int
	current   pdf.Point
}

// writes an Image  n-times to the pdf file. With crop lines and black border
func (this *PdfWriter) WriteImage(img image.Image, count int) {
	ref := this.addImage(img)
	for i := 0; i < count; i++ {
		this.setCurrentPoint()
		if this.CropLines {
			this.drawCropLines()
		}
		this.drawImageReference(ref)
		if this.Border {
			this.drawBorder()
		}
		this.imgCount++
	}
}

// based on the imageCount the lower left corner of the next image is computed
func (this *PdfWriter) setCurrentPoint() pdf.Point {
	switch {
	case this.imgCount%cardsPerPage == 0:
		this.newPage()
		this.current = pdf.Point{X: marginLeft, Y: marginBottom + ((pdf.Unit(rows - 1)) * space) + (pdf.Unit((rows - 1)) * cardHeight)}
		return this.current

	case this.imgCount%columns == 0:
		this.current = pdf.Point{X: marginLeft, Y: this.current.Y - space - cardHeight}
		return this.current
	}
	this.current = pdf.Point{X: this.current.X + space + cardWidth, Y: this.current.Y}
	return this.current
}

// adds an Image to the pdf and returns its reference. Note that the Image isn't yet drawn, it's just added.
func (this *PdfWriter) addImage(img image.Image) pdf.Reference {
	return this.doc.AddImage(img)
}

// draws an ImageReference in the pdf. The Image to be drawn needs to be added to the file first.
func (this *PdfWriter) drawImageReference(ref pdf.Reference) {
	this.page.DrawImageReference(ref, pdf.Rectangle{
		Min: pdf.Point{X: this.current.X + this.cardBorderPadding(), Y: this.current.Y + this.cardBorderPadding()},
		Max: pdf.Point{X: this.current.X + cardWidth - this.cardBorderPadding(), Y: this.current.Y + cardHeight - this.cardBorderPadding()},
	})
}

// draws a black border based on the current point
func (this *PdfWriter) drawBorder() {
	this.page.SetLineWidth(borderWidth)
	path := new(pdf.Path)
	path.Move(pdf.Point{this.current.X + mm, this.current.Y + mm})
	path.Line(pdf.Point{this.current.X + mm, this.current.Y + cardHeight - mm})
	path.Line(pdf.Point{this.current.X + cardWidth - mm, this.current.Y + cardHeight - mm})
	path.Line(pdf.Point{this.current.X + cardWidth - mm, this.current.Y + mm})
	path.Line(pdf.Point{this.current.X + mm - borderWidth/2, this.current.Y + mm})
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
		drawCropLine(pdf.Point{this.current.X, this.current.Y + cardHeight*val},
			pdf.Point{this.current.X - cropSpace, this.current.Y + cardHeight*val})

		// top
		drawCropLine(pdf.Point{this.current.X + cardWidth*val, this.current.Y + cardHeight},
			pdf.Point{this.current.X + cardWidth*val, this.current.Y + cardHeight + cropSpace})

		// right
		drawCropLine(pdf.Point{this.current.X + cardWidth, this.current.Y + cardHeight*val},
			pdf.Point{this.current.X + cardWidth + cropSpace, this.current.Y + cardHeight*val})

		// top
		drawCropLine(pdf.Point{this.current.X + cardWidth*val, this.current.Y},
			pdf.Point{this.current.X + cardWidth*val, this.current.Y - cropSpace})
	}
}

// closes the last page (if one exists) and creates a new empty one
func (this *PdfWriter) newPage() {
	if this.page != nil {
		this.page.Close()
	}
	this.page = this.doc.NewPage(pageWidth, pageHeight)
}

// the last page is closed. The objects are encoded to a pdf file and written to the given writer
func (this *PdfWriter) Close() error {
	if this.page == nil {
		this.newPage()
	}
	this.page.Close()
	return this.doc.Encode(this.writer)
}
func (this *PdfWriter) cardBorderPadding() pdf.Unit {
	if this.Border {
		return cardBorderPadding
	}
	return pdf.Inch * 0.0
}
