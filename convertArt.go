package main

import (
	"fmt"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/nfnt/resize"
	"html/template"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"math"
	"os"
	"path"
	"strings"
)

// Define the color blocks to use in the HTML
var blocks = map[string]colorful.Color{
	"Black Concrete":      colorful.Color{R: 0.078, G: 0.078, B: 0.078},
	"Red Concrete":        colorful.Color{R: 0.643, G: 0.224, B: 0.224},
	"Green Concrete":      colorful.Color{R: 0.459, G: 0.478, B: 0.224},
	"Brown Concrete":      colorful.Color{R: 0.545, G: 0.271, B: 0.204},
	"Blue Concrete":       colorful.Color{R: 0.224, G: 0.235, B: 0.545},
	"Purple Concrete":     colorful.Color{R: 0.392, G: 0.224, B: 0.518},
	"Cyan Concrete":       colorful.Color{R: 0.224, G: 0.569, B: 0.518},
	"Light Gray Concrete": colorful.Color{R: 0.518, G: 0.518, B: 0.518},
	"Gray Concrete":       colorful.Color{R: 0.318, G: 0.318, B: 0.318},
	"Pink Concrete":       colorful.Color{R: 0.929, G: 0.459, B: 0.576},
	"Lime Concrete":       colorful.Color{R: 0.596, G: 0.745, B: 0.216},
	"Yellow Concrete":     colorful.Color{R: 0.937, G: 0.745, B: 0.188},
	"Light Blue Concrete": colorful.Color{R: 0.345, G: 0.435, B: 0.576},
	"Magenta Concrete":    colorful.Color{R: 0.620, G: 0.224, B: 0.518},
	"Orange Concrete":     colorful.Color{R: 0.882, G: 0.380, B: 0.173},
	"White Concrete":      colorful.Color{R: 0.945, G: 0.945, B: 0.945},
}

// Function to find the closest Minecraft block color to a given color.
func closestColor(c color.Color) string {
	targetColor, _ := colorful.MakeColor(c)

	var (
		closestBlock string
		minDiff      float64 = math.MaxFloat64
	)

	for blockName, blockColor := range blocks {
		diff := targetColor.DistanceCIE76(blockColor)
		if diff < minDiff {
			minDiff = diff
			closestBlock = blockName
		}
	}

	return closestBlock
}

// Generate HTML for each pixel by finding the closest color from our Minecraft palette
func generateHTML(img image.Image, output string) error {
	bounds := img.Bounds()

	// Initialize the 2D slice of pixel block names
	pixels := make([][]string, bounds.Dy())
	for i := range pixels {
		pixels[i] = make([]string, bounds.Dx())
	}

	// Convert each pixel color to the closest Minecraft block color
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			color, _ := colorful.MakeColor(img.At(x, y))
			block := closestColor(color)
			pixels[y][x] = block
		}
	}

	// Create the multiply255 function
	multiply255 := template.FuncMap{
		"multiply255": func(c colorful.Color) string {
			r, g, b := c.RGB255()
			return fmt.Sprintf("%d,%d,%d", r, g, b)
		},
	}

	// Create a new template to add the function to the function map
	tmpl := template.New("template").Funcs(multiply255)
	// Parse the template string instead of file
	tmpl, err := tmpl.Parse(`<!DOCTYPE html>
<html>
<head>
	<title>Minecraft Pixel Art</title>
	<style>
		table {
			border-collapse: collapse;
		}
		td {
			width: 10px;
			height: 10px;
		}
	</style>
</head>
<body>
	<table>
		{{range .Pixels}}
		<tr>
			{{range .}}
			<td title="{{.}}" style="background-color:rgb({{index $.ColorMap . | multiply255}})"></td>
			{{end}}
		</tr>
		{{end}}
	</table>
</body>
</html>`)
	if err != nil {
		return err
	}

	// Create the output file
	outFile, err := os.Create(output)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Execute the template with the pixel data and color map
	return tmpl.Execute(outFile, map[string]interface{}{
		"Pixels":   pixels,
		"ColorMap": blocks,
	})
}

// Main function
func main() {
	// Read the directory
	files, err := ioutil.ReadDir("Input Images")
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to list images:", err)
		os.Exit(1)
	}

	// Create the output directory if it doesn't exist
	if _, err := os.Stat("Ascii Art"); os.IsNotExist(err) {
		err = os.Mkdir("Ascii Art", 0755)
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to create directory:", err)
			os.Exit(1)
		}
	}

	// Process each image
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		inFile, err := os.Open(path.Join("Input Images", file.Name()))
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to open image:", err)
			continue
		}

		img, _, err := image.Decode(inFile)
		inFile.Close()
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to decode image:", err)
			continue
		}

		// Resize the image to 128x128 while preserving the aspect ratio
		newWidth, newHeight := 128, 128
		imgWidth := img.Bounds().Max.X
		imgHeight := img.Bounds().Max.Y
		if imgWidth > imgHeight {
			newHeight = 128 * imgHeight / imgWidth
		} else {
			newWidth = 128 * imgWidth / imgHeight
		}
		img = resize.Resize(uint(newWidth), uint(newHeight), img, resize.Lanczos3)

		outputFile := path.Join("Ascii Art", strings.TrimSuffix(file.Name(), path.Ext(file.Name()))+".html")
		err = generateHTML(img, outputFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to generate HTML:", err)
		}
	}
}
