package image

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"os"

	xdraw "golang.org/x/image/draw"
)

// GIFCompressionOptions defines settings for GIF compression
type GIFCompressionOptions struct {
	ColorCount     int  // Number of colors in palette (2-256, default 256)
	ResizePercent  int  // Resize percentage (10-100, default 100 = no resize)
	LossyLevel     int  // Lossy compression level (0-100, 0 = lossless)
	OptimizeFrames bool // Remove redundant pixels in animation frames
	FrameSkip      int  // Skip every N frames for animations (0 = keep all)
}

// CompressGIF compresses a GIF file with the given options
func CompressGIF(inputPath, outputPath string, opts GIFCompressionOptions) error {
	// Open input file
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inputFile.Close()

	// Decode GIF
	g, err := gif.DecodeAll(inputFile)
	if err != nil {
		return fmt.Errorf("failed to decode GIF: %w", err)
	}

	// Apply frame skipping first (reduces work for other operations)
	if opts.FrameSkip > 0 && len(g.Image) > 1 {
		g = skipFrames(g, opts.FrameSkip)
	}

	// Apply resizing
	if opts.ResizePercent > 0 && opts.ResizePercent < 100 {
		g = resizeGIF(g, opts.ResizePercent)
	}

	// Apply color reduction - always process to re-optimize palette
	if opts.ColorCount > 0 && opts.ColorCount <= 256 {
		g = reduceColors(g, opts.ColorCount, opts.LossyLevel)
	}

	// Optimize frames (remove redundant pixels)
	if opts.OptimizeFrames && len(g.Image) > 1 {
		g = optimizeFrames(g)
	}

	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Encode GIF
	if err := gif.EncodeAll(outputFile, g); err != nil {
		return fmt.Errorf("failed to encode GIF: %w", err)
	}

	return nil
}

// skipFrames removes frames from the animation and adjusts timing
func skipFrames(g *gif.GIF, skip int) *gif.GIF {
	if skip <= 0 || len(g.Image) <= 1 {
		return g
	}

	newImages := make([]*image.Paletted, 0, len(g.Image)/(skip+1)+1)
	newDelays := make([]int, 0, len(g.Delay)/(skip+1)+1)
	newDisposal := make([]byte, 0, len(g.Disposal)/(skip+1)+1)

	accumulatedDelay := 0

	for i := 0; i < len(g.Image); i++ {
		accumulatedDelay += g.Delay[i]

		// Keep every (skip+1)th frame, or the last frame
		if i%(skip+1) == 0 || i == len(g.Image)-1 {
			newImages = append(newImages, g.Image[i])
			newDelays = append(newDelays, accumulatedDelay)
			if len(g.Disposal) > i {
				newDisposal = append(newDisposal, g.Disposal[i])
			}
			accumulatedDelay = 0
		}
	}

	return &gif.GIF{
		Image:           newImages,
		Delay:           newDelays,
		Disposal:        newDisposal,
		LoopCount:       g.LoopCount,
		BackgroundIndex: g.BackgroundIndex,
		Config:          g.Config,
	}
}

// resizeGIF scales all frames to a percentage of original size
func resizeGIF(g *gif.GIF, percent int) *gif.GIF {
	if percent >= 100 || percent <= 0 {
		return g
	}

	scale := float64(percent) / 100.0
	newWidth := int(float64(g.Config.Width) * scale)
	newHeight := int(float64(g.Config.Height) * scale)

	if newWidth < 1 {
		newWidth = 1
	}
	if newHeight < 1 {
		newHeight = 1
	}

	newImages := make([]*image.Paletted, len(g.Image))

	for i, frame := range g.Image {
		// Calculate new frame bounds
		bounds := frame.Bounds()
		newFrameWidth := int(float64(bounds.Dx()) * scale)
		newFrameHeight := int(float64(bounds.Dy()) * scale)
		newX := int(float64(bounds.Min.X) * scale)
		newY := int(float64(bounds.Min.Y) * scale)

		if newFrameWidth < 1 {
			newFrameWidth = 1
		}
		if newFrameHeight < 1 {
			newFrameHeight = 1
		}

		// Create temporary RGBA for high-quality scaling
		tmpRGBA := image.NewRGBA(image.Rect(0, 0, newFrameWidth, newFrameHeight))
		xdraw.CatmullRom.Scale(tmpRGBA, tmpRGBA.Bounds(), frame, bounds, xdraw.Over, nil)

		// Create new paletted image with same palette
		newFrame := image.NewPaletted(
			image.Rect(newX, newY, newX+newFrameWidth, newY+newFrameHeight),
			frame.Palette,
		)

		// Convert back to paletted
		draw.FloydSteinberg.Draw(newFrame, newFrame.Bounds(), tmpRGBA, image.Point{})

		newImages[i] = newFrame
	}

	return &gif.GIF{
		Image:           newImages,
		Delay:           g.Delay,
		Disposal:        g.Disposal,
		LoopCount:       g.LoopCount,
		BackgroundIndex: g.BackgroundIndex,
		Config: image.Config{
			ColorModel: g.Config.ColorModel,
			Width:      newWidth,
			Height:     newHeight,
		},
	}
}

