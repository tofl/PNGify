package main

import (
	"flag"
	"fmt"
	"github.com/tofl/PNG-encode-decode/image"
	"os"
	"time"
)

func main() {
	encodeCmd := flag.NewFlagSet("encode", flag.ExitOnError)
	encodeText := encodeCmd.String("t", "", "Text to encode")
	encodeFile := encodeCmd.String("f", "", "File to encode")

	decodeCmd := flag.NewFlagSet("decode", flag.ExitOnError)
	filePath := decodeCmd.String("p", "", "Path to the file to decode")

	if len(os.Args) < 2 {
		fmt.Println("A command must be specified.")
		os.Exit(1)
	}

	executionTime := time.Now()

	switch os.Args[1] {
	case "encode":
		_ = encodeCmd.Parse(os.Args[2:])
		var img *image.Image

		if *encodeText != "" {
			// Encode text
			img = image.NewImage([]byte(*encodeText))
		} else if *encodeFile != "" {
			content, err := os.ReadFile(*encodeFile)
			if err != nil {
				fmt.Println("Couldn't read file", *encodeFile)
				os.Exit(1)
			}

			fmt.Println("Encoding file...")

			img = image.NewImage(content)
			img.MakeText([]byte("filename"), []byte(*encodeFile))
		}

		img.MakeImage()

		break
	case "decode":
		_ = decodeCmd.Parse(os.Args[2:])

		f, err := os.Open(*filePath)
		defer f.Close()

		if err != nil {
			fmt.Println("Couldn't open the file")
			os.Exit(1)
		}

		fmt.Println("Decoding image...")

		data, fileName := image.Decode(f)

		if fileName == "" {
			fmt.Println(data)
		} else {
			// Save file
			f, err := os.Create(fileName)
			defer f.Close()

			if err != nil {
				panic("Couldn't create the output file.")
			}

			_, err = f.Write([]byte(data))

			if err != nil {
				panic("Couldn't create the output file.")
			}
		}
		break
	default:
		fmt.Println("Command not recognised")
		os.Exit(1)
	}

	elapsed := time.Since(executionTime)
	fmt.Println("Done in", elapsed)
}
