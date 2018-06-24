package fourfuse

type FileAccessCallback interface {
	onAccess(file *RemoteFile)
}

type FileAccessCallbackNoop struct {
}

func NewFileAccessCallbackNoop() *FileAccessCallbackNoop {
	return &FileAccessCallbackNoop{}
}

func (f FileAccessCallbackNoop) onAccess(file *RemoteFile) {
}
