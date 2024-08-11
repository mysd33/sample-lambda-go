/*
logging パッケージは、ログ出力に関する機能を提供するパッケージです。
*/
package logging

import "example.com/appbase/pkg/errors"

// LogError は、指定されたLoggerを使って、エラー情報をログ出力します
func LogError(log Logger, err error) {
	var (
		validationError *errors.ValidationError
		businessErrors  *errors.BusinessErrors
		systemError     *errors.SystemError
		otherError      *errors.OtherError
	)
	if errors.As(err, &validationError) {
		log.WarnWithCodableError(validationError)
	} else if errors.As(err, &businessErrors) {
		log.WarnWithMultiCodableError(businessErrors)
	} else if errors.As(err, &otherError) {
		log.WarnWithCodableError(otherError)
	} else if errors.As(err, &systemError) {
		log.ErrorWithCodableError(systemError)
	} else {
		log.ErrorWithUnexpectedError(err)
	}
}
