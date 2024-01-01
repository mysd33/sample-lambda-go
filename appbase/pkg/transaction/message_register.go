/*
transaction パッケージは、トランザクション管理に関する機能を提供するパッケージです。
*/
package transaction

import "example.com/appbase/pkg/transaction/entity"

// MessageRegisterer は、メッセージをトランザクションに登録するためのインターフェースです。
type MessageRegisterer interface {
	RegisterMessage(transaction Transaction, queueMessage *entity.QueueMessageItem) error
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
func (r *defaultMessageRegisterer) RegisterMessage(transaction Transaction, queueMessage *entity.QueueMessageItem) error {
	return r.queueMessageItemRepository.CreateOneWithTx(queueMessage)
}
