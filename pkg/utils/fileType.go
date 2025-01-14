package utils

func GetFileType(fileSuffix string) string {
	switch fileSuffix {
	case "jpg", "jpeg", "png", "gif", "bmp":
		return "image"
	case "mp4", "avi", "wmv", "flv", "mkv":
		return "video"
	case "mp3", "wav", "ogg":
		return "audio"
	default:
		return "file"
	}
}
func GetFileTypeEnum(fileType string) int8 {
	switch fileType {
	case "image":
		return 1
	case "video":
		return 2
	case "audio":
		return 3
	default:
		return 4
	}
}
