package main

import (
	"bytes"
	"fmt"

	"github.com/dslipak/pdf"
)


func main() {
	content, err := readPdf("cmd/sourcedata/RK.pdf") // Read local pdf file
	if err != nil {
		panic(err)
	}
	fmt.Println(content)
}

func readPdf(path string) (string, error) {
	f , err := pdf.Open(path)
	// remember close file
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

