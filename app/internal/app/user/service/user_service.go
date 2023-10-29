// serviceのパッケージ
package service

import (
	"app/internal/pkg/entity"
	"app/internal/pkg/repository"

	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/logging"
)

// UserService は、ユーザ管理業務のServiceインタフェースです。
type UserService interface {
	// Find は、userIdのユーザを照会します。
	Find(userId string) (*entity.User, error)
	// Register は、ユーザ名userNameのユーザを登録します。
	Register(userName string) (*entity.User, error)
}

// New は、UserServiceを作成します。
func New(log logging.Logger,
	config *config.Config,
	repository repository.UserRepository,
) UserService {
	return &userServiceImpl{log: log, config: config, repository: repository}
}

// userServiceImpl は、UserServiceを実装する構造体です。
type userServiceImpl struct {
	log        logging.Logger
	config     *config.Config
	repository repository.UserRepository
}

func (us *userServiceImpl) Register(userName string) (*entity.User, error) {
	//TODO: Viperによる設定ファイルの読み込みのとりあえずの確認
	us.log.Debug("hoge.name=%s", us.config.Hoge.Name)
	us.log.Debug("UserName=%s", userName)

	user := entity.User{Name: userName}
	return us.repository.PutUser(&user)
}

func (us *userServiceImpl) Find(userId string) (*entity.User, error) {
	return us.repository.GetUser(userId)
}
