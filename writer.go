package cardpdf

import (
	"bitbucket.org/zombiezen/gopdf/pdf"
	"image"
	"io"
)

const (
	Cm                pdf.Unit = pdf.Cm
	Inch              pdf.Unit = pdf.Inch
	pageWidth         pdf.Unit = pdf.A4Width
	pageHeight        pdf.Unit = pdf.A4Height
	space             pdf.Unit = 0.00 * pdf.Cm
	cropSpace         pdf.Unit = 0.5 * pdf.Cm
	cropLineWidth     pdf.Unit = 0.1
	markLength        pdf.Unit = 0.5 * pdf.Cm
	cardWidth         pdf.Unit = 2.5 * pdf.Inch
	cardHeight        pdf.Unit = 3.5 * pdf.Inch
	cardBorderPadding pdf.Unit = 0.165 * pdf.Cm
	mm                pdf.Unit = 0.1 * pdf.Cm
	borderWidth       pdf.Unit = 0.4 * pdf.Cm
	columns           int      = 3
	rows              int      = 3
	cardsPerPage      int      = columns * rows
	defaultMargin     pdf.Unit = 2 * pdf.Cm
)

// retuns a new pdf writer. This writer writes Images to pdf file the dimensions are definte as constants
func NewPdfWriter(writer io.Writer) PdfWriter {
	return PdfWriter{
		doc:         pdf.New(),
		writer:      writer,
		Border:      true,
		CropLines:   true,
		Columns:     columns,
		Rows:        rows,
		PageWidth:   pageWidth,
		PageHeight:  pageHeight,
		BorderWidth: borderWidth,
		Space:       space,
	}
}

type PdfWriter struct {
	Border, CropLines     bool
	Columns, Rows         int
	PageWidth, PageHeight pdf.Unit
	BorderWidth, Space    pdf.Unit

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
	case this.imgCount%this.cardsPerPage() == 0:
		this.newPage()
		this.current = pdf.Point{X: this.marginLeft(), Y: this.marginBottom() + ((pdf.Unit(this.Rows - 1)) * this.Space) + (pdf.Unit((this.Rows - 1)) * cardHeight)}
		return this.current

	case this.imgCount%this.Columns == 0:
		this.current = pdf.Point{X: this.marginLeft(), Y: this.current.Y - this.Space - cardHeight}
		return this.current
	}
	this.current = pdf.Point{X: this.current.X + this.Space + cardWidth, Y: this.current.Y}
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
	this.page.SetLineWidth(this.BorderWidth)
	path := new(pdf.Path)
	path.Move(pdf.Point{this.current.X + mm, this.current.Y + mm})
	path.Line(pdf.Point{this.current.X + mm, this.current.Y + cardHeight - mm})
	path.Line(pdf.Point{this.current.X + cardWidth - mm, this.current.Y + cardHeight - mm})
	path.Line(pdf.Point{this.current.X + cardWidth - mm, this.current.Y + mm})
	path.Line(pdf.Point{this.current.X + mm - this.BorderWidth/2, this.current.Y + mm})
	this.page.Stroke(path)
}

// draw a single line
func (this *PdfWriter) drawLine(from, to pdf.Point) {
	path := new(pdf.Path)
	path.Move(from)
	path.Line(to)
	this.page.Stroke(path)
}

// draws the crop lines based on the current point (8 per card)
func (this *PdfWriter) drawCropLines() {
	this.page.SetLineWidth(cropLineWidth)

	for i := 0; i < 2; i++ {
		val := pdf.Pt * pdf.Unit(float32(i))
		// left
		this.drawLine(pdf.Point{this.current.X, this.current.Y + cardHeight*val},
			pdf.Point{this.current.X - cropSpace, this.current.Y + cardHeight*val})

		// top
		this.drawLine(pdf.Point{this.current.X + cardWidth*val, this.current.Y + cardHeight},
			pdf.Point{this.current.X + cardWidth*val, this.current.Y + cardHeight + cropSpace})

		// right
		this.drawLine(pdf.Point{this.current.X + cardWidth, this.current.Y + cardHeight*val},
			pdf.Point{this.current.X + cardWidth + cropSpace, this.current.Y + cardHeight*val})

		// top
		this.drawLine(pdf.Point{this.current.X + cardWidth*val, this.current.Y},
			pdf.Point{this.current.X + cardWidth*val, this.current.Y - cropSpace})
	}
}

// closes the last page (if one exists) and creates a new empty one
func (this *PdfWriter) newPage() {
	if this.page != nil {
		this.page.Close()
	}
	this.page = this.doc.NewPage(this.PageWidth, this.PageHeight)
}

// the last page is closed. The objects are encoded to a pdf file and written to the given writer
func (this *PdfWriter) Close() error {
	if this.page == nil {
		this.newPage()
	}
	this.page.Close()
	return this.doc.Encode(this.writer)
}

func (this *PdfWriter) cardsPerPage() int {
	return this.Columns * this.Rows
}

func (this *PdfWriter) marginBottom() pdf.Unit {
	return (this.PageHeight - (pdf.Unit(this.Rows)*cardHeight + (pdf.Unit(this.Rows)-1)*space)) / 2
}
func (this *PdfWriter) marginLeft() pdf.Unit {
	return (this.PageWidth - (pdf.Unit(this.Columns)*cardWidth + (pdf.Unit(this.Columns)-1)*space)) / 2
}

// if PdfWriter.Border is true the default setting is returned, 0.0 otherwiese
func (this *PdfWriter) cardBorderPadding() pdf.Unit {
	if this.Border {
		return cardBorderPadding
	}
	return pdf.Unit(0.0)
}
