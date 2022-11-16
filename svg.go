// Package svg provides an API for generating Scalable Vector Graphics (SVG)
package svg

// package main
//
// 	import (
// 		"github.com/wildberries-ru/svgo"
// 		"os"
// 	)
//
// 	var (
// 		width = 500
// 		height = 500
// 		canvas = svg.New(os.Stdout)
// 	)
//
// 	func main() {
// 		canvas.Start(width, height)
// 		canvas.Circle(width/2, height/2, 100)
// 		canvas.Text(width/2, height/2, "Hello, SVG",
//                   "text-anchor:middle;font-size:30px;fill:white")
// 		canvas.End()
// 	}
//

import (
	"fmt"
	"io"

	"encoding/xml"
	"strings"
)

// SVG defines the location of the generated SVG
type SVG struct {
	Writer io.Writer
}

// Offcolor defines the offset and color for gradients
type Offcolor struct {
	Offset  uint8
	Color   string
	Opacity float64
}

// Filterspec defines the specification of SVG filters
type Filterspec struct {
	In, In2, Result string
}

const (
	svgtop = `<?xml version="1.0"?>
<svg`
	svginitfmt = `%s width="%d%s" height="%d%s"`
	svgns      = `
     xmlns="http://www.w3.org/2000/svg"
     xmlns:xlink="http://www.w3.org/1999/xlink">`
	vbfmt = `viewBox="%d %d %d %d"`

	emptyclose = "/>\n"
)

// New is the SVG constructor, specifying the io.Writer where the generated SVG is written.
func New(w io.Writer) *SVG { return &SVG{w} }

func (svg *SVG) print(a ...interface{}) (n int, errno error) {
	return fmt.Fprint(svg.Writer, a...)
}

func (svg *SVG) println(a ...interface{}) (n int, errno error) {
	return fmt.Fprintln(svg.Writer, a...)
}

func (svg *SVG) printf(format string, a ...interface{}) (n int, errno error) {
	return fmt.Fprintf(svg.Writer, format, a...)
}

func (svg *SVG) genattr(ns []string) {
	for _, v := range ns {
		svg.printf("\n     %s", v)
	}
	svg.println(svgns)
}

// Structure, Metadata, Scripting, Style, Transformation, and Links

// Start begins the SVG document with the width w and height h.
// Other attributes may be optionally added, for example viewbox or additional namespaces
// Standard Reference: http://www.w3.org/TR/SVG11/struct.html#SVGElement
func (svg *SVG) Start(w int, h int, ns ...string) {
	svg.printf(svginitfmt, svgtop, w, "", h, "")
	svg.genattr(ns)
}

// Startunit begins the SVG document, with width and height in the specified units
// Other attributes may be optionally added, for example viewbox or additional namespaces
func (svg *SVG) Startunit(w int, h int, unit string, ns ...string) {
	svg.printf(svginitfmt, svgtop, w, unit, h, unit)
	svg.genattr(ns)
}

// Startpercent begins the SVG document, with width and height as percentages
// Other attributes may be optionally added, for example viewbox or additional namespaces
func (svg *SVG) Startpercent(w int, h int, ns ...string) {
	svg.printf(svginitfmt, svgtop, w, "%", h, "%")
	svg.genattr(ns)
}

// Startview begins the SVG document, with the specified width, height, and viewbox
// Other attributes may be optionally added, for example viewbox or additional namespaces
func (svg *SVG) Startview(w, h, minx, miny, vw, vh int) {
	svg.Start(w, h, fmt.Sprintf(vbfmt, minx, miny, vw, vh))
}

// StartviewUnit begins the SVG document with the specified width, height, and unit
func (svg *SVG) StartviewUnit(w, h int, unit string, minx, miny, vw, vh int) {
	svg.Startunit(w, h, unit, fmt.Sprintf(vbfmt, minx, miny, vw, vh))
}

// Startraw begins the SVG document, passing arbitrary attributes
func (svg *SVG) Startraw(ns ...string) {
	svg.printf(svgtop)
	svg.genattr(ns)
}

// End the SVG document
func (svg *SVG) End() { svg.println("</svg>") }

// linkembed defines an element with a specified type,
// (for example "application/javascript", or "text/css").
// if the first variadic argument is a link, use only the link reference.
// Otherwise, treat those arguments as the text of the script (marked up as CDATA).
// if no data is specified, just close the element
func (svg *SVG) linkembed(tag string, scriptype string, data ...string) {
	svg.printf(`<%s type="%s"`, tag, scriptype)
	switch {
	case len(data) == 1 && islink(data[0]):
		svg.printf(" %s/>\n", href(data[0]))

	case len(data) > 0:
		svg.printf(">\n<![CDATA[\n")
		for _, v := range data {
			svg.println(v)
		}
		svg.printf("]]>\n</%s>\n", tag)

	default:
		svg.println(`/>`)
	}
}

// Script defines a script with a specified type, (for example "application/javascript").
func (svg *SVG) Script(scriptype string, data ...string) {
	svg.linkembed("script", scriptype, data...)
}

// Style defines the specified style (for example "text/css")
func (svg *SVG) Style(scriptype string, data ...string) {
	svg.linkembed("style", scriptype, data...)
}

// Gstyle begins a group, with the specified style.
// Standard Reference: http://www.w3.org/TR/SVG11/struct.html#GElement
func (svg *SVG) Gstyle(s string) { svg.println(group("style", s)) }

// Gtransform begins a group, with the specified transform
// Standard Reference: http://www.w3.org/TR/SVG11/coords.html#TransformAttribute
func (svg *SVG) Gtransform(s string) {
	svg.printf(`<g %s="%s">`, "transform", s)
	svg.println("")
}

// Translate begins coordinate translation, end with Gend()
// Standard Reference: http://www.w3.org/TR/SVG11/coords.html#TransformAttribute
func (svg *SVG) Translate(x, y int) { svg.Gtransform(translate(x, y)) }

// Scale scales the coordinate system by n, end with Gend()
// Standard Reference: http://www.w3.org/TR/SVG11/coords.html#TransformAttribute
func (svg *SVG) Scale(n float64) { svg.Gtransform(scale(n)) }

