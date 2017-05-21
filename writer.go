package cardpdf

import (
	"io"

	"github.com/jung-kurt/gofpdf"
)

const (
	// Pt is one Point
	Pt Unit = 1
	// Inch is 72 Points
	Inch Unit = 72
	// Cm is 28.85 Points
	Cm Unit = 28.35
	// Mm is Cm/10
	Mm Unit = Cm / 10
	// A4Width in Points
	A4Width Unit = 210 * Mm
	// A4Height in Points
	A4Height Unit = 297 * Mm

	pageWidth         Unit = A4Width
	pageHeight        Unit = A4Height
	space             Unit = 0.00 * Mm
	cropSpace         Unit = 5 * Mm
	cropLineWidth     Unit = 0.1
	markLength        Unit = 5 * Mm
	cardWidth         Unit = 2.5 * Inch
	cardHeight        Unit = 3.5 * Inch
	cardBorderPadding Unit = 1.65 * Mm
	borderWidth       Unit = 4 * Mm
	columns           int  = 3
	rows              int  = 3
	cardsPerPage      int  = columns * rows
	defaultMargin     Unit = 20 * Mm
)

// Unit is a device-independent dimensional type. This represents 1/72 of an inch.
type Unit float32

// Point is a 2D point.
type Point struct {
	X, Y Unit
}

