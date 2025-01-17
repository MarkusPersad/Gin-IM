package request

type FileDelete FileMerge

type FileDeletes struct {
	Deletes []FileDelete `json:"deletes" binding:"required" validate:"required,dive" field_error_info:"删除文件列表不能为空"`
}
