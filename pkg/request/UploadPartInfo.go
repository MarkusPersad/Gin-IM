package request

type PartInfo struct {
	UploadId string `json:"uploadId" binding:"required" validate:"required" field_error_info:"uploadId不能为空"`
	PartNums string `json:"partNums" binding:"required" validate:"required" field_error_info:"partNums不能为空"`
}
