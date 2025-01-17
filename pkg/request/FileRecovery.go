package request

type FileRecovery FileMerge
type FileRecoveryList struct {
	Recoveries []FileRecovery `json:"recoveries" binding:"required" validate:"required,dive" field_error_info:"恢复文件列表不能为空"`
}
