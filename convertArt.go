package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"math"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

type Block struct {
	Name  string
	Color color.RGBA
}

type PixelBlock struct {
	R    uint32
	G    uint32
	B    uint32
	Name string
}

var MinecraftPalette = []Block{
	{Name: "Grass Block", Color: color.RGBA{R: 124, G: 185, B: 72, A: 255}},  // Green
	{Name: "Lime Wool", Color: color.RGBA{R: 112, G: 185, B: 25, A: 255}},    // Green
	{Name: "Emerald Block", Color: color.RGBA{R: 41, G: 181, B: 58, A: 255}}, // Green
	{Name: "Green Wool", Color: color.RGBA{R: 85, G: 110, B: 27, A: 255}},    // Green

	{Name: "Dirt", Color: color.RGBA{R: 134, G: 96, B: 67, A: 255}},           // Brown
	{Name: "Spruce Planks", Color: color.RGBA{R: 115, G: 85, B: 49, A: 255}},  // Brown
	{Name: "Dark Oak Planks", Color: color.RGBA{R: 67, G: 43, B: 20, A: 255}}, // Brown
	{Name: "Brown Wool", Color: color.RGBA{R: 114, G: 71, B: 40, A: 255}},     // Brown

	{Name: "Cobblestone", Color: color.RGBA{R: 78, G: 78, B: 78, A: 255}}, // Gray
	{Name: "Stone", Color: color.RGBA{R: 125, G: 125, B: 125, A: 255}},    // Gray
	{Name: "Gravel", Color: color.RGBA{R: 132, G: 127, B: 127, A: 255}},   // Gray
	{Name: "Gray Wool", Color: color.RGBA{R: 63, G: 69, B: 72, A: 255}},   // Gray

	{Name: "Oak Planks", Color: color.RGBA{R: 162, G: 131, B: 78, A: 255}},        // Light Brown
	{Name: "Birch Planks", Color: color.RGBA{R: 204, G: 194, B: 150, A: 255}},     // Light Brown
	{Name: "Sandstone", Color: color.RGBA{R: 219, G: 207, B: 163, A: 255}},        // Light Brown
	{Name: "White Terracotta", Color: color.RGBA{R: 209, G: 178, B: 161, A: 255}}, // Light Brown

	{Name: "Sand", Color: color.RGBA{R: 220, G: 217, B: 149, A: 255}},      // Yellow
	{Name: "Yellow Wool", Color: color.RGBA{R: 181, G: 140, B: 9, A: 255}}, // Yellow
	{Name: "Gold Block", Color: color.RGBA{R: 247, G: 235, B: 76, A: 255}}, // Yellow
	{Name: "Hay Bale", Color: color.RGBA{R: 170, G: 133, B: 58, A: 255}},   // Yellow

	{Name: "Blue Wool", Color: color.RGBA{R: 53, G: 57, B: 157, A: 255}},           // Blue
	{Name: "Lapis Lazuli Block", Color: color.RGBA{R: 21, G: 119, B: 136, A: 255}}, // Blue
	{Name: "Cyan Wool", Color: color.RGBA{R: 21, G: 137, B: 145, A: 255}},          // Blue
	{Name: "Light Blue Wool", Color: color.RGBA{R: 74, G: 128, B: 255, A: 255}},    // Blue

	{Name: "Red Wool", Color: color.RGBA{R: 160, G: 39, B: 34, A: 255}},       // Red
	{Name: "Redstone Block", Color: color.RGBA{R: 164, G: 23, B: 10, A: 255}}, // Red
	{Name: "Brick Block", Color: color.RGBA{R: 145, G: 61, B: 56, A: 255}},    // Red
	{Name: "Nether Brick", Color: color.RGBA{R: 45, G: 22, B: 27, A: 255}},    // Red

	{Name: "White Wool", Color: color.RGBA{R: 222, G: 222, B: 222, A: 255}},   // White
	{Name: "Snow Block", Color: color.RGBA{R: 249, G: 254, B: 254, A: 255}},   // White
	{Name: "Quartz Block", Color: color.RGBA{R: 236, G: 233, B: 226, A: 255}}, // White
	{Name: "Iron Block", Color: color.RGBA{R: 219, G: 219, B: 219, A: 255}},   // White

	{Name: "Black Wool", Color: color.RGBA{R: 21, G: 22, B: 26, A: 255}},    // Black
	{Name: "Obsidian", Color: color.RGBA{R: 16, G: 12, B: 26, A: 255}},      // Black
	{Name: "Coal Block", Color: color.RGBA{R: 16, G: 16, B: 16, A: 255}},    // Black
	{Name: "Black Concrete", Color: color.RGBA{R: 8, G: 10, B: 15, A: 255}}, // Black
}

