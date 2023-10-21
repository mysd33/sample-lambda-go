// repositoryのパッケージ
package repository

import (
	"app/internal/pkg/code"
	"app/internal/pkg/entity"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/logging"
)

// NewUserRepositoryForRestAPI は、REST APIのためのUserRepository実装を作成します。
func NewUserRepositoryForRestAPI(log logging.Logger) UserRepository {
	return &userRepositoryImplByRestAPI{log: log}
}

type userRepositoryImplByRestAPI struct {
	log logging.Logger
}

// GetUser implements UserRepository.
func (ur *userRepositoryImplByRestAPI) GetUser(userId string) (*entity.User, error) {
	baseUrl := os.Getenv("USERS_API_BASE_URL")
	url := fmt.Sprintf("%s/users-api/v1/users/%s", baseUrl, userId)
	ur.log.Debug("url:%s", url)

	// TODO: AP基盤機能化
	response, err := http.Get(url)
	if err != nil {
		return nil, errors.NewSystemError(err, code.E_EX_9001)
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.NewSystemError(err, code.E_EX_9001)
	}

	var user entity.User
	if err = json.Unmarshal(data, &user); err != nil {
		return nil, errors.NewSystemError(err, code.E_EX_9001)
	}
	return &user, nil
}

// PutUser implements UserRepository.
func (*userRepositoryImplByRestAPI) PutUser(user *entity.User) (*entity.User, error) {
	// TODO:実装
	panic("unimplemented")
}