// reduceColors reduces the color palette of each frame
func reduceColors(g *gif.GIF, colorCount int, lossyLevel int) *gif.GIF {
	if colorCount < 2 {
		colorCount = 2
	}
	if colorCount > 256 {
		colorCount = 256
	}

	newImages := make([]*image.Paletted, len(g.Image))

	for i, frame := range g.Image {
		// Create a new reduced palette based on actual colors used
		newPalette := quantizePalette(frame, colorCount)

		// Create new paletted image with reduced palette
		newFrame := image.NewPaletted(frame.Bounds(), newPalette)

		// Use dithering for better quality with fewer colors
		if lossyLevel > 50 {
			// Simple copy for high lossy level (faster, more artifacts)
			draw.Draw(newFrame, newFrame.Bounds(), frame, frame.Bounds().Min, draw.Src)
		} else {
			// Floyd-Steinberg dithering for better quality
			draw.FloydSteinberg.Draw(newFrame, newFrame.Bounds(), frame, frame.Bounds().Min)
		}

		newImages[i] = newFrame
	}

	return &gif.GIF{
		Image:           newImages,
		Delay:           g.Delay,
		Disposal:        g.Disposal,
		LoopCount:       g.LoopCount,
		BackgroundIndex: g.BackgroundIndex,
		Config:          g.Config,
	}
}

// quantizePalette creates a reduced color palette from an image using median cut algorithm
func quantizePalette(img *image.Paletted, maxColors int) color.Palette {
	bounds := img.Bounds()

	// Count actual colors used in the image
	colorUsage := make(map[uint32]int)
	colorMap := make(map[uint32]color.Color)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.At(x, y)
			r, g, b, a := c.RGBA()
			key := (r>>8)<<24 | (g>>8)<<16 | (b>>8)<<8 | (a >> 8)
			colorUsage[key]++
			colorMap[key] = c
		}
	}

	// If we already have fewer colors than requested, still rebuild palette
	// to ensure it's optimized
	type colorEntry struct {
		key   uint32
		c     color.Color
		count int
	}

	entries := make([]colorEntry, 0, len(colorUsage))
	for key, count := range colorUsage {
		entries = append(entries, colorEntry{key, colorMap[key], count})
	}

	// Sort by frequency (most used first)
	for i := 0; i < len(entries)-1; i++ {
		maxIdx := i
		for j := i + 1; j < len(entries); j++ {
			if entries[j].count > entries[maxIdx].count {
				maxIdx = j
			}
		}
		entries[i], entries[maxIdx] = entries[maxIdx], entries[i]
	}

	// Take the most frequently used colors up to maxColors
	paletteSize := maxColors
	if len(entries) < paletteSize {
		paletteSize = len(entries)
	}

	newPalette := make(color.Palette, paletteSize)
	for i := 0; i < paletteSize; i++ {
		newPalette[i] = entries[i].c
	}

	// Ensure we have at least one transparent color if original had one
	hasTransparent := false
	for _, c := range newPalette {
		_, _, _, a := c.RGBA()
		if a == 0 {
			hasTransparent = true
			break
		}
	}

	if !hasTransparent && paletteSize < maxColors {
		// Check if original had transparency
		for _, c := range img.Palette {
			_, _, _, a := c.RGBA()
			if a == 0 {
				newPalette = append(newPalette, color.Transparent)
				break
			}
		}
	}

	return newPalette
}

// optimizeFrames removes redundant pixels between consecutive frames
func optimizeFrames(g *gif.GIF) *gif.GIF {
	if len(g.Image) <= 1 {
		return g
	}

	// Create a canvas to track the current display state
	canvas := image.NewRGBA(image.Rect(0, 0, g.Config.Width, g.Config.Height))

	newImages := make([]*image.Paletted, len(g.Image))
	newDisposal := make([]byte, len(g.Image))

	// First frame is always kept as-is
	newImages[0] = g.Image[0]
	if len(g.Disposal) > 0 {
		newDisposal[0] = g.Disposal[0]
	}

	// Draw first frame to canvas
	draw.Draw(canvas, g.Image[0].Bounds(), g.Image[0], g.Image[0].Bounds().Min, draw.Over)

	for i := 1; i < len(g.Image); i++ {
		frame := g.Image[i]
		bounds := frame.Bounds()

		// Find the bounding box of changed pixels
		minX, minY := bounds.Max.X, bounds.Max.Y
		maxX, maxY := bounds.Min.X, bounds.Min.Y
		hasChanges := false

		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				frameColor := frame.At(x, y)
				canvasColor := canvas.At(x, y)

				r1, g1, b1, a1 := frameColor.RGBA()
				r2, g2, b2, a2 := canvasColor.RGBA()

				if r1 != r2 || g1 != g2 || b1 != b2 || a1 != a2 {
					hasChanges = true
					if x < minX {
						minX = x
					}
					if x > maxX {
						maxX = x
					}
					if y < minY {
						minY = y
					}
					if y > maxY {
						maxY = y
					}
				}
			}
		}

		if !hasChanges {
			// No changes, create a minimal 1x1 transparent frame
			newFrame := image.NewPaletted(image.Rect(0, 0, 1, 1), frame.Palette)
			newImages[i] = newFrame
		} else {
			// Create optimized frame with only changed region
			optimizedBounds := image.Rect(minX, minY, maxX+1, maxY+1)
			newFrame := image.NewPaletted(optimizedBounds, frame.Palette)

			for y := optimizedBounds.Min.Y; y < optimizedBounds.Max.Y; y++ {
				for x := optimizedBounds.Min.X; x < optimizedBounds.Max.X; x++ {
					newFrame.Set(x, y, frame.At(x, y))
				}
			}

			newImages[i] = newFrame
		}

		// Update disposal
		if len(g.Disposal) > i {
			newDisposal[i] = g.Disposal[i]
		} else {
			newDisposal[i] = gif.DisposalNone
		}

		// Update canvas for next frame comparison
		draw.Draw(canvas, frame.Bounds(), frame, frame.Bounds().Min, draw.Over)
	}

	return &gif.GIF{
		Image:           newImages,
		Delay:           g.Delay,
		Disposal:        newDisposal,
		LoopCount:       g.LoopCount,
		BackgroundIndex: g.BackgroundIndex,
		Config:          g.Config,
	}
}