// ScaleXY scales the coordinate system by dx and dy, end with Gend()
// Standard Reference: http://www.w3.org/TR/SVG11/coords.html#TransformAttribute
func (svg *SVG) ScaleXY(dx, dy float64) { svg.Gtransform(scaleXY(dx, dy)) }

// SkewX skews the x coordinate system by angle a, end with Gend()
// Standard Reference: http://www.w3.org/TR/SVG11/coords.html#TransformAttribute
func (svg *SVG) SkewX(a float64) { svg.Gtransform(skewX(a)) }

// SkewY skews the y coordinate system by angle a, end with Gend()
// Standard Reference: http://www.w3.org/TR/SVG11/coords.html#TransformAttribute
func (svg *SVG) SkewY(a float64) { svg.Gtransform(skewY(a)) }

// SkewXY skews x and y coordinates by ax, ay respectively, end with Gend()
// Standard Reference: http://www.w3.org/TR/SVG11/coords.html#TransformAttribute
func (svg *SVG) SkewXY(ax, ay float64) { svg.Gtransform(skewX(ax) + " " + skewY(ay)) }

// Rotate rotates the coordinate system by r degrees, end with Gend()
// Standard Reference: http://www.w3.org/TR/SVG11/coords.html#TransformAttribute
func (svg *SVG) Rotate(r float64) { svg.Gtransform(rotate(r)) }

// TranslateRotate translates the coordinate system to (x,y), then rotates to r degrees, end with Gend()
func (svg *SVG) TranslateRotate(x, y int, r float64) {
	svg.Gtransform(translate(x, y) + " " + rotate(r))
}

// RotateTranslate rotates the coordinate system r degrees, then translates to (x,y), end with Gend()
func (svg *SVG) RotateTranslate(x, y int, r float64) {
	svg.Gtransform(rotate(r) + " " + translate(x, y))
}

// Group begins a group with arbitrary attributes
func (svg *SVG) Group(s ...string) { svg.printf("<g %s\n", endstyle(s, `>`)) }

// Gid begins a group, with the specified id
func (svg *SVG) Gid(s string) {
	svg.print(`<g id="`)
	xml.Escape(svg.Writer, []byte(s))
	svg.println(`">`)
}

// Gend ends a group (must be paired with Gsttyle, Gtransform, Gid).
func (svg *SVG) Gend() { svg.println(`</g>`) }

// ClipPath defines a clip path
func (svg *SVG) ClipPath(s ...string) { svg.printf(`<clipPath %s`, endstyle(s, `>`)) }

// ClipEnd ends a ClipPath
func (svg *SVG) ClipEnd() {
	svg.println(`</clipPath>`)
}

// Def begins a defintion block.
// Standard Reference: http://www.w3.org/TR/SVG11/struct.html#DefsElement
func (svg *SVG) Def() { svg.println(`<defs>`) }

// DefEnd ends a defintion block.
func (svg *SVG) DefEnd() { svg.println(`</defs>`) }

// Marker defines a marker
// Standard reference: http://www.w3.org/TR/SVG11/painting.html#MarkerElement
func (svg *SVG) Marker(id string, x, y, width, height int, s ...string) {
	svg.printf(`<marker id="%s" refX="%d" refY="%d" markerWidth="%d" markerHeight="%d" %s`,
		id, x, y, width, height, endstyle(s, ">\n"))
}

// MarkerEnd ends a marker
func (svg *SVG) MarkerEnd() { svg.println(`</marker>`) }

// Pattern defines a pattern with the specified dimensions.
// The putype can be either "user" or "obj", which sets the patternUnits
// attribute to be either userSpaceOnUse or objectBoundingBox
// Standard reference: http://www.w3.org/TR/SVG11/pservers.html#Patterns
func (svg *SVG) Pattern(id string, x, y, width, height int, putype string, s ...string) {
	puattr := "userSpaceOnUse"
	if putype != "user" {
		puattr = "objectBoundingBox"
	}
	svg.printf(`<pattern id="%s" x="%d" y="%d" width="%d" height="%d" patternUnits="%s" %s`,
		id, x, y, width, height, puattr, endstyle(s, ">\n"))
}

// PatternEnd ends a marker
func (svg *SVG) PatternEnd() { svg.println(`</pattern>`) }

// Desc specified the text of the description tag.
// Standard Reference: http://www.w3.org/TR/SVG11/struct.html#DescElement
func (svg *SVG) Desc(s string) { svg.tt("desc", s) }

// Title specified the text of the title tag.
// Standard Reference: http://www.w3.org/TR/SVG11/struct.html#TitleElement
func (svg *SVG) Title(s string) { svg.tt("title", s) }

// Link begins a link named "name", with the specified title.
// Standard Reference: http://www.w3.org/TR/SVG11/linking.html#Links
func (svg *SVG) Link(href string, title string) {
	svg.printf("<a xlink:href=\"%s\" xlink:title=\"", href)
	xml.Escape(svg.Writer, []byte(title))
	svg.println("\">")
}

// LinkEnd ends a link.
func (svg *SVG) LinkEnd() { svg.println(`</a>`) }

// Use places the object referenced at link at the location x, y, with optional style.
// Standard Reference: http://www.w3.org/TR/SVG11/struct.html#UseElement
func (svg *SVG) Use(x int, y int, link string, s ...string) {
	svg.printf(`<use %s %s %s`, loc(x, y), href(link), endstyle(s, emptyclose))
}

// Mask creates a mask with a specified id, dimension, and optional style.
func (svg *SVG) Mask(id string, x int, y int, w int, h int, s ...string) {
	svg.printf(`<mask id="%s" x="%d" y="%d" width="%d" height="%d" %s`, id, x, y, w, h, endstyle(s, `>`))
}

// MaskEnd ends a Mask.
func (svg *SVG) MaskEnd() { svg.println(`</mask>`) }

// Shapes

