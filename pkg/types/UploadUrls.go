package types

type UploadUrls struct {
	UploadId  string   `json:"uploadId"`
	Urls      []string `json:"urls"`
	Completed []string `json:"completed"`
}
