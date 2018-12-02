package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	"golang.org/x/image/webp"
)

func logAndExit(msg string, err error) {
	if msg != "" {
		fmt.Fprintf(os.Stderr, "%s: %v\n", msg, err)
	} else {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
	os.Exit(-1)
}

// ImageType ...
type ImageType string

// ImageTypes supported
var ImageTypes = struct {
	JPEG        ImageType
	PNG         ImageType
	BMP         ImageType
	TIFF        ImageType
	GIF         ImageType
	WEBP        ImageType
	UNSUPPORTED ImageType
}{
	JPEG:        "jpeg",
	PNG:         "png",
	BMP:         "bmp",
	TIFF:        "tiff",
	GIF:         "gif",
	WEBP:        "webp",
	UNSUPPORTED: "unsupported",
}

func getImageType(fileExt string) ImageType {
	switch strings.ToLower(fileExt) {
	case "jpg":
		fallthrough
	case "jpeg":
		return ImageTypes.JPEG
	case "png":
		return ImageTypes.PNG
	case "bmp":
		return ImageTypes.BMP
	case "tiff":
		return ImageTypes.TIFF
	case "gif":
		return ImageTypes.GIF
	case "webp":
		return ImageTypes.WEBP
	default:
		return ImageTypes.UNSUPPORTED
	}
}

func createFile(filePath string) *os.File {
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		err := os.Remove(filePath)
		if err != nil {
			logAndExit("Error deleting file '%s':", err)
		}
	}

	file, err := os.Create(filePath)
	if err != nil {
		logAndExit(fmt.Sprintf("error creating file '%s':", filePath), err)
	}
	return file
}

func loadImage(fileName string, imageType ImageType) *image.Image {
	file, errOpen := os.Open(fileName)
	if errOpen != nil {
		logAndExit(fmt.Sprintf("error when opening file '%s':", fileName), errOpen)
	}
	defer file.Close()

	var imageData image.Image
	var err error
	switch imageType {
	case ImageTypes.JPEG:
		imageData, err = jpeg.Decode(file)
	case ImageTypes.PNG:
		imageData, _, err = image.Decode(file)
	case ImageTypes.BMP:
		imageData, err = bmp.Decode(file)
	case ImageTypes.TIFF:
		imageData, err = tiff.Decode(file)
	case ImageTypes.GIF:
		imageData, err = gif.Decode(file)
	case ImageTypes.WEBP:
		imageData, err = webp.Decode(file)
	case ImageTypes.UNSUPPORTED:
		logAndExit("", fmt.Errorf("error when loading image '%s': unsupported type '%s'", fileName, imageType))
	}

	if err != nil {
		logAndExit(fmt.Sprintf("error when decoding image from file '%s'", fileName), err)
	}

	return &imageData
}

func encodeImageToBase64(img *image.Image, imageType ImageType) string {
	var buff bytes.Buffer
	var err error
	var imageTypeStr string
	switch imageType {
	case ImageTypes.JPEG:
		err = jpeg.Encode(&buff, *img, nil)
		imageTypeStr = "jpeg"
	case ImageTypes.PNG:
		err = png.Encode(&buff, *img)
		imageTypeStr = "png"
	case ImageTypes.BMP:
		err = bmp.Encode(&buff, *img)
		imageTypeStr = "bmp"
	case ImageTypes.TIFF:
		err = tiff.Encode(&buff, *img, nil)
		imageTypeStr = "tiff"
	case ImageTypes.GIF:
		err = gif.Encode(&buff, *img, nil)
		imageTypeStr = "gif"
	case ImageTypes.WEBP:
		fallthrough
	case ImageTypes.UNSUPPORTED:
		logAndExit("", fmt.Errorf("error when encoding image to base64: image type %s is not supported", imageType))
	}

	if err != nil {
		logAndExit("error when encoding image to base64", err)
	}

	return "data:image/" + imageTypeStr + ";base64," + base64.StdEncoding.EncodeToString(buff.Bytes())
}

