package utils

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/nfnt/resize"
)

// ImageConfig holds compression configuration
type ImageConfig struct {
	MaxWidth    uint
	MaxHeight   uint
	Quality     int // 1-100 for JPEG
	CompressPng bool
}

// DefaultAvatarConfig returns default config for avatars
func DefaultAvatarConfig() ImageConfig {
	return ImageConfig{
		MaxWidth:  800,
		MaxHeight: 800,
		Quality:   85,
		CompressPng: false,
	}
}

// DefaultVisaDocConfig returns default config for visa documents
func DefaultVisaDocConfig() ImageConfig {
	return ImageConfig{
		MaxWidth:  1200,
		MaxHeight: 1200,
		Quality:   80,
		CompressPng: false,
	}
}

// CompressImage compresses an image file to reduce file size while maintaining quality
func CompressImage(inputPath, outputPath string, config ImageConfig) error {
	// Open the input file
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Decode the image
	var img image.Image
	ext := strings.ToLower(filepath.Ext(inputPath))
	
	switch ext {
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(file)
		if err != nil {
			return fmt.Errorf("failed to decode JPEG: %w", err)
		}
	case ".png":
		img, err = png.Decode(file)
		if err != nil {
			return fmt.Errorf("failed to decode PNG: %w", err)
		}
	default:
		// For non-image files (PDF, DOC, etc.), just copy the file
		return copyFile(inputPath, outputPath)
	}

	// Resize the image if needed
	bounds := img.Bounds()
	width := uint(bounds.Dx())
	height := uint(bounds.Dy())

	// Only resize if image is larger than max dimensions
	if width > config.MaxWidth || height > config.MaxHeight {
		img = resize.Thumbnail(config.MaxWidth, config.MaxHeight, img, resize.Lanczos3)
	}

	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Encode the compressed image
	if ext == ".png" {
		encoder := png.Encoder{CompressionLevel: png.BestCompression}
		if config.CompressPng {
			err = encoder.Encode(outFile, img)
		} else {
			// For PNG, use best compression
			err = png.Encode(outFile, img)
		}
		if err != nil {
			return fmt.Errorf("failed to encode PNG: %w", err)
		}
	} else {
		// JPEG encoding with quality setting
		err = jpeg.Encode(outFile, img, &jpeg.Options{Quality: config.Quality})
		if err != nil {
			return fmt.Errorf("failed to encode JPEG: %w", err)
		}
	}

	return nil
}

// CompressAndSave compresses an image and saves it to the specified path
func CompressAndSave(inputPath, outputPath string, config ImageConfig) error {
	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Compress the image
	if err := CompressImage(inputPath, outputPath, config); err != nil {
		return err
	}

	// Get file sizes to log compression ratio
	inputInfo, _ := os.Stat(inputPath)
	outputInfo, _ := os.Stat(outputPath)
	
	if inputInfo != nil && outputInfo != nil {
		inputSize := inputInfo.Size()
		outputSize := outputInfo.Size()
		if inputSize > 0 {
			compressionRatio := float64(outputSize) / float64(inputSize) * 100
			fmt.Printf("Image compressed: %.2f%% of original size\n", compressionRatio)
		}
	}

	return nil
}

// copyFile copies a file from source to destination
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// GetImageDimensions returns the dimensions of an image file
func GetImageDimensions(filePath string) (width, height int, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	img, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, err
	}

	return img.Width, img.Height, nil
}

