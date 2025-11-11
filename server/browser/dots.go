package browser

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func dotHandler(w http.ResponseWriter, r *http.Request) {
	const (
		defaultWidth  = 64
		defaultHeight = 32
		defaultRadius = 0.3
		maxDimension  = 256
	)

	parseIntParam := func(r *http.Request, name string, defaultValue int) (int, error) {
		if v := r.URL.Query().Get(name); v != "" {
			v, err := strconv.Atoi(v)
			if err != nil {
				return 0, fmt.Errorf("parameter %q must be an integer", name)
			}

			if v <= 0 || v > maxDimension {
				return 0, fmt.Errorf("parameter %q must be between 1 and %d", name, maxDimension)
			}

			if v != 0 {
				return v, nil
			}
		}
		return defaultValue, nil
	}

	width, err := parseIntParam(r, "w", defaultWidth)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	height, err := parseIntParam(r, "h", defaultHeight)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	radius := defaultRadius
	if v := r.URL.Query().Get("r"); v != "" {
		v, err := strconv.ParseFloat(v, 64)
		if err != nil {
			http.Error(w, `parameter "r" must be a float`, http.StatusBadRequest)
			return
		}

		if v <= 0 || v > 1 {
			http.Error(w, `parameter "r" must be between 0 and 1`, http.StatusBadRequest)
			return
		}

		if v != 0 {
			radius = v
		}

	}

	var b bytes.Buffer
	b.Grow(64 + width*height*48)

	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" fill="#fff">\n`,
		width, height,
	))

	radiusStr := strconv.FormatFloat(radius, 'f', -1, 64)
	for y := range height {
		cy := strconv.FormatFloat(float64(y)+0.5, 'f', -1, 64)
		for x := range width {
			cx := strconv.FormatFloat(float64(x)+0.5, 'f', -1, 64)
			b.WriteString(`<circle cx="`)
			b.WriteString(cx)
			b.WriteString(`" cy="`)
			b.WriteString(cy)
			b.WriteString(`" r="`)
			b.WriteString(radiusStr)
			b.WriteString(`"/>`)
			b.WriteString("\n")
		}
	}

	b.WriteString("</svg>\n")

	sum := md5.Sum(b.Bytes())
	etag := strconv.Quote(hex.EncodeToString(sum[:]))

	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	w.Header().Set("ETag", etag)

	http.ServeContent(w, r, "dots.svg", time.Time{}, bytes.NewReader(b.Bytes()))
}
