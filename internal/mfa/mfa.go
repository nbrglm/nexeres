// Package mfa provides functionality for multi-factor authentication.
package mfa

import (
	"github.com/nbrglm/nexeres/internal/models"
	"github.com/nbrglm/nexeres/internal/otp"
)

func GenerateBackupCodes(numCodes int) (models.BackupCodes, error) {
	var backupCodes models.BackupCodes
	for range numCodes {
		code, err := otp.NewAlphaNumericOTP(10)
		if err != nil {
			return nil, err
		}
		backupCodes = append(backupCodes, models.BackupCode{
			Code: code,
			Used: false,
		})
	}
	return backupCodes, nil
}
