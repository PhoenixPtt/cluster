package drawapi

import (
	"image"
	"image/color"

	"fmt"
	"image/draw"
	"image/png"
	"os"

	"strconv"

	"github.com/ajstarks/svgo"
	"github.com/vdobler/chart"
	"github.com/vdobler/chart/imgg"
	"github.com/vdobler/chart/svgg"
	"github.com/vdobler/chart/txtg"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

var Background = color.RGBA{0xff, 0xff, 0xff, 0xff}

// var pngsize int
var XYpoints plotter.XYs

type Dumper struct {
	N, M, W, H, Cnt           int
	S                         *svg.SVG
	I                         *image.RGBA
	svgFile, imgFile, txtFile *os.File
}

func NewDumper(name string, n, m, w, h int) *Dumper {
	var err error
	dumper := Dumper{N: n, M: m, W: w, H: h}

	dumper.svgFile, err = os.Create(name + ".svg")
	if err != nil {
		panic(err)
	}
	dumper.S = svg.New(dumper.svgFile)
	dumper.S.Start(n*w, m*h)
	dumper.S.Title(name)
	dumper.S.Rect(0, 0, n*w, m*h, "fill: #ffffff")

	dumper.imgFile, err = os.Create(name + ".png")
	if err != nil {
		panic(err)
	}
	dumper.I = image.NewRGBA(image.Rect(0, 0, n*w, m*h))
	bg := image.NewUniform(color.RGBA{0xff, 0xff, 0xff, 0xff})
	draw.Draw(dumper.I, dumper.I.Bounds(), bg, image.ZP, draw.Src)

	dumper.txtFile, err = os.Create(name + ".txt")
	if err != nil {
		panic(err)
	}

	return &dumper
}
func (d *Dumper) Close() {
	png.Encode(d.imgFile, d.I)
	d.imgFile.Close()

	d.S.End()
	d.svgFile.Close()

	d.txtFile.Close()
}

func (d *Dumper) Plot(c chart.Chart) {
	row, col := d.Cnt/d.N, d.Cnt%d.N

	igr := imgg.AddTo(d.I, col*d.W, row*d.H, d.W, d.H, color.RGBA{0xff, 0xff, 0xff, 0xff}, nil, nil)
	c.Plot(igr)

	sgr := svgg.AddTo(d.S, col*d.W, row*d.H, d.W, d.H, "", 12, color.RGBA{0xff, 0xff, 0xff, 0xff})
	c.Plot(sgr)

	tgr := txtg.New(100, 30)
	c.Plot(tgr)
	d.txtFile.Write([]byte(tgr.String() + "\n\n\n"))

	d.Cnt++

}

func DrawBar(title string, names []string, cpudata []float64, memdata []float64) {

	const (
		width  = 500
		height = 300
		N      = 1
		M      = 1
	)
	fmt.Println("names", names, "DrawBar", cpudata, "memdatas", memdata)

	charts := make([]chart.Chart, 0, N*M)

	// Bar charts
	ebit := chart.BarChart{Title: title}

	//containers name/id
	ebit.XRange.Category = names //[]string{"2007", "2008", "2009", "2010"}
	ebit.XRange.Label, ebit.YRange.Label = "Container", "resource percent"
	ebit.Key.Pos, ebit.Key.Cols, ebit.Key.Border = "otc", 2, -1
	ebit.YRange.ShowZero = true
	ebit.ShowVal = 1

	num := len(cpudata)
	var xlab []float64
	for i := 0; i < num; i++ {
		str := strconv.Itoa(i)
		strTofloat, err := strconv.ParseFloat(str, 64)
		if err != nil {
			// strTofloat = 0.0
			return
		}
		xlab = append(xlab, strTofloat)
	}

	fmt.Println("DrawBar numnumnum", num, xlab)

	ebit.AddDataPair("cpu%", xlab, cpudata,
		chart.Style{Symbol: '#', LineColor: color.NRGBA{0x30, 0x30, 0xff, 0xff}, LineWidth: 2, FillColor: color.NRGBA{0xcb, 0xcb, 0xff, 0xff}})
	ebit.AddDataPair("mem%", xlab, memdata,
		chart.Style{Symbol: 'O', LineColor: color.NRGBA{0xe0, 0x44, 0x44, 0xff}, LineWidth: 2, FillColor: color.NRGBA{0xf6, 0xb5, 0xcc, 0xff}})
	charts = append(charts, &ebit)
	dumper := NewDumper(title, N, M, width, height)
	fmt.Println("dumper.N", dumper.N)
	defer dumper.Close()
	for _, c := range charts {
		dumper.Plot(c)
		c.Reset()
	}
}

func DrawPlot(title string, xlabel string, ylabel string, curValue [2]float64) {

	var val plotter.XY
	val.X = curValue[0]
	val.Y = curValue[1]

	//judge XY num
	num := len(XYpoints)
	if num >= 10 {
		var xy plotter.XYs
		for i := 1; i < 10; i++ {
			xy = append(xy, XYpoints[i])
		}
		xy = append(xy, val)
		XYpoints = xy
	} else {
		XYpoints = append(XYpoints, val)
	}
	fmt.Println("aaaaaaaaaaaaa", XYpoints)

	p, _ := plot.New()
	p.Title.Text = title
	p.X.Label.Text = xlabel
	p.Y.Label.Text = ylabel
	p.Y.Max = 0.3
	plotutil.AddLinePoints(p, XYpoints)

	p.Save(5*vg.Inch, 5*vg.Inch, title+".png")
}
