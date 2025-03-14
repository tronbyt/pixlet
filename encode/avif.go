package encode

import (
	"fmt"
	"image"
	"image/draw"
	"runtime"
	"sync"
	"time"
	"unsafe"

	"github.com/ebitengine/purego"
)

var initOnce sync.Once

const (
	AVIF_HEADER_MINI      = 1
	AVIF_RESULT_OK        = 0
	AVIF_QUALITY_LOSSLESS = int32(100)

	AVIF_SPEED_DEFAULT = int32(-1)
	AVIF_SPEED_SLOWEST = int32(0)
	AVIF_SPEED_FASTEST = int32(10)
)

// Renders screens to AVIF. Optionally pass filters for postprocessing
// each individual frame.
func (s *Screens) EncodeAVIF(maxDuration int, filters ...ImageFilter) ([]byte, error) {
	initOnce.Do(initAvif)

	images, err := s.render(filters...)
	if err != nil {
		return nil, err
	}

	if len(images) == 0 {
		return []byte{}, nil
	}

	chroma := avifPixelFormatYuv444
	remainingDuration := maxDuration

	encoder := avifEncoderCreate()
	defer avifEncoderDestroy(encoder)

	encoder.MaxThreads = int32(runtime.NumCPU())
	encoder.Quality = AVIF_QUALITY_LOSSLESS
	encoder.QualityAlpha = AVIF_QUALITY_LOSSLESS
	encoder.Speed = AVIF_SPEED_SLOWEST
	encoder.HeaderFormat = AVIF_HEADER_MINI
	encoder.Timescale = uint64(time.Second / time.Millisecond)

	for _, im := range images {
		frameDuration := int(s.delay)

		if maxDuration > 0 {
			if frameDuration > remainingDuration {
				frameDuration = remainingDuration
			}
			remainingDuration -= frameDuration
		}

		i := imageToRGBA(im)

		img := avifImageCreate(i.Bounds().Dx(), i.Bounds().Dy(), 8, chroma)
		defer avifImageDestroy(img)

		var rgb avifRGBImage
		avifRGBImageSetDefaults(&rgb, img)

		rgb.MaxThreads = int32(runtime.NumCPU())
		rgb.AlphaPremultiplied = 1
		rgb.Pixels = (*uint8)(&i.Pix[0])
		rgb.RowBytes = uint32(i.Stride)

		if ret := avifImageRGBToYUV(img, &rgb); ret != AVIF_RESULT_OK {
			return nil, fmt.Errorf("error converting RGB to YUV: %s", avifResultToString(ret))
		}

		flags := 0
		if len(images) == 1 {
			flags = avifAddImageFlagSingle
		}
		if ret := avifEncoderAddImage(encoder, img, uint64(frameDuration), flags); ret != AVIF_RESULT_OK {
			return nil, fmt.Errorf("error adding frame: %s", avifResultToString(ret))
		}

		if maxDuration > 0 && remainingDuration <= 0 {
			break
		}
	}

	var output avifRWData
	defer avifRWDataFree(&output)

	if ret := avifEncoderFinish(encoder, &output); ret != AVIF_RESULT_OK {
		return nil, fmt.Errorf("error encoding animation: %s", avifResultToString(ret))
	}

	return unsafe.Slice(output.Data, output.Size), nil
}

func imageToRGBA(src image.Image) *image.RGBA {
	if dst, ok := src.(*image.RGBA); ok {
		return dst
	}

	b := src.Bounds()
	dst := image.NewRGBA(b)
	draw.Draw(dst, dst.Bounds(), src, b.Min, draw.Src)

	return dst
}

var (
	libavif uintptr

	avifVersion                func() string
	avifRGBImageSetDefaults    func(*avifRGBImage, *avifImage)
	avifRGBImageAllocatePixels func(*avifRGBImage) int
	avifRGBImageFreePixels     func(*avifRGBImage)
	avifImageRGBToYUV          func(*avifImage, *avifRGBImage) int
	avifImageCreate            func(int, int, int, int) *avifImage
	avifImageDestroy           func(*avifImage)
	avifEncoderCreate          func() *avifEncoder
	avifEncoderDestroy         func(*avifEncoder)
	avifEncoderAddImage        func(*avifEncoder, *avifImage, uint64, int) int
	avifEncoderFinish          func(*avifEncoder, *avifRWData) int
	avifRWDataFree             func(*avifRWData)
	avifResultToString         func(int) string
)

const (
	avifPixelFormatYuv444 = 1
	avifPixelFormatYuv422 = 2
	avifPixelFormatYuv420 = 3

	avifAddImageFlagSingle = 2
)

type avifEncoderData struct{}

type avifCodecSpecificOptions struct{}

type avifScalingMode struct {
	Horizontal avifFraction
	Vertical   avifFraction
}

type avifFraction struct {
	N int32
	D int32
}

