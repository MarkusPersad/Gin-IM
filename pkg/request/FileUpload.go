package request

type FileUpload struct {
	UploadId    string `json:"uploadId" binding:"required" validate:"required" field_error_info:"uploadId不能为空"`
	Md5         string `json:"md5" binding:"required" validate:"required" field_error_info:"md5不能为空"`
	Sha1        string `json:"sha1" binding:"required" validate:"required" field_error_info:"sha1不能为空"`
	FileName    string `json:"fileName" binding:"required" validate:"required" field_error_info:"fileName不能为空"`
	ChunkSize   int64  `json:"chunkSize" binding:"required" validate:"required" field_error_info:"chunkSize不能为空"`
	ChunkNumber int64  `json:"chunkNumber" binding:"required" validate:"required" field_error_info:"chunkNumber不能为空"`
}