// Circle centered at x,y, with radius r, with optional style.
// Standard Reference: http://www.w3.org/TR/SVG11/shapes.html#CircleElement
func (svg *SVG) Circle(x int, y int, r int, s ...string) {
	svg.printf(`<circle cx="%d" cy="%d" r="%d" %s`, x, y, r, endstyle(s, emptyclose))
}

// Ellipse centered at x,y, centered at x,y with radii w, and h, with optional style.
// Standard Reference: http://www.w3.org/TR/SVG11/shapes.html#EllipseElement
func (svg *SVG) Ellipse(x int, y int, w int, h int, s ...string) {
	svg.printf(`<ellipse cx="%d" cy="%d" rx="%d" ry="%d" %s`,
		x, y, w, h, endstyle(s, emptyclose))
}

// Polygon draws a series of line segments using an array of x, y coordinates, with optional style.
// Standard Reference: http://www.w3.org/TR/SVG11/shapes.html#PolygonElement
func (svg *SVG) Polygon(x []int, y []int, s ...string) {
	svg.poly(x, y, "polygon", s...)
}

// Rect draws a rectangle with upper left-hand corner at x,y, with width w, and height h, with optional style
// Standard Reference: http://www.w3.org/TR/SVG11/shapes.html#RectElement
func (svg *SVG) Rect(x int, y int, w int, h int, s ...string) {
	// svg.printf(`<rect %s %s`, dim(x, y, w, h), endstyle(s, emptyclose))

	svg.printf(`<rect `)
	svg.printf(`x="%d" y="%d" width="%d" height="%d"`, x, y, w, h)

	if len(s) > 0 {
		for i := 0; i < len(s); i++ {
			if strings.Index(s[i], "=") > 0 {
				svg.printf(` %s`, s[i])
			} else {
				svg.printf(` style="%s"`, s[i])
			}
		}
	}

	svg.printf(` %s`, emptyclose)
}

// CenterRect draws a rectangle with its center at x,y, with width w, and height h, with optional style
func (svg *SVG) CenterRect(x int, y int, w int, h int, s ...string) {
	svg.Rect(x-(w/2), y-(h/2), w, h, s...)
}

// Roundrect draws a rounded rectangle with upper the left-hand corner at x,y,
// with width w, and height h. The radii for the rounded portion
// are specified by rx (width), and ry (height).
// Style is optional.
// Standard Reference: http://www.w3.org/TR/SVG11/shapes.html#RectElement
func (svg *SVG) Roundrect(x int, y int, w int, h int, rx int, ry int, s ...string) {
	svg.printf(`<rect %s rx="%d" ry="%d" %s`, dim(x, y, w, h), rx, ry, endstyle(s, emptyclose))
}

// Square draws a square with upper left corner at x,y with sides of length l, with optional style.
func (svg *SVG) Square(x int, y int, l int, s ...string) {
	svg.Rect(x, y, l, l, s...)
}

// Paths

// Path draws an arbitrary path, the caller is responsible for structuring the path data
func (svg *SVG) Path(d string, s ...string) {
	svg.printf(`<path d="%s" %s`, d, endstyle(s, emptyclose))
}

// Arc draws an elliptical arc, with optional style, beginning coordinate at sx,sy, ending coordinate at ex, ey
// width and height of the arc are specified by ax, ay, the x axis rotation is r
// if sweep is true, then the arc will be drawn in a "positive-angle" direction (clockwise), if false,
// the arc is drawn counterclockwise.
// if large is true, the arc sweep angle is greater than or equal to 180 degrees,
// otherwise the arc sweep is less than 180 degrees
// http://www.w3.org/TR/SVG11/paths.html#PathDataEllipticalArcCommands
func (svg *SVG) Arc(sx int, sy int, ax int, ay int, r int, large bool, sweep bool, ex int, ey int, s ...string) {
	svg.printf(`%s A%s %d %s %s %s" %s`,
		ptag(sx, sy), coord(ax, ay), r, onezero(large), onezero(sweep), coord(ex, ey), endstyle(s, emptyclose))
}

// Bezier draws a cubic bezier curve, with optional style, beginning at sx,sy, ending at ex,ey
// with control points at cx,cy and px,py.
// Standard Reference: http://www.w3.org/TR/SVG11/paths.html#PathDataCubicBezierCommands
func (svg *SVG) Bezier(sx int, sy int, cx int, cy int, px int, py int, ex int, ey int, s ...string) {
	svg.printf(`%s C%s %s %s" %s`,
		ptag(sx, sy), coord(cx, cy), coord(px, py), coord(ex, ey), endstyle(s, emptyclose))
}

// Qbez draws a quadratic bezier curver, with optional style
// beginning at sx,sy, ending at ex, sy with control points at cx, cy
// Standard Reference: http://www.w3.org/TR/SVG11/paths.html#PathDataQuadraticBezierCommands
func (svg *SVG) Qbez(sx int, sy int, cx int, cy int, ex int, ey int, s ...string) {
	svg.printf(`%s Q%s %s" %s`,
		ptag(sx, sy), coord(cx, cy), coord(ex, ey), endstyle(s, emptyclose))
}

// Qbezier draws a Quadratic Bezier curve, with optional style, beginning at sx, sy, ending at tx,ty
// with control points are at cx,cy, ex,ey.
// Standard Reference: http://www.w3.org/TR/SVG11/paths.html#PathDataQuadraticBezierCommands
func (svg *SVG) Qbezier(sx int, sy int, cx int, cy int, ex int, ey int, tx int, ty int, s ...string) {
	svg.printf(`%s Q%s %s T%s" %s`,
		ptag(sx, sy), coord(cx, cy), coord(ex, ey), coord(tx, ty), endstyle(s, emptyclose))
}

// Lines

// Line draws a straight line between two points, with optional style.
// Standard Reference: http://www.w3.org/TR/SVG11/shapes.html#LineElement
func (svg *SVG) Line(x1 int, y1 int, x2 int, y2 int, s ...string) {
	svg.printf(`<line x1="%d" y1="%d" x2="%d" y2="%d" %s`, x1, y1, x2, y2, endstyle(s, emptyclose))
}

