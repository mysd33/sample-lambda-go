// repositoryのパッケージ
package repository

import (
	"app/internal/pkg/code"
	"app/internal/pkg/entity"
	"encoding/json"
	"fmt"
	"os"

	"example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/httpclient"
	"example.com/appbase/pkg/logging"
)

// NewUserRepositoryForRestAPI は、REST APIのためのUserRepository実装を作成します。
func NewUserRepositoryForRestAPI(httpClient httpclient.HttpClient, log logging.Logger) UserRepository {
	return &userRepositoryImplByRestAPI{httpClient: httpClient, log: log, baseUrl: os.Getenv("USERS_API_BASE_URL")}
}

type userRepositoryImplByRestAPI struct {
	httpClient httpclient.HttpClient
	log        logging.Logger
	baseUrl    string
}

// GetUser implements UserRepository.
func (ur *userRepositoryImplByRestAPI) GetUser(userId string) (*entity.User, error) {
	url := fmt.Sprintf("%s/users-api/v1/users/%s", ur.baseUrl, userId)
	ur.log.Debug("url:%s", url)

	response, err := ur.httpClient.Get(url, nil, nil)
	if err != nil {
		return nil, errors.NewSystemError(err, code.E_EX_9001)
	}
	var user entity.User
	if err = json.Unmarshal(response.Body, &user); err != nil {
		return nil, errors.NewSystemError(err, code.E_EX_9001)
	}
	return &user, nil
}

// PutUser implements UserRepository.
func (*userRepositoryImplByRestAPI) PutUser(user *entity.User) (*entity.User, error) {
	// TODO:実装
	panic("unimplemented")
}
