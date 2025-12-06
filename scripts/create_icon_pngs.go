package main

import (
	"fmt"
	"image"
	"image/png"
	"os"

	"github.com/nfnt/resize"
	"golang.org/x/image/draw"
)

func main() {
	// Read the source icon
	srcFile, err := os.Open("internal/assets/real_icon.ico")
	if err != nil {
		fmt.Printf("Error opening source icon: %v\n", err)
		os.Exit(1)
	}
	defer srcFile.Close()

	// Decode the icon
	srcImg, _, err := image.Decode(srcFile)
	if err != nil {
		fmt.Printf("Error decoding icon: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Source image: %dx%d\n", srcImg.Bounds().Dx(), srcImg.Bounds().Dy())

	// Create images at different sizes
	sizes := []int{16, 32, 48, 256}
	
	for _, size := range sizes) {
		// Resize image
		resized := resize.Resize(uint(size), uint(size), srcImg, resize.Lanczos3)
		
		// Save as PNG for verification
		outFile, err := os.Create(fmt.Sprintf("internal/assets/icon_%dx%d.png", size, size))
		if err != nil {
			fmt.Printf("Error creating output file: %v\n", err)
			continue
		}
		
		err = png.Encode(outFile, resized)
		outFile.Close()
		
		if err != nil {
			fmt.Printf("Error encoding PNG: %v\n", err)
		} else {
			fmt.Printf("Created %dx%d version\n", size, size)
		}
	}
	
	fmt.Println("\n✓ PNG files created. Please use an online tool to convert them to a multi-resolution ICO file:")
	fmt.Println("  1. Visit: https://convertio.co/png-ico/ or https://www.icoconverter.com/")
	fmt.Println("  2. Upload all PNG files (16x16, 32x32, 48x48, 256x256)")
	fmt.Println("  3. Download the combined ICO file")
	fmt.Println("  4. Replace internal/assets/real_icon.ico with the new file")
}
