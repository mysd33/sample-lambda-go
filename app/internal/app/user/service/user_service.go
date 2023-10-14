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
	// Regist は、ユーザ名userNameのユーザを登録します。
	Regist(userName string) (*entity.User, error)
}

// New は、UserServiceを作成します。
func New(log logging.Logger,
	config *config.Config,
	repository repository.UserRepository,
) UserService {
	return &UserServiceImpl{log: log, config: config, repository: repository}
}

// UserServiceImpl は、UserServiceを実装する構造体です。
type UserServiceImpl struct {
	log        logging.Logger
	config     *config.Config
	repository repository.UserRepository
}

func (us *UserServiceImpl) Regist(userName string) (*entity.User, error) {
	//TODO: Viperによる設定ファイルの読み込みのとりあえずの確認
	us.log.Info("hoge.name=%s", us.config.Hoge.Name)

	//Zapによるログ出力の例
	us.log.Info("UserName=%s", userName)

	user := entity.User{}
	user.Name = userName
	return us.repository.PutUser(&user)
}

func (us *UserServiceImpl) Find(userId string) (*entity.User, error) {
	return us.repository.GetUser(userId)
}