// NewPdfWriter creates a new pdf writer. The writer writes Images to pdf file
// the dimensions are definte as constants
func NewPdfWriter() *PdfWriter {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetTitle("Your Proxies", true)
	pdf.SetAuthor("proxy-mat.appspot.com", true)
	pdf.SetDrawColor(0, 0, 0)
	pdf.SetCompression(false)

	return &PdfWriter{
		doc:         pdf,
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

// PdfWriter can write images to a pdf with borders and crop lines
type PdfWriter struct {
	// Should the Image have a black Border? Should crop lines be drawn?
	Border, CropLines bool
	// The number of columns/rows per page
	Columns, Rows int
	// the size of a page
	PageWidth, PageHeight Unit
	// how thick is the border. The space between images (default to 0.0)
	BorderWidth, Space Unit

	doc      *gofpdf.Fpdf
	imgCount int
	current  Point
}

// WriteImage writes an Image  n-times to the pdf file. With crop lines and black border.
// The name is used as identifier, so each image is only  added to the file once and afterwards refernced.
func (w *PdfWriter) WriteImage(r io.Reader, name string, count int) {

	w.addImage(r, name)
	for i := 0; i < count; i++ {
		w.incrPosition()
		if w.CropLines {
			w.drawCropLines()
		}
		w.drawImage(name)
		if w.Border {
			w.drawBorder()
		}
		w.imgCount++
	}
}

// Output encodes the everything that has been added so far to a pdf and writes it to the given Writer.
func (w *PdfWriter) Output(writer io.Writer) error {
	return w.doc.Output(writer)
}

// incrPosition calculates the lower left corner of the next image based on the imageCount
func (w *PdfWriter) incrPosition() {
	switch {
	case w.imgCount%w.cardsPerPage() == 0:
		w.newPage()
		w.current = Point{
			X: w.marginLeft(),
			Y: w.marginBottom() +
				((Unit(w.Rows - 1)) * w.Space) +
				(Unit((w.Rows - 1)) * cardHeight)}

	case w.imgCount%w.Columns == 0:
		w.current = Point{
			X: w.marginLeft(),
			Y: w.current.Y - w.Space - cardHeight}
	}
	w.current = Point{
		X: w.current.X + w.Space + cardWidth,
		Y: w.current.Y}
}

//addImage adds an Image to the pdf and returns its reference. Note that the Image
// isn't yet drawn, it's just added to the file.
func (w *PdfWriter) addImage(r io.Reader, name string) *gofpdf.ImageInfoType {
	iit := w.doc.RegisterImageOptionsReader(name, gofpdf.ImageOptions{
		ImageType: "JPG",
		ReadDpi:   false,
	}, r)
	return iit
}

// drawImage draws a visible image in the pdf. The Image to be drawn needs to be added
// to the file first. It's referenced by name.
func (w *PdfWriter) drawImage(name string) {
	w.doc.ImageOptions(name,
		float64((w.current.X+w.cardBorderPadding())/Mm),
		float64((w.current.Y+w.cardBorderPadding())/Mm),
		float64((cardWidth-(2*w.cardBorderPadding()))/Mm),
		float64((cardHeight-(2*w.cardBorderPadding()))/Mm),
		false,
		gofpdf.ImageOptions{
			ImageType: "JPG",
			ReadDpi:   false,
		}, 0, "")
}

// draws a black border based on the current position
func (w *PdfWriter) drawBorder() {
	lineTo := func(x, y Unit) {
		w.doc.LineTo(float64(x/Mm), float64(y/Mm))
	}

	w.doc.SetLineWidth(float64(w.BorderWidth / Mm))

	w.doc.MoveTo(float64((w.current.X+Mm)/Mm), float64((w.current.Y+Mm)/Mm))
	lineTo(w.current.X+Mm, w.current.Y+cardHeight-Mm)
	lineTo(w.current.X+cardWidth-Mm, w.current.Y+cardHeight-Mm)
	lineTo(w.current.X+cardWidth-Mm, w.current.Y+Mm)
	lineTo(w.current.X+Mm-w.BorderWidth/2, w.current.Y+Mm)

	w.doc.ClosePath()
	w.doc.DrawPath("D")
}

// draw a single line from point to point
func (w *PdfWriter) drawLine(from, to Point) {
	w.doc.MoveTo(float64(from.X/Mm), float64(from.Y/Mm))
	w.doc.LineTo(float64(to.X/Mm), float64(to.Y/Mm))
}

// draws crop lines based on the current position (8 per card)
func (w *PdfWriter) drawCropLines() {
	w.doc.SetLineWidth(float64(cropLineWidth))
	for i := 0; i < 2; i++ {
		swtch := Pt * Unit(float32(i))
		/* on the first iteration one line is drawn on each side. The swtch is 0
		so all the shifting by cardHeight or cardWidth is disabled. On the second
		iteration the swtch is 1, meaning the shifting is turned on. On each side
		one line is drawn, but w time shifted by cardHeight or cardWidth

		First:
		 |
			–
		–
			|

		Second
			|
		–
			 –
		 |
		*/

		// left
		w.drawLine(
			Point{
				X: w.current.X,
				Y: w.current.Y + cardHeight*swtch},
			Point{
				X: w.current.X - cropSpace,
				Y: w.current.Y + cardHeight*swtch})

		// top
		w.drawLine(
			Point{
				X: w.current.X + cardWidth*swtch,
				Y: w.current.Y + cardHeight},
			Point{
				X: w.current.X + cardWidth*swtch,
				Y: w.current.Y + cardHeight + cropSpace})

		// right
		w.drawLine(
			Point{
				X: w.current.X + cardWidth,
				Y: w.current.Y + cardHeight*swtch},
			Point{
				X: w.current.X + cardWidth + cropSpace,
				Y: w.current.Y + cardHeight*swtch})

		// top
		w.drawLine(
			Point{
				X: w.current.X + cardWidth*swtch,
				Y: w.current.Y},
			Point{
				X: w.current.X + cardWidth*swtch,
				Y: w.current.Y - cropSpace})
	}
	w.doc.DrawPath("D")
}

// closes the last page (if one exists) and creates a new empty one
func (w *PdfWriter) newPage() {
	w.doc.AddPage()
}

// number of cards that can be added to a single page (Rows x Columns)
func (w *PdfWriter) cardsPerPage() int {
	return w.Columns * w.Rows
}

// the margin from the bottom of the page (same as from top of the page)
func (w *PdfWriter) marginBottom() Unit {
	res := (w.PageHeight - (Unit(w.Rows)*cardHeight +
		(Unit(w.Rows)-1)*w.Space)) / 2
	if res > -1 {
		return res
	}
	return -1
}

// the margin from the left of the page (same as from right of the page)
func (w *PdfWriter) marginLeft() Unit {
	res := (w.PageWidth - (Unit(w.Columns)*cardWidth +
		(Unit(w.Columns)-1)*w.Space)) / 2
	if res > -1 {
		return res
	}
	return -1
}

// if PdfWriter.Border is true the default setting is returned, 0.0 otherwiese
func (w *PdfWriter) cardBorderPadding() Unit {
	if w.Border {
		return cardBorderPadding
	}
	return Unit(0.0)
}
