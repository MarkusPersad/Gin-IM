package request

type FileUpload struct {
	Md5         string `json:"md5" binding:"required" validate:"required" field_error_info:"md5不能为空"`
	Sha1        string `json:"sha1" binding:"required" validate:"required" field_error_info:"sha1不能为空"`
	FileName    string `json:"fileName" binding:"required" validate:"required" field_error_info:"fileName不能为空"`
	ChunkSize   uint64 `json:"chunkSize" binding:"required" validate:"required" field_error_info:"chunkSize不能为空"`
	ChunkNumber int    `json:"chunkNumber" binding:"required" validate:"required,min=1" field_error_info:"chunkNumber最小为1"`
}
