package loader

//go:generate go tool enumer -type ImageFormat -trimprefix Image -transform lower -output image_format_string.go

type ImageFormat uint8

const (
	ImageWebP ImageFormat = iota
	ImageGIF
)

func (f ImageFormat) ContentType() string {
	switch f {
	case ImageGIF:
		return "image/gif"
	case ImageWebP:
		return "image/webp"
	default:
		return ""
	}
}
