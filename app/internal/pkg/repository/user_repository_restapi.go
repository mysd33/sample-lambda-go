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
	// REST APIの呼び出し
	response, err := ur.httpClient.Get(url, nil, nil)
	if err != nil {
		return nil, errors.NewSystemError(err, code.E_EX_9001)
	}
	// レスポンスデータをアンマーシャル
	var user entity.User
	if err = json.Unmarshal(response.Body, &user); err != nil {
		return nil, errors.NewSystemError(err, code.E_EX_9001)
	}
	if response.StatusCode != 200 {
		// TODO: 200以外の処理
		return nil, errors.NewBusinessError(code.W_EX_8001, "xxxx")
	}
	return &user, nil
}

// PutUser implements UserRepository.
func (ur *userRepositoryImplByRestAPI) PutUser(user *entity.User) (*entity.User, error) {
	url := fmt.Sprintf("%s/users-api/v1/users", ur.baseUrl)
	ur.log.Debug("url:%s", url)
	// リクエストデータをアンマーシャル
	data, err := json.MarshalIndent(user, "", "    ")
	if err != nil {
		return nil, errors.NewSystemError(err, code.E_EX_9001)
	}
	// REST APIの呼び出し
	response, err := ur.httpClient.Post(url, nil, data)
	if err != nil {
		return nil, errors.NewSystemError(err, code.E_EX_9001)
	}
	if response.StatusCode != 200 {
		// TODO: 200以外の処理
		return nil, errors.NewBusinessError(code.W_EX_8001, "xxxx")
	}
	// レスポンスデータをアンマーシャル
	var newUser entity.User
	if err = json.Unmarshal(response.Body, &newUser); err != nil {
		return nil, errors.NewSystemError(err, code.E_EX_9001)
	}
	return &newUser, nil
}