func decodeImageFromBase64(data []byte) *image.Image {
	var imageType ImageType
	switch {
	case bytes.Index(data, []byte("data:image/jpeg")) == 0:
		imageType = ImageTypes.JPEG
	case bytes.Index(data, []byte("data:image/png")) == 0:
		imageType = ImageTypes.PNG
	case bytes.Index(data, []byte("data:image/bmp")) == 0:
		imageType = ImageTypes.BMP
	case bytes.Index(data, []byte("data:image/tiff")) == 0:
		imageType = ImageTypes.TIFF
	case bytes.Index(data, []byte("data:image/gif")) == 0:
		imageType = ImageTypes.GIF
	case bytes.Index(data, []byte("data:image/webp")) == 0:
		imageType = ImageTypes.WEBP
	default:
		imageType = ImageTypes.UNSUPPORTED
	}

	search := []byte("base64,")
	if idx := bytes.Index(data, search); idx > -1 {
		src := data[idx+len(search):]
		if _, err := base64.StdEncoding.Decode(data, src); err != nil {
			logAndExit("error when decoding image from base64", err)
		}
	}

	var imageData image.Image
	var err error
	dataBuffer := bytes.NewBuffer(data)
	switch imageType {
	case ImageTypes.JPEG:
		imageData, err = jpeg.Decode(dataBuffer)
	case ImageTypes.PNG:
		imageData, _, err = image.Decode(dataBuffer)
	case ImageTypes.BMP:
		imageData, err = bmp.Decode(dataBuffer)
	case ImageTypes.TIFF:
		imageData, err = tiff.Decode(dataBuffer)
	case ImageTypes.GIF:
		imageData, err = gif.Decode(dataBuffer)
	case ImageTypes.WEBP:
		imageData, err = webp.Decode(dataBuffer)
	case ImageTypes.UNSUPPORTED:
		// atempt to decode anyway
		imageData, _, err = image.Decode(dataBuffer)
	}

	if err != nil {
		logAndExit(fmt.Sprintf("error when decoding image data of type '%s'", imageType), err)
	}

	return &imageData
}

func uint8Diff(a uint8, b uint8) uint8 {
	if a > b {
		return a - b
	}
	return b - a
}

var colorTolerance uint8 = 110
var colorToleranceUniform uint8 = 100

func sameColor(a *color.RGBA, b *color.RGBA) bool {
	aa := *a
	bb := *b
	dR := uint8Diff(aa.R, bb.R)
	dG := uint8Diff(aa.G, bb.G)
	dB := uint8Diff(aa.B, bb.B)

	t := colorTolerance
	if dR == dG && dG == dB {
		t = colorToleranceUniform
	}

	return dR <= t && dG <= t && dB <= t
}

func makeBackgroundTransparent(img *image.Image) (bool, *image.RGBA) {
	imageData := *img
	imageRGBA := image.NewRGBA(imageData.Bounds())
	draw.Draw(imageRGBA, imageData.Bounds(), imageData, image.ZP, draw.Src)
	if imageRGBA.Opaque() {
		backgroundColor := imageRGBA.RGBAAt(0, 0)
		bounds := imageRGBA.Bounds()
		width := bounds.Dx()
		height := bounds.Dy()
		for x := 0; x < width; x++ {
			for y := 0; y < height; y++ {
				color := imageRGBA.RGBAAt(x, y)
				if sameColor(&color, &backgroundColor) {
					color.A = 0
					imageRGBA.SetRGBA(x, y, color)
				}
			}
		}
		return true, imageRGBA
	}
	return false, nil
}

func main() {
	if len(os.Args) < 2 {
		logAndExit("", errors.New("image file path required - e.g. red-jpg.jpg"))
	}

	fileName := os.Args[1] // e.g. "red-jpg.jpg"
	pipeThroughBase64 := false
	if len(os.Args) > 2 {
		ptb64, err := strconv.ParseBool(strings.ToLower(os.Args[2]))
		if err != nil {
			logAndExit(fmt.Sprintf("second argument has to be true or false - got %s", os.Args[2]), err)
		}
		pipeThroughBase64 = ptb64
	}

	fileExt := filepath.Ext(fileName)
	imageType := getImageType(fileExt[1:])
	fileNameNoExt := fileName[0 : len(fileName)-len(fileExt)]

	imageData := loadImage(fileName, imageType)

	if pipeThroughBase64 {
		base64Encoded := encodeImageToBase64(imageData, imageType)
		imageData = decodeImageFromBase64([]byte(base64Encoded))
	}

	ok, imageRGBA := makeBackgroundTransparent(imageData)
	if !ok {
		logAndExit("", errors.New("image not converted - it was probably already transparent"))
	}

	outFileName := "out__" + fileNameNoExt + ".png"
	outFile := createFile(outFileName)
	defer outFile.Close()

	errEncode := png.Encode(outFile, imageRGBA)
	if errEncode != nil {
		logAndExit(fmt.Sprintf("error when encoding image file '%s':", outFileName), errEncode)
	}
}
