/*
transaction パッケージは、トランザクション管理に関する機能を提供するパッケージです。
*/
package transaction

import "example.com/appbase/pkg/transaction/model"

// MessageRegisterer は、メッセージをトランザクションに登録するためのインターフェースです。
type MessageRegisterer interface {
	// メッセージ情報を登録
	RegisterMessage(queueMessage *model.QueueMessageItem) error
	// メッセージ情報のステータスを追加更新
	UpdateMessage(queueMessage *model.QueueMessageItem) error
}

// NewMessageRegisterer は、MessageRegistererを作成します。
func NewMessageRegisterer(queueMessageItemRepository QueueMessageItemRepository) MessageRegisterer {
	return &defaultMessageRegisterer{
		queueMessageItemRepository: queueMessageItemRepository,
	}
}

// defaultMessageRegisterer は、MessageRegistererの実装です。
type defaultMessageRegisterer struct {
	queueMessageItemRepository QueueMessageItemRepository
}

// RegisterMessage implements MessageRegisterer.
func (r *defaultMessageRegisterer) RegisterMessage(queueMessage *model.QueueMessageItem) error {
	return r.queueMessageItemRepository.CreateOneWithTx(queueMessage)
}

// UpdateMessage implements MessageRegisterer.
func (r *defaultMessageRegisterer) UpdateMessage(queueMessage *model.QueueMessageItem) error {
	return r.queueMessageItemRepository.UpdateOneWithTx(queueMessage)
}