// Polyline draws connected lines between coordinates, with optional style.
// Standard Reference: http://www.w3.org/TR/SVG11/shapes.html#PolylineElement
func (svg *SVG) Polyline(x []int, y []int, s ...string) {
	svg.poly(x, y, "polyline", s...)
}

// Image places at x,y (upper left hand corner), the image with
// width w, and height h, referenced at link, with optional style.
// Standard Reference: http://www.w3.org/TR/SVG11/struct.html#ImageElement
func (svg *SVG) Image(x int, y int, w int, h int, link string, s ...string) {
	svg.printf(`<image %s %s %s`, dim(x, y, w, h), href(link), endstyle(s, emptyclose))
}

// Text places the specified text, t at x,y according to the style specified in s
// Standard Reference: http://www.w3.org/TR/SVG11/text.html#TextElement
func (svg *SVG) Text(x int, y int, t string, s ...string) {
	svg.printf(`<text %s %s`, loc(x, y), endstyle(s, ">"))
	xml.Escape(svg.Writer, []byte(t))
	svg.println(`</text>`)
}

// Textspan begins text, assuming a tspan will be included, end with TextEnd()
// Standard Reference: https://www.w3.org/TR/SVG11/text.html#TSpanElement
func (svg *SVG) Textspan(x int, y int, t string, s ...string) {
	svg.printf(`<text %s %s`, loc(x, y), endstyle(s, ">"))
	xml.Escape(svg.Writer, []byte(t))
}

// Span makes styled spanned text, should be proceeded by Textspan
// Standard Reference: https://www.w3.org/TR/SVG11/text.html#TSpanElement
func (svg *SVG) Span(t string, s ...string) {
	if len(s) == 0 {
		xml.Escape(svg.Writer, []byte(t))
		return
	}
	svg.printf(`<tspan %s`, endstyle(s, ">"))
	xml.Escape(svg.Writer, []byte(t))
	svg.printf(`</tspan>`)
}

// TextEnd ends spanned text
// Standard Reference: https://www.w3.org/TR/SVG11/text.html#TSpanElement
func (svg *SVG) TextEnd() {
	svg.println(`</text>`)
}

// Textpath places text optionally styled text along a previously defined path
// Standard Reference: http://www.w3.org/TR/SVG11/text.html#TextPathElement
func (svg *SVG) Textpath(t string, pathid string, s ...string) {
	svg.printf("<text %s<textPath xlink:href=\"%s\">", endstyle(s, ">"), pathid)
	xml.Escape(svg.Writer, []byte(t))
	svg.println(`</textPath></text>`)
}

// Textlines places a series of lines of text starting at x,y, at the specified size, fill, and alignment.
// Each line is spaced according to the spacing argument
func (svg *SVG) Textlines(x, y int, s []string, size, spacing int, fill, align string) {
	svg.Gstyle(fmt.Sprintf("font-size:%dpx;fill:%s;text-anchor:%s", size, fill, align))
	for _, t := range s {
		svg.Text(x, y, t)
		y += spacing
	}
	svg.Gend()
}

// Colors

// RGB specifies a fill color in terms of a (r)ed, (g)reen, (b)lue triple.
// Standard reference: http://www.w3.org/TR/css3-color/
func (svg *SVG) RGB(r int, g int, b int) string {
	return fmt.Sprintf(`fill:rgb(%d,%d,%d)`, r, g, b)
}

// RGBA specifies a fill color in terms of a (r)ed, (g)reen, (b)lue triple and opacity.
func (svg *SVG) RGBA(r int, g int, b int, a float64) string {
	return fmt.Sprintf(`fill-opacity:%.2f; %s`, a, svg.RGB(r, g, b))
}

// Gradients

// LinearGradient constructs a linear color gradient identified by id,
// along the vector defined by (x1,y1), and (x2,y2).
// The stop color sequence defined in sc. Coordinates are expressed as percentages.
func (svg *SVG) LinearGradient(id string, x1, y1, x2, y2 uint8, sc []Offcolor) {
	svg.printf("<linearGradient id=\"%s\" x1=\"%d%%\" y1=\"%d%%\" x2=\"%d%%\" y2=\"%d%%\">\n",
		id, pct(x1), pct(y1), pct(x2), pct(y2))
	svg.stopcolor(sc)
	svg.println("</linearGradient>")
}

// RadialGradient constructs a radial color gradient identified by id,
// centered at (cx,cy), with a radius of r.
// (fx, fy) define the location of the focal point of the light source.
// The stop color sequence defined in sc.
// Coordinates are expressed as percentages.
func (svg *SVG) RadialGradient(id string, cx, cy, r, fx, fy uint8, sc []Offcolor) {
	svg.printf("<radialGradient id=\"%s\" cx=\"%d%%\" cy=\"%d%%\" r=\"%d%%\" fx=\"%d%%\" fy=\"%d%%\">\n",
		id, pct(cx), pct(cy), pct(r), pct(fx), pct(fy))
	svg.stopcolor(sc)
	svg.println("</radialGradient>")
}

// stopcolor is a utility function used by the gradient functions
// to define a sequence of offsets (expressed as percentages) and colors
func (svg *SVG) stopcolor(oc []Offcolor) {
	for _, v := range oc {
		svg.printf("<stop offset=\"%d%%\" stop-color=\"%s\" stop-opacity=\"%.2f\"/>\n",
			pct(v.Offset), v.Color, v.Opacity)
	}
}

// Filter Effects:
// Most functions have common attributes (in, in2, result) defined in type Filterspec
// used as a common first argument.

// Filter begins a filter set
// Standard reference: http://www.w3.org/TR/SVG11/filters.html#FilterElement
func (svg *SVG) Filter(id string, s ...string) {
	svg.printf(`<filter id="%s" %s`, id, endstyle(s, ">\n"))
}

// Fend ends a filter set
// Standard reference: http://www.w3.org/TR/SVG11/filters.html#FilterElement
func (svg *SVG) Fend() {
	svg.println(`</filter>`)
}

