package leaflet

import (
	"embed"
	"fmt"
	"image"
	"io"
	"strings"

	"image/color"
	"image/draw"
	_ "image/png" // enable PNG decoder

	"github.com/fawkesley/pollution-printouts/addresspollution"
	"github.com/fogleman/gg"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

var (
	jpegQuality   = 90
	dpi           = float64(300)
	fontSize      = float64(12)
	textPositionX = 1190 // x coordinate where to start rendering text
	textPositionY = 980  // y coordinate
	textColour    = color.RGBA{0, 0, 0, 255}
	hinting       = "full" // "none" / "full"
)

// RenderPNG adds the secret password to the right gift voucher background and writes the JPG
// data to w.
func RenderPNG(p addresspollution.PollutionLevels, w io.Writer) error {

	bgImage, err := loadBackground()
	if err != nil {
		return err
	}

	imgWidth := bgImage.Bounds().Dx()
	imgHeight := bgImage.Bounds().Dy()

	dc := gg.NewContext(imgWidth, imgHeight)
	dc.DrawImage(bgImage, 0, 0)

	renderPollutionDescription(dc, p, imgWidth)
	renderAddress(dc, p, imgWidth)
	renderNumPollutants(dc, p, imgWidth)
	renderPollutantDialValues(dc, p, imgWidth)
	renderPollutantDialComments(dc, p, imgWidth)

	// x := float64(imgWidth / 2)
	// y := float64((imgHeight / 2) - 80)

	// return dc.Image(), nil
	if err := dc.EncodePNG(w); err != nil {
		return err
	}

	// var err error

	// img, ctxt := initImageAndContext(background)

	// ctxt.SetFontSize(float64(48))

	// pt := freetype.Pt(textPositionX, textPositionY) // draw the text
	// _, err = ctxt.DrawString("TODO", pt)
	// if err != nil {
	// 	return fmt.Errorf("failed to draw string: %v", err)
	// }

	// b := bufio.NewWriter(w)
	// err = jpeg.Encode(b, img, &jpeg.Options{Quality: jpegQuality})
	// if err != nil {
	// 	return fmt.Errorf("failed to encode JPG format: %v", err)
	// }
	// err = b.Flush()
	// if err != nil {
	// 	return fmt.Errorf("failed to flush: %v", err)
	// }
	return nil
}

func renderPollutionDescription(dc *gg.Context, p addresspollution.PollutionLevels, w int) error {
	if err := dc.LoadFontFace("leaflet/assets/fonts/ARIALBD.TTF", 4*48); err != nil {
		return err
	}

	maxWidth := float64(2000)

	dc.SetColor(color.Black)
	dc.DrawStringWrapped(
		strings.ToUpper(p.PollutionDescription),
		float64(w)/2,
		500,
		0.5,
		0.5,
		maxWidth,
		1.5,
		gg.AlignCenter,
	)

	return nil
}

func renderAddress(dc *gg.Context, p addresspollution.PollutionLevels, w int) error {
	if err := dc.LoadFontFace("leaflet/assets/fonts/ARIALBD.TTF", 4*24); err != nil {
		return err
	}

	maxWidth := float64(w - 100)

	dc.SetColor(color.Black)
	dc.DrawStringWrapped(
		p.FormattedAddress,
		float64(w)/2,
		1400,
		0.5,
		0.5,
		maxWidth,
		1.5,
		gg.AlignCenter,
	)

	return nil
}

func renderNumPollutants(dc *gg.Context, p addresspollution.PollutionLevels, w int) error {
	if err := dc.LoadFontFace("leaflet/assets/fonts/ARIALBD.TTF", 23*4); err != nil {
		return err
	}

	maxWidth := float64(w - 100)

	dc.SetColor(color.Black)
	dc.DrawStringWrapped(
		fmt.Sprintf("%d", p.NumPollutantsExceedingLimits()),
		1125,
		1858,
		0.5,
		0.5,
		maxWidth,
		1.5,
		gg.AlignCenter,
	)

	return nil
}

func renderPollutantDialValues(dc *gg.Context, p addresspollution.PollutionLevels, w int) error {
	if err := dc.LoadFontFace("leaflet/assets/fonts/ARIALBD.TTF", 26*4); err != nil {
		return err
	}

	maxWidth := float64(500)

	dc.SetColor(color.Black)

	dc.DrawStringWrapped(
		fmt.Sprintf("%.1f", p.Pm2_5),
		380,
		2350,
		0.5,
		0.5,
		maxWidth,
		1.5,
		gg.AlignCenter,
	)

	dc.DrawStringWrapped(
		fmt.Sprintf("%.1f", p.Pm10),
		float64(w)/2,
		2350,
		0.5,
		0.5,
		maxWidth,
		1.5,
		gg.AlignCenter,
	)

	dc.DrawStringWrapped(
		fmt.Sprintf("%.1f", p.No2),
		float64(w-380),
		2350,
		0.5,
		0.5,
		maxWidth,
		1.5,
		gg.AlignCenter,
	)

	return nil
}

func renderPollutantDialComments(dc *gg.Context, p addresspollution.PollutionLevels, w int) error {
	if err := dc.LoadFontFace("leaflet/assets/fonts/ARIALBD.TTF", 26*4); err != nil {
		return err
	}

	// red := color.RGBA{0, 255, 255, 255}
	red := color.RGBA{255, 0, 0, 255}

	dc.SetColor(red)
	// dc.SetColor(color.Black)

	maxWidth := float64(1000)

	y := float64(3030)

	dc.DrawStringWrapped(
		fmt.Sprintf("%s safe level", p.Pm2_5SafeLevelDescription()),
		380,
		y,
		0.5,
		0.5,
		maxWidth,
		1.5,
		gg.AlignCenter,
	)

	dc.DrawStringWrapped(
		fmt.Sprintf("%s safe level", p.Pm10SafeLevelDescription()),
		float64(w)/2,
		y,
		0.5,
		0.5,
		maxWidth,
		1.5,
		gg.AlignCenter,
	)

	dc.DrawStringWrapped(
		fmt.Sprintf("%s safe level", p.No2SafeLevelDescription()),
		float64(w-380),
		y,
		0.5,
		0.5,
		maxWidth,
		1.5,
		gg.AlignCenter,
	)

	return nil
}

//go:embed assets/images/*.png
var backgrounds embed.FS

//go:embed assets/fonts/Merriweather-Bold.ttf
var fontBytes []byte

var passwordFont *truetype.Font

func init() {
	if err := loadFont(); err != nil {
		panic(err)
	}
}

func loadFont() error {
	if len(fontBytes) == 0 {
		panic("fontBytes is 0 bytes, not embedded successfully")
	}

	var err error
	passwordFont, err = freetype.ParseFont(fontBytes)
	if err != nil {
		return fmt.Errorf("failed to parse font bytes: %v", err)
	}

	fmt.Printf("parsed freetype font\n")

	return nil
}

func loadBackground() (image.Image, error) {
	reader, err := backgrounds.Open("assets/images/page-1-background.png")
	if err != nil {
		return nil, fmt.Errorf("failed to open background: %v", err)
	}

	defer reader.Close()
	background, _, err := image.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	return background, nil
}

// initImageAndContext returns a ? and a freetype context.
func initImageAndContext(background image.Image) (*image.RGBA, *freetype.Context) {
	img := image.NewRGBA(background.Bounds())
	draw.Draw(img, img.Bounds(), background, image.ZP, draw.Src)
	c := freetype.NewContext()
	c.SetDPI(dpi)
	c.SetFont(passwordFont)
	c.SetFontSize(fontSize)
	c.SetClip(img.Bounds())
	c.SetDst(img)

	c.SetSrc(image.NewUniform(textColour))
	switch hinting {
	default:
		c.SetHinting(font.HintingNone)
	case "full":
		c.SetHinting(font.HintingFull)
	}

	return img, c
}
