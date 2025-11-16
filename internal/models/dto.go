package models

import "io"

type PhotoData struct {
	Data        io.Reader
	FileSize    int64
	FileName    string
	ContentType string
}
