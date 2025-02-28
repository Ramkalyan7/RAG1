package main

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/dslipak/pdf"
)

func (app *application) ReadPdfFile(path string) (string, error) {
	f, err := pdf.Open(path)
	defer f.Trailer().Reader().Close()

	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	b, err := f.GetPlainText()
	if err != nil {
		return "", err
	}
	buf.ReadFrom(b)
	return buf.String(), nil
}

func (app *application) CreateOverlappingChunks(textData string, windowSize, overlap int) ([]string, error) {
	//re := regexp.MustCompile(`\n\s*\n`)
	paragraphs := strings.Split(textData, ".")

	if len(paragraphs) == 0 {
		return nil, fmt.Errorf("could not find any content")
	}

	step := windowSize - overlap
	if step <= 0 {
		return nil, fmt.Errorf("WindowSize must be greater than overlap")
	}

	var chunks []string
	for i := 0; i < len(paragraphs); i += step {
		end := i + windowSize
		if end > len(paragraphs) {
			end = len(paragraphs)
		}
		// Join the paragraphs back together for each chunk.
		chunk := strings.Join(paragraphs[i:end], "\n\n")
		chunks = append(chunks, chunk)
		// If we've reached the end, break out of the loop.
		if end == len(paragraphs) {
			break
		}
	}
	return chunks, nil
}