// FeBlend specifies a Blend filter primitive
// Standard reference: http://www.w3.org/TR/SVG11/filters.html#feBlendElement
func (svg *SVG) FeBlend(fs Filterspec, mode string, s ...string) {
	switch mode {
	case "normal", "multiply", "screen", "darken", "lighten":
		break
	default:
		mode = "normal"
	}
	svg.printf(`<feBlend %s mode="%s" %s`,
		fsattr(fs), mode, endstyle(s, emptyclose))
}

// FeColorMatrix specifies a color matrix filter primitive, with matrix values
// Standard reference: http://www.w3.org/TR/SVG11/filters.html#feColorMatrixElement
func (svg *SVG) FeColorMatrix(fs Filterspec, values [20]float64, s ...string) {
	svg.printf(`<feColorMatrix %s type="matrix" values="`, fsattr(fs))
	for _, v := range values {
		svg.printf(`%g `, v)
	}
	svg.printf(`" %s`, endstyle(s, emptyclose))
}

// FeColorMatrixHue specifies a color matrix filter primitive, with hue rotation values
// Standard reference: http://www.w3.org/TR/SVG11/filters.html#feColorMatrixElement
func (svg *SVG) FeColorMatrixHue(fs Filterspec, value float64, s ...string) {
	if value < -360 || value > 360 {
		value = 0
	}
	svg.printf(`<feColorMatrix %s type="hueRotate" values="%g" %s`,
		fsattr(fs), value, endstyle(s, emptyclose))
}

// FeColorMatrixSaturate specifies a color matrix filter primitive, with saturation values
// Standard reference: http://www.w3.org/TR/SVG11/filters.html#feColorMatrixElement
func (svg *SVG) FeColorMatrixSaturate(fs Filterspec, value float64, s ...string) {
	if value < 0 || value > 1 {
		value = 1
	}
	svg.printf(`<feColorMatrix %s type="saturate" values="%g" %s`,
		fsattr(fs), value, endstyle(s, emptyclose))
}

// FeColorMatrixLuminence specifies a color matrix filter primitive, with luminence values
// Standard reference: http://www.w3.org/TR/SVG11/filters.html#feColorMatrixElement
func (svg *SVG) FeColorMatrixLuminence(fs Filterspec, s ...string) {
	svg.printf(`<feColorMatrix %s type="luminenceToAlpha" %s`,
		fsattr(fs), endstyle(s, emptyclose))
}

// FeComponentTransfer begins a feComponent filter element
// Standard reference: http://www.w3.org/TR/SVG11/filters.html#feComponentTransferElement
func (svg *SVG) FeComponentTransfer() {
	svg.println(`<feComponentTransfer>`)
}

// FeCompEnd ends a feComponent filter element
// Standard reference: http://www.w3.org/TR/SVG11/filters.html#feComponentTransferElement
func (svg *SVG) FeCompEnd() {
	svg.println(`</feComponentTransfer>`)
}

// FeComposite specifies a feComposite filter primitive
// Standard reference: http://www.w3.org/TR/SVG11/filters.html#feCompositeElement
func (svg *SVG) FeComposite(fs Filterspec, operator string, k1, k2, k3, k4 int, s ...string) {
	switch operator {
	case "over", "in", "out", "atop", "xor", "arithmetic":
		break
	default:
		operator = "over"
	}
	svg.printf(`<feComposite %s operator="%s" k1="%d" k2="%d" k3="%d" k4="%d" %s`,
		fsattr(fs), operator, k1, k2, k3, k4, endstyle(s, emptyclose))
}

// FeConvolveMatrix specifies a feConvolveMatrix filter primitive
// Standard referencd: http://www.w3.org/TR/SVG11/filters.html#feConvolveMatrixElement
func (svg *SVG) FeConvolveMatrix(fs Filterspec, matrix [9]int, s ...string) {
	svg.printf(`<feConvolveMatrix %s kernelMatrix="%d %d %d %d %d %d %d %d %d" %s`,
		fsattr(fs),
		matrix[0], matrix[1], matrix[2],
		matrix[3], matrix[4], matrix[5],
		matrix[6], matrix[7], matrix[8], endstyle(s, emptyclose))
}

// FeDiffuseLighting specifies a diffuse lighting filter primitive,
// a container for light source elements, end with DiffuseEnd()
// Standard reference: http://www.w3.org/TR/SVG11/filters.html#feComponentTransferElement
func (svg *SVG) FeDiffuseLighting(fs Filterspec, scale, constant float64, s ...string) {
	svg.printf(`<feDiffuseLighting %s surfaceScale="%g" diffuseConstant="%g" %s`,
		fsattr(fs), scale, constant, endstyle(s, `>`))
}

// FeDiffEnd ends a diffuse lighting filter primitive container
// Standard reference: http://www.w3.org/TR/SVG11/filters.html#feDiffuseLightingElement
func (svg *SVG) FeDiffEnd() {
	svg.println(`</feDiffuseLighting>`)
}

// FeDisplacementMap specifies a feDisplacementMap filter primitive
// Standard reference: http://www.w3.org/TR/SVG11/filters.html#feDisplacementMapElement
func (svg *SVG) FeDisplacementMap(fs Filterspec, scale float64, xchannel, ychannel string, s ...string) {
	svg.printf(`<feDisplacementMap %s scale="%g" xChannelSelector="%s" yChannelSelector="%s" %s`,
		fsattr(fs), scale, imgchannel(xchannel), ychannel, endstyle(s, emptyclose))
}

// FeDistantLight specifies a feDistantLight filter primitive
// Standard reference: http://www.w3.org/TR/SVG11/filters.html#feDistantLightElement
func (svg *SVG) FeDistantLight(fs Filterspec, azimuth, elevation float64, s ...string) {
	svg.printf(`<feDistantLight %s azimuth="%g" elevation="%g" %s`,
		fsattr(fs), azimuth, elevation, endstyle(s, emptyclose))
}

// FeFlood specifies a flood filter primitive
// Standard reference: http://www.w3.org/TR/SVG11/filters.html#feFloodElement
func (svg *SVG) FeFlood(fs Filterspec, color string, opacity float64, s ...string) {
	svg.printf(`<feFlood %s flood-color="%s" flood-opacity="%g" %s`,
		fsattr(fs), color, opacity, endstyle(s, emptyclose))
}

