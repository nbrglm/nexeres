package models

type BackupCodes []BackupCode

type BackupCode struct {
	Code string `json:"code"`
	Used bool   `json:"used"`
}