type avifEncoder struct {
	CodecChoice       uint32
	MaxThreads        int32
	Speed             int32
	KeyframeInterval  int32
	Timescale         uint64
	RepetitionCount   int32
	ExtraLayerCount   uint32
	Quality           int32
	QualityAlpha      int32
	MinQuantizer      int32
	MaxQuantizer      int32
	MinQuantizerAlpha int32
	MaxQuantizerAlpha int32
	TileRowsLog2      int32
	TileColsLog2      int32
	AutoTiling        int32
	ScalingMode       avifScalingMode
	IoStats           avifIOStats
	Diag              avifDiagnostics
	Data              *avifEncoderData
	CsOptions         *avifCodecSpecificOptions
	HeaderFormat      uint32
	_                 [4]byte
}

type avifIOStats struct {
	ColorOBUSize uint64
	AlphaOBUSize uint64
}

type avifDiagnostics struct {
	Error [256]uint8
}

type avifImage struct {
	Width                   uint32
	Height                  uint32
	Depth                   uint32
	YuvFormat               uint32
	YuvRange                uint32
	YuvChromaSamplePosition uint32
	YuvPlanes               [3]*uint8
	YuvRowBytes             [3]uint32
	ImageOwnsYUVPlanes      int32
	AlphaPlane              *uint8
	AlphaRowBytes           uint32
	ImageOwnsAlphaPlane     int32
	AlphaPremultiplied      int32
	Icc                     avifRWData
	ColorPrimaries          uint16
	TransferCharacteristics uint16
	MatrixCoefficients      uint16
	Clli                    avifContentLightLevelInformationBox
	TransformFlags          uint32
	Pasp                    avifPixelAspectRatioBox
	Clap                    avifCleanApertureBox
	Irot                    avifImageRotation
	Imir                    avifImageMirror
	Exif                    avifRWData
	Xmp                     avifRWData
}

type avifRGBImage struct {
	Width              uint32
	Height             uint32
	Depth              uint32
	Format             uint32
	ChromaUpsampling   uint32
	ChromaDownsampling uint32
	AvoidLibYUV        int32
	IgnoreAlpha        int32
	AlphaPremultiplied int32
	IsFloat            int32
	MaxThreads         int32
	Pixels             *uint8
	RowBytes           uint32
	_                  [4]byte
}
type avifRWData struct {
	Data *uint8
	Size uint64
}

type avifContentLightLevelInformationBox struct {
	MaxCLL  uint16
	MaxPALL uint16
}

type avifPixelAspectRatioBox struct {
	HSpacing uint32
	VSpacing uint32
}

type avifCleanApertureBox struct {
	WidthN    uint32
	WidthD    uint32
	HeightN   uint32
	HeightD   uint32
	HorizOffN uint32
	HorizOffD uint32
	VertOffN  uint32
	VertOffD  uint32
}

type avifImageRotation struct {
	Angle uint8
}

type avifImageMirror struct {
	Axis uint8
}

func loadAvif() (uintptr, error) {
	libname := "libavif.so"
	if runtime.GOOS == "windows" {
		libname = "avif.dll"
	} else if runtime.GOOS == "darwin" {
		libname = "libavif.dylib"
	}

	lib, err := loadLibrary(libname)
	if err != nil {
		return 0, err
	}

	return lib, nil
}

func initAvif() {
	var err error
	libavif, err = loadAvif()
	if err != nil {
		panic(fmt.Sprintf("error loading libavif: %s\n", err))
	}

	purego.RegisterLibFunc(&avifVersion, libavif, "avifVersion")
	purego.RegisterLibFunc(&avifEncoderCreate, libavif, "avifEncoderCreate")
	purego.RegisterLibFunc(&avifEncoderDestroy, libavif, "avifEncoderDestroy")
	purego.RegisterLibFunc(&avifEncoderAddImage, libavif, "avifEncoderAddImage")
	purego.RegisterLibFunc(&avifEncoderFinish, libavif, "avifEncoderFinish")
	purego.RegisterLibFunc(&avifRWDataFree, libavif, "avifRWDataFree")
	purego.RegisterLibFunc(&avifRGBImageSetDefaults, libavif, "avifRGBImageSetDefaults")
	purego.RegisterLibFunc(&avifRGBImageAllocatePixels, libavif, "avifRGBImageAllocatePixels")
	purego.RegisterLibFunc(&avifRGBImageFreePixels, libavif, "avifRGBImageFreePixels")
	purego.RegisterLibFunc(&avifImageCreate, libavif, "avifImageCreate")
	purego.RegisterLibFunc(&avifImageDestroy, libavif, "avifImageDestroy")
	purego.RegisterLibFunc(&avifImageRGBToYUV, libavif, "avifImageRGBToYUV")
	purego.RegisterLibFunc(&avifResultToString, libavif, "avifResultToString")

	version := avifVersion()
	fmt.Printf("using libavif %s\n", version)

	var major, minor, patch int
	_, _ = fmt.Sscanf(version, "%d.%d.%d", &major, &minor, &patch)
	if major < 1 || (major == 1 && minor < 1) {
		panic("minimum required libavif version is 1.1.0")
	}
}
