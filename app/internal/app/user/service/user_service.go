// serviceのパッケージ
package service

import (
	"app/internal/pkg/message"
	"app/internal/pkg/model"
	"app/internal/pkg/repository"

	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/logging"
)

// UserService は、ユーザ管理業務のServiceインタフェースです。
type UserService interface {
	// Find は、userIdのユーザを照会します。
	Find(userId string) (*model.User, error)
	// Register は、ユーザ名userNameのユーザを登録します。
	Register(userName string) (*model.User, error)
}

// New は、UserServiceを作成します。
func New(logger logging.Logger,
	config config.Config,
	repository repository.UserRepository,
) UserService {
	return &userServiceImpl{logger: logger, config: config, repository: repository}
}

// userServiceImpl は、UserServiceを実装する構造体です。
type userServiceImpl struct {
	logger     logging.Logger
	config     config.Config
	repository repository.UserRepository
}

func (us *userServiceImpl) Register(userName string) (*model.User, error) {
	//設定の読み込みのとりあえずの確認
	us.logger.Debug("hoge_name=%s", us.config.Get("hoge_name", "not found"))
	us.logger.Info(message.I_EX_0002, us.config.Get("hoge_name", "not found"))

	us.logger.Debug("UserName=%s", userName)

	user := model.User{Name: userName}
	return us.repository.CreateOne(&user)
}

func (us *userServiceImpl) Find(userId string) (*model.User, error) {
	return us.repository.FindOne(userId)
}
