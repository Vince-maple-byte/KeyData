package sstable

type FilePath struct {
	Path string
}

func NewPath(path string) *FilePath {
	return &FilePath{Path: path}
}
