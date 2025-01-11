package enums

type Status int8

const (
	LogIn Status = iota
	LogOut
	Forbid
)
