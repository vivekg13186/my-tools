package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/phpdave11/gofpdf"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: img2pdf image1 image2 ... output.pdf")
		os.Exit(1)
	}

	output := os.Args[len(os.Args)-1]
	images := os.Args[1 : len(os.Args)-1]

	pdf := gofpdf.New("P", "mm", "A4", "")

	for _, image := range images {
		if _, err := os.Stat(image); err != nil {
			fmt.Printf("Skipping %s: %v\n", image, err)
			continue
		}

		// Read image dimensions
		info := pdf.RegisterImageOptions(
			image,
			gofpdf.ImageOptions{ReadDpi: true},
		)

		if info == nil {
			fmt.Println("Skipping unsupported image:", image)
			continue
		}

		// Create page matching image orientation
		width := info.Width()
		height := info.Height()

		if width > height {
			pdf.AddPageFormat("L", gofpdf.SizeType{Wd: 297, Ht: 210})
		} else {
			pdf.AddPageFormat("P", gofpdf.SizeType{Wd: 210, Ht: 297})
		}

		// Fit image inside page, preserving aspect ratio
		pageW, pageH := pdf.GetPageSize()

		margin := 5.0
		maxW := pageW - 2*margin
		maxH := pageH - 2*margin

		scale := maxW / width
		if height*scale > maxH {
			scale = maxH / height
		}

		imgW := width * scale
		imgH := height * scale

		x := (pageW - imgW) / 2
		y := (pageH - imgH) / 2

		pdf.ImageOptions(
			image,
			x, y,
			imgW, imgH,
			false,
			gofpdf.ImageOptions{ImageType: filepath.Ext(image)[1:]},
			0,
			"",
		)

		fmt.Println("Added:", image)
	}

	if pdf.Err() {
		fmt.Println("PDF error:", pdf.Error())
		os.Exit(1)
	}

	if err := pdf.OutputFileAndClose(output); err != nil {
		fmt.Println("Error creating PDF:", err)
		os.Exit(1)
	}

	fmt.Println("PDF created:", output)
}
