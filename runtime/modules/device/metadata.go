package device

type Metadata struct {
	Width  int
	Height int
	Is2x   bool
}

func (m Metadata) ScaledWidth() int {
	if m.Is2x {
		return m.Width * 2
	}
	return m.Width
}

func (m Metadata) ScaledHeight() int {
	if m.Is2x {
		return m.Height * 2
	}
	return m.Height
}