// FeFunc{linear|Gamma|Table|Discrete} specify various types of feFunc{R|G|B|A} filter primitives
// Standard reference: http://www.w3.org/TR/SVG11/filters.html#feComponentTransferElement

// FeFuncLinear specifies a linear style function for the feFunc{R|G|B|A} filter element
// Standard reference: http://www.w3.org/TR/SVG11/filters.html#feComponentTransferElement
func (svg *SVG) FeFuncLinear(channel string, slope, intercept float64) {
	svg.printf(`<feFunc%s type="linear" slope="%g" intercept="%g"%s`,
		imgchannel(channel), slope, intercept, emptyclose)
}

// FeFuncGamma specifies the curve values for gamma correction for the feFunc{R|G|B|A} filter element
// Standard reference: http://www.w3.org/TR/SVG11/filters.html#feComponentTransferElement
func (svg *SVG) FeFuncGamma(channel string, amplitude, exponent, offset float64) {
	svg.printf(`<feFunc%s type="gamma" amplitude="%g" exponent="%g" offset="%g"%s`,
		imgchannel(channel), amplitude, exponent, offset, emptyclose)
}

// FeFuncTable specifies the table of values for the feFunc{R|G|B|A} filter element
// Standard reference: http://www.w3.org/TR/SVG11/filters.html#feComponentTransferElement
func (svg *SVG) FeFuncTable(channel string, tv []float64) {
	svg.printf(`<feFunc%s type="table"`, imgchannel(channel))
	svg.tablevalues(`tableValues`, tv)
}

// FeFuncDiscrete specifies the discrete values for the feFunc{R|G|B|A} filter element
// Standard reference: http://www.w3.org/TR/SVG11/filters.html#feComponentTransferElement
func (svg *SVG) FeFuncDiscrete(channel string, tv []float64) {
	svg.printf(`<feFunc%s type="discrete"`, imgchannel(channel))
	svg.tablevalues(`tableValues`, tv)
}

// FeGaussianBlur specifies a Gaussian Blur filter primitive
// Standard reference: http://www.w3.org/TR/SVG11/filters.html#feGaussianBlurElement
func (svg *SVG) FeGaussianBlur(fs Filterspec, stdx, stdy float64, s ...string) {
	if stdx < 0 {
		stdx = 0
	}
	if stdy < 0 {
		stdy = 0
	}
	svg.printf(`<feGaussianBlur %s stdDeviation="%g %g" %s`,
		fsattr(fs), stdx, stdy, endstyle(s, emptyclose))
}

// FeImage specifies a feImage filter primitive
// Standard reference: http://www.w3.org/TR/SVG11/filters.html#feImageElement
func (svg *SVG) FeImage(href string, result string, s ...string) {
	svg.printf(`<feImage xlink:href="%s" result="%s" %s`,
		href, result, endstyle(s, emptyclose))
}

// FeMerge specifies a feMerge filter primitive, containing feMerge elements
// Standard reference: http://www.w3.org/TR/SVG11/filters.html#feMergeElement
func (svg *SVG) FeMerge(nodes []string, s ...string) {
	svg.println(`<feMerge>`)
	for _, n := range nodes {
		svg.printf("<feMergeNode in=\"%s\"/>\n", n)
	}
	svg.println(`</feMerge>`)
}

// FeMorphology specifies a feMorphologyLight filter primitive
// Standard reference: http://www.w3.org/TR/SVG11/filters.html#feMorphologyElement
func (svg *SVG) FeMorphology(fs Filterspec, operator string, xradius, yradius float64, s ...string) {
	switch operator {
	case "erode", "dilate":
		break
	default:
		operator = "erode"
	}
	svg.printf(`<feMorphology %s operator="%s" radius="%g %g" %s`,
		fsattr(fs), operator, xradius, yradius, endstyle(s, emptyclose))
}

// FeOffset specifies the feOffset filter primitive
// Standard reference: http://www.w3.org/TR/SVG11/filters.html#feOffsetElement
func (svg *SVG) FeOffset(fs Filterspec, dx, dy int, s ...string) {
	svg.printf(`<feOffset %s dx="%d" dy="%d" %s`,
		fsattr(fs), dx, dy, endstyle(s, emptyclose))
}

// FePointLight specifies a fePpointLight filter primitive
// Standard reference: http://www.w3.org/TR/SVG11/filters.html#fePointLightElement
func (svg *SVG) FePointLight(x, y, z float64, s ...string) {
	svg.printf(`<fePointLight x="%g" y="%g" z="%g" %s`,
		x, y, z, endstyle(s, emptyclose))
}

// FeSpecularLighting specifies a specular lighting filter primitive,
// a container for light source elements, end with SpecularEnd()
// Standard reference: http://www.w3.org/TR/SVG11/filters.html#feSpecularLightingElement
func (svg *SVG) FeSpecularLighting(fs Filterspec, scale, constant float64, exponent int, color string, s ...string) {
	svg.printf(`<feSpecularLighting %s surfaceScale="%g" specularConstant="%g" specularExponent="%d" lighting-color="%s" %s`,
		fsattr(fs), scale, constant, exponent, color, endstyle(s, ">\n"))
}

// FeSpecEnd ends a specular lighting filter primitive container
// Standard reference: http://www.w3.org/TR/SVG11/filters.html#feSpecularLightingElement
func (svg *SVG) FeSpecEnd() {
	svg.println(`</feSpecularLighting>`)
}

// FeSpotLight specifies a feSpotLight filter primitive
// Standard reference: http://www.w3.org/TR/SVG11/filters.html#feSpotLightElement
func (svg *SVG) FeSpotLight(fs Filterspec, x, y, z, px, py, pz float64, s ...string) {
	svg.printf(`<feSpotLight %s x="%g" y="%g" z="%g" pointsAtX="%g" pointsAtY="%g" pointsAtZ="%g" %s`,
		fsattr(fs), x, y, z, px, py, pz, endstyle(s, emptyclose))
}

