/*
idempotency パッケージは、イベントの重複によるLambdaの二重実行を防止し冪等性を担保するための機能を提供します。
*/
package idempotency

import (
	"time"

	"example.com/appbase/pkg/apcontext"
	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/date"
	"example.com/appbase/pkg/dynamodb"
	myerrors "example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/idempotency/model"
	"example.com/appbase/pkg/idempotency/tables"
	"example.com/appbase/pkg/logging"
	"example.com/appbase/pkg/message"
	"github.com/cockroachdb/errors"
)

const (
	// 冪等性管理テーブルにアイテムの有効期間（TTL）のプロパティ名
	IDEMPOTENCY_TTL_SECOUND = "IDEMPOTENCY_TTL_SECOUND"
	// 冪等性管理テーブルにアイテムの有効期間（TTL）のデフォルト値（1時間=3600秒）
	DEFAULT_IDEMPOTENCY_TTL_SECOUND = 60 * 60
)

// 二重実行防止チェック時に発生したエラー
var (
	InprogressProcessIdempotencyError = errors.New("すでに実行中の処理が存在します。")
	CompletedProcessIdempotencyError  = errors.New("完了済の処理が存在します。")
)

// IdempotencyFunc は、冪等性を担保したい処理を表す関数です。
type IdempotencyFunc func() (any, error)

// IdempotencyManager は、冪等性を担保する処理を実行するためのインターフェースです。
type IdempotencyManager interface {
	// ProcessIdempotency は、同一のidemptencyKeyに対して二重実行されず冪等性を担保したい処理を実行します。
	ProcessIdempotency(idempotencyKey string, idempotencyFunc IdempotencyFunc) (any, error)
}

// NewIdempotencyManager は、IdempotencyManagerを作成します。
func NewIdempotencyManager(logger logging.Logger, dateManager date.DateManager,
	config config.Config, repository IdempotencyRepository) IdempotencyManager {
	return &defaultIdempotencyManager{
		logger:      logger,
		dateManager: dateManager,
		config:      config,
		repository:  repository,
	}
}

// defaultIdempotencyManager は、IdempotencyManagerのデフォルト実装です。
type defaultIdempotencyManager struct {
	logger      logging.Logger
	dateManager date.DateManager
	config      config.Config
	repository  IdempotencyRepository
}

