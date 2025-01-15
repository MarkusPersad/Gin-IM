package request

type FileDownload struct {
	Md5  string `json:"md5" binding:"required" validate:"required" field_error_info:"md5不能为空"`
	Sha1 string `json:"sha1" binding:"required" validate:"required" field_error_info:"sha1不能为空"`
}