func closestColor(c color.Color) string {
	minDistance := math.MaxFloat64
	closestColor := ""

	r, g, b, _ := c.RGBA()
	for _, block := range MinecraftPalette {
		distance := colorDistance(block.Color, color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255})
		if distance < minDistance {
			minDistance = distance
			closestColor = block.Name
		}
	}
	return closestColor
}

func colorDistance(c1, c2 color.RGBA) float64 {
	dr := float64(c1.R) - float64(c2.R)
	dg := float64(c1.G) - float64(c2.G)
	db := float64(c1.B) - float64(c2.B)
	return math.Sqrt(dr*dr + dg*dg + db*db)
}

func generateHTML(img image.Image, output string) error {
	bounds := img.Bounds()
	imgWidth, imgHeight := bounds.Max.X, bounds.Max.Y
	cellSize := 100 / float64(max(imgWidth, imgHeight)) // Calculating the cell size
	pixels := make([]PixelBlock, imgWidth*imgHeight)

	for y := 0; y < imgHeight; y++ {
		for x := 0; x < imgWidth; x++ {
			c := img.At(x, y)
			r, g, b, _ := c.RGBA()
			blockName := closestColor(c)
			pixels[y*imgWidth+x] = PixelBlock{R: r >> 8, G: g >> 8, B: b >> 8, Name: blockName}
		}
	}

	tmpl := template.Must(template.New("").Parse(`
    <!DOCTYPE html>
    <html>
        <head>
            <style>
                .table { width: 100vw; height: 100vh; display: flex; flex-wrap: wrap; }
                .cell { width: {{.CellSize}}vw; height: {{.CellSize}}vh; background-color: rgb({{.R}}, {{.G}}, {{.B}}); }
            </style>
        </head>
        <body>
            <div class="table">
                {{range .Pixels}}
                <div class="cell" title="{{.Name}}" style="background-color: rgb({{.R}}, {{.G}}, {{.B}});"></div>
                {{end}}
            </div>
        </body>
    </html>
    `))

	file, err := os.Create(output)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, map[string]interface{}{
		"Pixels":   pixels,
		"CellSize": cellSize,
	})
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	// List all files in the "Input Images" directory
	files, err := ioutil.ReadDir("Input Images")
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to list images:", err)
		os.Exit(1)
	}

	// Create the "Ascii Art" directory if it doesn't exist
	if _, err := os.Stat("Ascii Art"); os.IsNotExist(err) {
		err = os.Mkdir("Ascii Art", 0755)
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to create directory:", err)
			os.Exit(1)
		}
	}

	// Process each image
	for _, file := range files {
		// Skip directories
		if file.IsDir() {
			continue
		}

		// Open the image file
		inFile, err := os.Open(path.Join("Input Images", file.Name()))
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to open image:", err)
			continue // Skip this file and try the next one
		}

		// Decode the image
		img, _, err := image.Decode(inFile)
		inFile.Close()
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to decode image:", err)
			continue // Skip this file and try the next one
		}

		// Generate the HTML file in the "Ascii Art" directory
		outputFile := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name())) + ".html"
		err = generateHTML(img, path.Join("Ascii Art", outputFile))
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to generate HTML:", err)
		}
	}
}
