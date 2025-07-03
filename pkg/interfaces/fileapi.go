package interfaces

type FileName string
type URI string

type FileAPIFileSpec struct {
	LocalFilename FileName
	RemoteURI     URI
}

// A file api target at an AI provider
type FileAPI interface {
	SendFile(f *FileAPIFileSpec) error
}