// FeTile specifies the tile utility filter primitive
// Standard reference: http://www.w3.org/TR/SVG11/filters.html#feTileElement
func (svg *SVG) FeTile(fs Filterspec, in string, s ...string) {
	svg.printf(`<feTile %s %s`, fsattr(fs), endstyle(s, emptyclose))
}

// FeTurbulence specifies a turbulence filter primitive
// Standard reference: http://www.w3.org/TR/SVG11/filters.html#feTurbulenceElement
func (svg *SVG) FeTurbulence(fs Filterspec, ftype string, bfx, bfy float64, octaves int, seed int64, stitch bool, s ...string) {
	if bfx < 0 || bfx > 1 {
		bfx = 0
	}
	if bfy < 0 || bfy > 1 {
		bfy = 0
	}
	switch ftype[0:1] {
	case "f", "F":
		ftype = "fractalNoise"
	case "t", "T":
		ftype = "turbulence"
	default:
		ftype = "turbulence"
	}

	var ss string
	if stitch {
		ss = "stitch"
	} else {
		ss = "noStitch"
	}
	svg.printf(`<feTurbulence %s type="%s" baseFrequency="%.2f %.2f" numOctaves="%d" seed="%d" stitchTiles="%s" %s`,
		fsattr(fs), ftype, bfx, bfy, octaves, seed, ss, endstyle(s, emptyclose))
}

// Filter Effects convenience functions, modeled after CSS versions

// Blur emulates the CSS blur filter
func (svg *SVG) Blur(p float64) {
	svg.FeGaussianBlur(Filterspec{}, p, p)
}

// Brightness emulates the CSS brightness filter
func (svg *SVG) Brightness(p float64) {
	svg.FeComponentTransfer()
	svg.FeFuncLinear("R", p, 0)
	svg.FeFuncLinear("G", p, 0)
	svg.FeFuncLinear("B", p, 0)
	svg.FeCompEnd()
}

// Contrast emulates the CSS contrast filter
// func (svg *SVG) Contrast(p float64) {
// }

// Dropshadow emulates the CSS dropshadow filter
// func (svg *SVG) Dropshadow(p float64) {
// }

// Grayscale eumulates the CSS grayscale filter
func (svg *SVG) Grayscale() {
	svg.FeColorMatrixSaturate(Filterspec{}, 0)
}

// HueRotate eumulates the CSS huerotate filter
func (svg *SVG) HueRotate(a float64) {
	svg.FeColorMatrixHue(Filterspec{}, a)
}

// Invert eumulates the CSS invert filter
func (svg *SVG) Invert() {
	svg.FeComponentTransfer()
	svg.FeFuncTable("R", []float64{1, 0})
	svg.FeFuncTable("G", []float64{1, 0})
	svg.FeFuncTable("B", []float64{1, 0})
	svg.FeCompEnd()
}

// Saturate eumulates the CSS saturate filter
func (svg *SVG) Saturate(p float64) {
	svg.FeColorMatrixSaturate(Filterspec{}, p)
}

// Sepia applies a sepia tone, emulating the CSS sepia filter
func (svg *SVG) Sepia() {
	var sepiamatrix = [20]float64{
		0.280, 0.450, 0.05, 0, 0,
		0.140, 0.390, 0.04, 0, 0,
		0.080, 0.280, 0.03, 0, 0,
		0, 0, 0, 1, 0,
	}
	svg.FeColorMatrix(Filterspec{}, sepiamatrix)
}

// Animation

// Animate animates the specified link, using the specified attribute
// The animation starts at coordinate from, terminates at to, and repeats as specified
func (svg *SVG) Animate(link, attr string, from, to int, duration float64, repeat int, s ...string) {
	svg.printf(`<animate %s attributeName="%s" from="%d" to="%d" dur="%gs" repeatCount="%s" %s`,
		href(link), attr, from, to, duration, repeatString(repeat), endstyle(s, emptyclose))
}

// AnimateMotion animates the referenced object along the specified path
func (svg *SVG) AnimateMotion(link, path string, duration float64, repeat int, s ...string) {
	svg.printf(`<animateMotion %s dur="%gs" repeatCount="%s" %s<mpath %s/></animateMotion>
`, href(link), duration, repeatString(repeat), endstyle(s, ">"), href(path))
}

// AnimateTransform animates in the context of SVG transformations
func (svg *SVG) AnimateTransform(link, ttype, from, to string, duration float64, repeat int, s ...string) {
	svg.printf(`<animateTransform %s attributeName="transform" type="%s" from="%s" to="%s" dur="%gs" repeatCount="%s" %s`,
		href(link), ttype, from, to, duration, repeatString(repeat), endstyle(s, emptyclose))
}

// AnimateTranslate animates the translation transformation
func (svg *SVG) AnimateTranslate(link string, fx, fy, tx, ty int, duration float64, repeat int, s ...string) {
	svg.AnimateTransform(link, "translate", coordpair(fx, fy), coordpair(tx, ty), duration, repeat, s...)
}

// AnimateRotate animates the rotation transformation
func (svg *SVG) AnimateRotate(link string, fs, fc, fe, ts, tc, te int, duration float64, repeat int, s ...string) {
	svg.AnimateTransform(link, "rotate", sce(fs, fc, fe), sce(ts, tc, te), duration, repeat, s...)
}

// AnimateScale animates the scale transformation
func (svg *SVG) AnimateScale(link string, from, to, duration float64, repeat int, s ...string) {
	svg.AnimateTransform(link, "scale", fmt.Sprintf("%g", from), fmt.Sprintf("%g", to), duration, repeat, s...)
}

// AnimateSkewX animates the skewX transformation
func (svg *SVG) AnimateSkewX(link string, from, to, duration float64, repeat int, s ...string) {
	svg.AnimateTransform(link, "skewX", fmt.Sprintf("%g", from), fmt.Sprintf("%g", to), duration, repeat, s...)
}

// AnimateSkewY animates the skewY transformation
func (svg *SVG) AnimateSkewY(link string, from, to, duration float64, repeat int, s ...string) {
	svg.AnimateTransform(link, "skewY", fmt.Sprintf("%g", from), fmt.Sprintf("%g", to), duration, repeat, s...)
}

