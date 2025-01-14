package defines

const (
	Timeout              = 500
	PASSWORD_REGEX       = `^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[A-Za-z\d]{8,32}$`
	FIELD_ERROR_INFO     = "field_error_info"
	CAPTCHA              = "captcha:"
	CAPTCHA_TIMEOUT      = 5 * 60
	TOKEN_EXPIRE         = 24
	USER_TOKEN_KEY       = "user_token:"
	USER_TOKEN           = 60 * 60 * 24
	MESSAGE_SEND_TIMEOUT = 5
	SYSTEM_MESSAGE       = "system"
	FILE_SHORT_SIGN      = 24
)
