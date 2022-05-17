package utils

import "fmt"

const (
	PackageName    = "fairdatasociety.fairpass"
	PodName        = "fairpass"
	PasswordsTable = "passwords"
	NotesTable     = "notes"
	AppName        = "FairPass"
	LoginWelcome   = "FairPass"
)

var (
	ErrBlankUsername        = fmt.Errorf("username cannot be blank")
	ErrBlankPassword        = fmt.Errorf("password cannot be blank")
	ErrBlankConfirmPassword = fmt.Errorf("confirm password cannot be blank")
	ErrPasswordMismatch     = fmt.Errorf("passwords do not match")
	ErrBlankBee             = fmt.Errorf("please enter bee endpoint")
	ErrBlankBatchId         = fmt.Errorf("please enter batch ID")
	ErrBlankRPC             = fmt.Errorf("please enter rpc endpoint")
)
