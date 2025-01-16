package types

type UploadInfo struct {
	UploadId    string
	ObjectName  string
	ChunkSize   uint64
	ChunkNumber int
	ContentType string
}
