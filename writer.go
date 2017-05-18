package cardpdf

import (
	"io"

	"github.com/jung-kurt/gofpdf"
)

const (
	Pt                Unit = 1
	Inch              Unit = 72
	Cm                Unit = 28.35
	Mm                Unit = Cm / 10
	A4Width           Unit = 210 * Mm
	A4Height          Unit = 297 * Mm
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

// Unit is a device-independent dimensional type.  On a new canvas, this
// represents 1/72 of an inch.
type Unit float32

// Point is a 2D point.
type Point struct {
	X, Y Unit
}

// retuns a new pdf writer. This writer writes Images to pdf file
// the dimensions are definte as constants
func NewPdfWriter(writer io.WriteCloser) PdfWriter {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetDrawColor(0, 0, 0)
	pdf.SetCompression(false)
	return PdfWriter{
		doc:         pdf,
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
	PageWidth, PageHeight Unit
	BorderWidth, Space    Unit

	writer   io.WriteCloser
	doc      *gofpdf.Fpdf
	imgCount int
	current  Point
}

// writes an Image  n-times to the pdf file. With crop lines and black border
func (w *PdfWriter) WriteImage(r io.Reader, name string, count int) {

	ref := w.addImage(r, name)
	for i := 0; i < count; i++ {
		w.setCurrentPoint()
		if w.CropLines {
			w.drawCropLines()
		}
		w.drawImageReference(name, ref)
		if w.Border {
			w.drawBorder()
		}
		w.imgCount++
	}

}

// the last page is closed. The objects are encoded to
// a pdf file and written to the given writer
func (w *PdfWriter) Close() error {
	return w.doc.OutputAndClose(w.writer)
}

// based on the imageCount the lower left corner of the next image is computed
func (w *PdfWriter) setCurrentPoint() Point {
	switch {
	case w.imgCount%w.cardsPerPage() == 0:
		w.newPage()
		w.current = Point{
			X: w.marginLeft(),
			Y: w.marginBottom() +
				((Unit(w.Rows - 1)) * w.Space) +
				(Unit((w.Rows - 1)) * cardHeight)}
		return w.current

	case w.imgCount%w.Columns == 0:
		w.current = Point{
			X: w.marginLeft(),
			Y: w.current.Y - w.Space - cardHeight}
		return w.current
	}
	w.current = Point{
		X: w.current.X + w.Space + cardWidth,
		Y: w.current.Y}
	return w.current
}

// adds an Image to the pdf and returns its reference. Note that the Image
// isn't yet drawn, it's just added.
func (w *PdfWriter) addImage(r io.Reader, name string) *gofpdf.ImageInfoType {
	return w.doc.RegisterImageOptionsReader(name, gofpdf.ImageOptions{
		ImageType: "JPG",
		ReadDpi:   true,
	}, r)
}

// draws an ImageReference in the pdf. The Image to be drawn needs to be added
// to the file first.

func (w *PdfWriter) drawImageReference(name string, ref *gofpdf.ImageInfoType) {
	w.doc.ImageOptions(name,
		float64((w.current.X+w.cardBorderPadding())/Mm),
		float64((w.current.Y+w.cardBorderPadding())/Mm),
		float64((cardWidth-(2*w.cardBorderPadding()))/Mm),
		float64((cardHeight-(2*w.cardBorderPadding()))/Mm),
		false,
		gofpdf.ImageOptions{
			ImageType: "JPG",
			ReadDpi:   true,
		},
		0,
		"",
	)
	// w.page.DrawImageReference(ref, pdf.Rectangle{
	// 	Min: pdf.Point{
	// 		X: w.current.X + w.cardBorderPadding(),
	// 		Y: w.current.Y + w.cardBorderPadding()},
	// 	Max: pdf.Point{
	// 		X: w.current.X + cardWidth - w.cardBorderPadding(),
	// 		Y: w.current.Y + cardHeight - w.cardBorderPadding()},
	// })
}

// draws a black border based on the current point
func (w *PdfWriter) drawBorder() {
	w.doc.SetLineWidth(float64(w.BorderWidth / Mm))
	w.doc.MoveTo(float64((w.current.X+Mm)/Mm), float64((w.current.Y+Mm)/Mm))
	w.lineTo(w.current.X+Mm, w.current.Y+cardHeight-Mm)
	w.lineTo(w.current.X+cardWidth-Mm, w.current.Y+cardHeight-Mm)
	w.lineTo(w.current.X+cardWidth-Mm, w.current.Y+Mm)
	w.lineTo(w.current.X+Mm-w.BorderWidth/2, w.current.Y+Mm)
	w.doc.ClosePath()
	w.doc.DrawPath("D")
}

func (w *PdfWriter) lineTo(x, y Unit) {
	w.doc.LineTo(float64(x/Mm), float64(y/Mm))
}

// draw a single line
func (w *PdfWriter) drawLine(from, to Point) {
	w.doc.MoveTo(float64(from.X/Mm), float64(from.Y/Mm))
	w.lineTo(to.X, to.Y)
	w.doc.ClosePath()
	w.doc.DrawPath("D")
}

// draws the crop lines based on the current point (8 per card)
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
}

// closes the last page (if one exists) and creates a new empty one
func (w *PdfWriter) newPage() {
	w.doc.AddPage()
}

func (w *PdfWriter) cardsPerPage() int {
	return w.Columns * w.Rows
}

func (w *PdfWriter) marginBottom() Unit {
	return (w.PageHeight - (Unit(w.Rows)*cardHeight +
		(Unit(w.Rows)-1)*space)) / 2
}

func (w *PdfWriter) marginLeft() Unit {
	return (w.PageWidth - (Unit(w.Columns)*cardWidth +
		(Unit(w.Columns)-1)*space)) / 2
}

// if PdfWriter.Border is true the default setting is returned, 0.0 otherwiese
func (w *PdfWriter) cardBorderPadding() Unit {
	if w.Border {
		return cardBorderPadding
	}
	return Unit(0.0)
}