// Utility

// Grid draws a grid at the specified coordinate, dimensions, and spacing, with optional style.
func (svg *SVG) Grid(x int, y int, w int, h int, n int, s ...string) {

	if len(s) > 0 {
		svg.Gstyle(s[0])
	}
	for ix := x; ix <= x+w; ix += n {
		svg.Line(ix, y, ix, y+h)
	}

	for iy := y; iy <= y+h; iy += n {
		svg.Line(x, iy, x+w, iy)
	}
	if len(s) > 0 {
		svg.Gend()
	}

}

// Support functions

// coordpair returns a coordinate pair as a string
func coordpair(x, y int) string {
	return fmt.Sprintf("%d %d", x, y)
}

// sce makes start, center, end coordinates string for animate transformations
func sce(start, center, end int) string {
	return fmt.Sprintf("%d %d %d", start, center, end)
}

// repeatString computes the repeat string for animation methods
// repeat <= 0 --> "indefinite", otherwise the integer string
func repeatString(n int) string {
	if n > 0 {
		return fmt.Sprintf("%d", n)
	}
	return "indefinite"
}

// style returns a style name,attribute string
func style(s string) string {
	if len(s) > 0 {
		return fmt.Sprintf(`style="%s"`, s)
	}
	return s
}

// pp returns a series of polygon points
func (svg *SVG) pp(x []int, y []int, tag string) {
	svg.print(tag)
	if len(x) != len(y) {
		svg.print(" ")
		return
	}
	lx := len(x) - 1
	for i := 0; i < lx; i++ {
		svg.print(coord(x[i], y[i]) + " ")
	}
	svg.print(coord(x[lx], y[lx]))
}

// endstyle modifies an SVG object, with either a series of name="value" pairs,
// or a single string containing a style
func endstyle(s []string, endtag string) string {
	if len(s) > 0 {
		nv := ""
		for i := 0; i < len(s); i++ {
			if strings.Index(s[i], "=") > 0 {
				nv += (s[i]) + " "
			} else {
				nv += style(s[i]) + " "
			}
		}
		return nv + endtag
	}
	return endtag

}

// tt creates a xml element, tag containing s
func (svg *SVG) tt(tag string, s string) {
	svg.print("<" + tag + ">")
	xml.Escape(svg.Writer, []byte(s))
	svg.println("</" + tag + ">")
}

// poly compiles the polygon element
func (svg *SVG) poly(x []int, y []int, tag string, s ...string) {
	svg.pp(x, y, "<"+tag+" points=\"")
	svg.print(`" ` + endstyle(s, "/>\n"))
}

// onezero returns "0" or "1"
func onezero(flag bool) string {
	if flag {
		return "1"
	}
	return "0"
}

// pct returns a percetage, capped at 100
func pct(n uint8) uint8 {
	if n > 100 {
		return 100
	}
	return n
}

// islink determines if a string is a script reference
func islink(link string) bool {
	return strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "#") ||
		strings.HasPrefix(link, "../") || strings.HasPrefix(link, "./")
}

// group returns a group element
func group(tag string, value string) string { return fmt.Sprintf(`<g %s="%s">`, tag, value) }

// scale return the scale string for the transform
func scale(n float64) string { return fmt.Sprintf(`scale(%g)`, n) }

// scaleXY return the scale string for the transform
func scaleXY(dx, dy float64) string { return fmt.Sprintf(`scale(%g,%g)`, dx, dy) }

// skewx returns the skewX string for the transform
func skewX(angle float64) string { return fmt.Sprintf(`skewX(%g)`, angle) }

// skewx returns the skewX string for the transform
func skewY(angle float64) string { return fmt.Sprintf(`skewY(%g)`, angle) }

// rotate returns the rotate string for the transform
func rotate(r float64) string { return fmt.Sprintf(`rotate(%g)`, r) }

// translate returns the translate string for the transform
func translate(x, y int) string { return fmt.Sprintf(`translate(%d,%d)`, x, y) }

// coord returns a coordinate string
func coord(x int, y int) string { return fmt.Sprintf(`%d,%d`, x, y) }

// ptag returns the beginning of the path element
func ptag(x int, y int) string { return fmt.Sprintf(`<path d="M%s`, coord(x, y)) }

// loc returns the x and y coordinate attributes
func loc(x int, y int) string { return fmt.Sprintf(`x="%d" y="%d"`, x, y) }

// href returns the href name and attribute
func href(s string) string { return fmt.Sprintf(`xlink:href="%s"`, s) }

// dim returns the dimension string (x, y coordinates and width, height)
func dim(x int, y int, w int, h int) string {
	return fmt.Sprintf(`x="%d" y="%d" width="%d" height="%d"`, x, y, w, h)
}

// fsattr returns the XML attribute representation of a filterspec, ignoring empty attributes
func fsattr(s Filterspec) string {
	attrs := ""
	if len(s.In) > 0 {
		attrs += fmt.Sprintf(`in="%s" `, s.In)
	}
	if len(s.In2) > 0 {
		attrs += fmt.Sprintf(`in2="%s" `, s.In2)
	}
	if len(s.Result) > 0 {
		attrs += fmt.Sprintf(`result="%s" `, s.Result)
	}
	return attrs
}

// tablevalues outputs a series of values as a XML attribute
func (svg *SVG) tablevalues(s string, t []float64) {
	svg.printf(` %s="`, s)
	for i := 0; i < len(t)-1; i++ {
		svg.printf("%g ", t[i])
	}
	svg.printf(`%g"%s`, t[len(t)-1], emptyclose)
}

// imgchannel validates the image channel indicator
func imgchannel(c string) string {
	switch c {
	case "R", "G", "B", "A":
		return c
	case "r", "g", "b", "a":
		return strings.ToUpper(c)
	case "red", "green", "blue", "alpha":
		return strings.ToUpper(c[0:1])
	case "Red", "Green", "Blue", "Alpha":
		return c[0:1]
	}
	return "R"
}
