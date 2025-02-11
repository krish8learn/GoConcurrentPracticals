package digitalsig

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
)

// sha1sig return SHA1 signature in the format "35aabcd5a32e01d18a5ef688111624f3c547e13b"
func sha1Sig(data []byte) (string, error) {
	w := sha1.New()
	r := bytes.NewReader(data)
	if _, err := io.Copy(w, r); err != nil {
		return "", err
	}

	sig := fmt.Sprintf("%x", w.Sum(nil))
	return sig, nil
}

type File struct {
	Name      string
	Content   []byte
	Signature string
}

// instance of this object will store the result from the channel
type OperationResult struct {
	FileName string
	Matched  bool
	Err      error
}

// this function will give struture to call the Sha1Sig as Go-routines
func SigGoWorker(data File, res chan<- OperationResult) {
	sig, err := sha1Sig(data.Content)
	ops := OperationResult{
		FileName: data.Name,
		Matched:  data.Signature == sig,
		Err:      err,
	}
	res <- ops
}

// ValidateSigs return slice of OK files and slice of mismatched files
func ValidateSigs(files []File) ([]string, []string, error) {
	var okFiles []string
	var badFiles []string

	result := make(chan OperationResult)

	// fan out
	for _, file := range files {
		go SigGoWorker(file, result)
	}

	for range files {
		gotRes := <-result
		if !gotRes.Matched || gotRes.Err != nil {
			badFiles = append(badFiles, gotRes.FileName)
		} else {
			okFiles = append(okFiles, gotRes.FileName)
		}
	}

	return okFiles, badFiles, nil
}
