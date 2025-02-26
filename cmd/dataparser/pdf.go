package dataparser

import (
	"bytes"
	"github.com/dslipak/pdf"
)


func ReadPdfFile()(string,error){
	content, err := readPdf("cmd/sourcedata/RK.pdf") // Read local pdf file
	if err != nil {
		return "",err
	}
	return content,nil;
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