// ProcessIdempotency implements IdempotencyManager.
func (i *defaultIdempotencyManager) ProcessIdempotency(idempotencyKey string, idempotencyFunc IdempotencyFunc) (any, error) {
	// 冪等性の管理を開始
	err := i.startProcessIdepotency(idempotencyKey)
	if err != nil {
		return nil, err
	}
	// 冪等性を担保したい処理を実行
	result, err := idempotencyFunc()
	if err != nil {
		// エラー発生時は、冪等性テーブルのアイテムを削除
		derr := i.deleteIdempotencyItem(idempotencyKey)
		if derr != nil {
			// 削除に失敗した場合は、エラーを返却
			return nil, derr
		}
		return nil, err
	}
	// 冪等性の管理を完了
	err = i.completeProcessIdepotency(idempotencyKey)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// startProcessIdepotency は、冪等性の管理を開始します。
func (i *defaultIdempotencyManager) startProcessIdepotency(idempotencyKey string) error {
	// 処理中状態で冪等性テーブルにアイテム保存
	err := i.saveIdempotencyInprogress(idempotencyKey)
	if err != nil {
		// キー重複エラーの場合は二重実行とみなす
		if errors.Is(err, dynamodb.ErrKeyDuplicaiton) {
			// 既存のアイテムのステータスを取得
			item, ferr := i.repository.FindOne(idempotencyKey)
			if ferr != nil {
				// 取得に失敗した場合には、そのエラーを返却
				return ferr
			}
			if item.Status == tables.STATUS_COMPLETE {
				// 処理完了の場合の冪等性エラーを返却
				return myerrors.NewOtherError(CompletedProcessIdempotencyError, message.W_FW_8007, idempotencyKey)
			} else {
				// 処理中の場合の冪等性エラーを返却
				return myerrors.NewOtherError(InprogressProcessIdempotencyError, message.W_FW_8007, idempotencyKey)
			}
		}
		return err
	}
	return nil
}

// completeProcessIdepotency は、冪等性の管理を完了します。
func (i *defaultIdempotencyManager) completeProcessIdepotency(idempotencyKey string) error {
	err := i.updateIdempotencyItemComplete(idempotencyKey)
	if err != nil {
		// エラー発生の場合は、アイテムを削除してエラーを返却
		derr := i.deleteIdempotencyItem(idempotencyKey)
		if derr != nil {
			// 削除に失敗した場合は、エラーを返却
			return derr
		}
		return err
	}
	return nil
}

// saveIdempotencyInprogress は、冪等性の管理情報を処理中状態で保存します。
func (i *defaultIdempotencyManager) saveIdempotencyInprogress(idempotencyKey string) error {
	// 有効期限を取得
	expiry := i.getExpiry()
	// 処理中状態の有効期限を取得
	inprogressExpiry := i.getInprogressExpiryInMillis()

	item := &model.IdempotencyItem{
		IdempotencyKey:   idempotencyKey,
		Expiry:           expiry,
		InprogressExpiry: inprogressExpiry,
		Status:           tables.STATUS_INPROGRESS,
	}
	// 冪等性管理テーブルにアイテム保存
	err := i.repository.CreateOne(item)
	if err != nil {
		return err
	}
	return nil
}

// updateIdempotencyItemComplete は、冪等性の管理情報を完了状態に更新します。
func (i *defaultIdempotencyManager) updateIdempotencyItemComplete(idempotencyKey string) error {
	// 有効期限を取得
	expiry := i.getExpiry()

	item := &model.IdempotencyItem{
		IdempotencyKey: idempotencyKey,
		Expiry:         expiry,
		Status:         tables.STATUS_COMPLETE,
	}
	// 冪等性管理テーブルのアイテムのステータスを更新
	err := i.repository.UpdateOne(item)
	if err != nil {
		return err
	}
	return nil
}

// deleteIdempotencyItem は、冪等性の管理情報を削除します。
func (i *defaultIdempotencyManager) deleteIdempotencyItem(idempotencyKey string) error {
	// 冪等性管理テーブルのアイテムを削除
	err := i.repository.DeleteOne(idempotencyKey)
	if err != nil {
		return err
	}
	return nil
}

// getExpiry は、有効期限を取得します。
func (i *defaultIdempotencyManager) getExpiry() int64 {
	now := i.dateManager.GetSystemDate()

	expiresAfterSeconds := i.config.GetInt(IDEMPOTENCY_TTL_SECOUND, DEFAULT_IDEMPOTENCY_TTL_SECOUND)
	period := time.Duration(expiresAfterSeconds) * time.Second
	expiry := now.Add(period).Unix()
	i.logger.Debug("有効期限: %ds", expiry)
	return expiry
}

// getInprogressExpiryInMillis は、処理中状態の有効期限をミリ秒で取得します。
func (i *defaultIdempotencyManager) getInprogressExpiryInMillis() int64 {
	var expiry int64
	now := i.dateManager.GetSystemDate()
	remainingTimeInMillis := i.getRemainingTimeInMillis()
	i.logger.Debug("Lambdaの残り処理時間: %dms", remainingTimeInMillis)
	if remainingTimeInMillis > 0 {
		period := time.Duration(remainingTimeInMillis) * time.Millisecond
		expiry = now.Add(period).Unix() * 1000
	} else {
		// remainingTimeInMillisを取得できなかった場合は、現在時刻をそのまま有効期限とする
		i.logger.Warn(message.W_FW_8006, remainingTimeInMillis)
		expiry = now.Unix() * 1000
	}
	i.logger.Debug("処理中状態の有効期限: %dms", expiry)
	return expiry
}

// getRemainingTimeInMillis は、Lambdaの残り処理時間をミリ秒で取得します。
func (i *defaultIdempotencyManager) getRemainingTimeInMillis() int64 {
	ctx := apcontext.Context
	if ctx == nil {
		// contextが取得できなかった場合は0を返す
		return 0
	}
	deadline, ok := ctx.Deadline()
	if !ok {
		// deadlineが取得できなかった場合は0を返す
		return 0
	}
	remainingTime := time.Until(deadline).Milliseconds()
	return remainingTime
}
