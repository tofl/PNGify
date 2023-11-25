package image

import (
	"bytes"
	"compress/zlib"
	crand "crypto/rand"
	"encoding/binary"
	"hash/crc32"
	"math"
	"os"
)

func concatenateSlices[T any](slices ...[]T) []T {
	var tmp []T

	for _, next := range slices {
		tmp = append(tmp, next...)
	}

	return tmp
}

type Image struct {
	text              []byte
	extraBytes        int
	squareLengthBytes int
	chunkSignature    []byte
	chunkIhdr         []byte
	chunkText         [][]byte
	chunkIdat         [][]byte
	chunkIend         []byte
}

func NewImage(text []byte) *Image {
	image := Image{
		text:           text,
		chunkSignature: []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
		chunkIend:      []byte{0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82},
	}

	image.makeIhdr()

	extraBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(extraBytes, uint32(image.extraBytes))
	image.makeText([]byte("xtra"), extraBytes)

	image.makeIdat()

	return &image
}

func (i *Image) makeIhdr() {
	additionalBytes := 0
	if len(i.text)%3 > 0 {
		additionalBytes = 3 - (len(i.text) % 3)
	}

	// Dimension of a square (length of one side)
	pixelsLength := math.Sqrt(float64(len(i.text)+additionalBytes) / 3)
	pixelsLength = math.Ceil(pixelsLength)
	i.squareLengthBytes = int(pixelsLength) * 3

	// The number of bytes added to the original []byte
	deltaBytes := int((math.Pow(pixelsLength, 2) * 3) - float64(len(i.text)))
	i.extraBytes = deltaBytes

	randomBytes := make([]byte, deltaBytes)
	_, _ = crand.Read(randomBytes)
	i.text = concatenateSlices(i.text, randomBytes)

	/*
		--- IHDR ---
		Length: 13 (4 bytes)
		...
		Type: IHDR (4 bytes)
		...
		Width: (4 bytes)
		Height: (4 bytes)
		Bit depth: 8 (1 byte)
		Color type: 2 -> RGB (1 byte)
		Compression method: 0 (1 byte)
		Filter method: 0 (1 byte)
		Interlace method: 0 (1 byte)
		...
		CRC: (4 bytes)
	*/
	ihdrLength := []byte{0x00, 0x00, 0x00, 0x0D}
	ihdrType := []byte{0x49, 0x48, 0x44, 0x52}

	ihdrWidth := make([]byte, 4)
	ihdrHeight := make([]byte, 4)
	binary.BigEndian.PutUint32(ihdrWidth, uint32(pixelsLength))
	binary.BigEndian.PutUint32(ihdrHeight, uint32(pixelsLength))

	ihdrBitDepth := []byte{0x08}
	ihdrColorType := []byte{0x02}
	ihdrCompressionMethod := []byte{0x00}
	ihdrFilterMethod := []byte{0x00}
	ihdrInterlaceMethod := []byte{0x00}

	full := concatenateSlices(ihdrLength, ihdrType, ihdrWidth, ihdrHeight, ihdrBitDepth, ihdrColorType, ihdrCompressionMethod, ihdrFilterMethod, ihdrInterlaceMethod)
	ihdrCRC := crc32.ChecksumIEEE(full[4:])
	ihdrCRCByte := make([]byte, 4)
	binary.BigEndian.PutUint32(ihdrCRCByte, ihdrCRC)

	i.chunkIhdr = concatenateSlices(full, ihdrCRCByte)
}

func (i *Image) makeText(textName, text []byte) {
	textType := []byte{0x74, 0x45, 0x58, 0x54}

	textData := concatenateSlices(textName, []byte{0x00}, text)

	textLength := make([]byte, 4)
	binary.BigEndian.PutUint32(textLength, uint32(len(textData)))

	full := concatenateSlices(textLength, textType, textData)
	textCRC := crc32.ChecksumIEEE(full[4:])
	textCRCByte := make([]byte, 4)
	binary.BigEndian.PutUint32(textCRCByte, textCRC)

	i.chunkText = append(i.chunkText, concatenateSlices(full, textCRCByte))
}

func (i *Image) makeIdat() {

	var scanlines []byte
	for el := 0; el < len(i.text); {
		scanlines = append(scanlines, 0x00)
		scanlines = append(scanlines, i.text[el:el+i.squareLengthBytes]...)

		el = el + i.squareLengthBytes
	}

	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	_, _ = w.Write(scanlines)
	_ = w.Close()
	chunkDataCompressed := b.Bytes()

	for el, chunkSize := 0, 16_000; el < len(chunkDataCompressed); {

		if len(chunkDataCompressed[el:]) < chunkSize {
			chunkSize = len(chunkDataCompressed[el:])
		}

		idatData := chunkDataCompressed[el : el+chunkSize]

		idatLength := make([]byte, 4)
		binary.BigEndian.PutUint32(idatLength, uint32(len(idatData)))

		var newIdat []byte
		newIdat = append(newIdat, idatLength...)
		newIdat = append(newIdat, []byte("IDAT")...)
		newIdat = append(newIdat, idatData...)

		idatCRC := crc32.ChecksumIEEE(newIdat[4:])
		idatCRCByte := make([]byte, 4)
		binary.BigEndian.PutUint32(idatCRCByte, idatCRC)

		newIdat = append(newIdat, idatCRCByte...)

		i.chunkIdat = append(i.chunkIdat, newIdat)

		el = el + chunkSize

	}

}

func (i *Image) MakeImage() {
	textChunks := concatenateSlices(i.chunkText...)
	dataChunks := concatenateSlices(i.chunkIdat...)

	imageData := concatenateSlices(i.chunkSignature, i.chunkIhdr, textChunks, dataChunks, i.chunkIend)

	f, err := os.Create("output.png")
	defer f.Close()

	if err != nil {
		panic("Couldn't create the output file.")
	}

	_, err = f.Write(imageData)

	if err != nil {
		panic("Couldn't create the output file.")
	}
}
