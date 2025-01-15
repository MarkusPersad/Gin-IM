package request

type FileDelete struct {
	Seleted []uint `json:"seleted" binding:"required" validate:"required" field_error_info:"请选择文件"`
}
