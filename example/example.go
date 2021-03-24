package example

import (
	_ "embed"
)

const (
	serverURL   = "https://www.planttext.com/api/plantuml"
	encodedPUML = "SYWkIImgAStDKN2jICmjo4dbSifFKj2rKt3CoKnELR1Io4ZDoSddSaZDIodDpG44003__m00"
)

//go:embed example.png
var examplePNG string

// PNGFile ...
func PNGFile() string { return examplePNG }

// PNGLink ...
func PNGLink() string { return serverURL + "/png/" + encodedPUML }

//go:embed example.puml
var examplePUML string

// PUML ...
func PUML() string { return examplePUML }

//go:embed example.svg
var exampleSVG string

// SVGFile ...
func SVGFile() string { return exampleSVG }

// SVGLink ...
func SVGLink() string { return serverURL + "/svg/" + encodedPUML }

//go:embed example.txt
var exampleTXT string

// TXTFile ...
func TXTFile() string { return exampleTXT }

// TXTLink ...
func TXTLink() string { return serverURL + "/txt/" + encodedPUML }
