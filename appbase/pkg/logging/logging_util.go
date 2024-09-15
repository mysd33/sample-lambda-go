/*
logging パッケージは、ログ出力に関する機能を提供するパッケージです。
*/
package logging

import "example.com/appbase/pkg/errors"

// LogError は、指定されたLoggerを使って、エラー情報をログ出力します
func LogError(logger Logger, err error) {
	var (
		validationError *errors.ValidationError
		businessErrors  *errors.BusinessErrors
		systemError     *errors.SystemError
		otherError      *errors.OtherError
	)
	if errors.As(err, &validationError) {
		logger.WarnWithCodableError(validationError)
	} else if errors.As(err, &businessErrors) {
		logger.WarnWithMultiCodableError(businessErrors)
	} else if errors.As(err, &otherError) {
		logger.WarnWithCodableError(otherError)
	} else if errors.As(err, &systemError) {
		logger.ErrorWithCodableError(systemError)
	} else {
		logger.ErrorWithUnexpectedError(err)
	}
}
