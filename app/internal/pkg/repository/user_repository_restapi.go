// repositoryのパッケージ
package repository

import (
	"app/internal/pkg/entity"
	"app/internal/pkg/message"
	"encoding/json"
	"fmt"

	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/errors"
	"example.com/appbase/pkg/httpclient"
	"example.com/appbase/pkg/logging"
)

const (
	USERS_API_BASE_URL = "USERS_API_BASE_URL"
)

// NewUserRepositoryForRestAPI は、REST APIのためのUserRepository実装を作成します。
func NewUserRepositoryForRestAPI(httpClient httpclient.HttpClient, log logging.Logger, config config.Config) UserRepository {
	return &userRepositoryImplByRestAPI{httpClient: httpClient, log: log, config: config}
}

type userRepositoryImplByRestAPI struct {
	httpClient httpclient.HttpClient
	log        logging.Logger
	config     config.Config
}

// FindOne implements UserRepository.
func (ur *userRepositoryImplByRestAPI) FindOne(userId string) (*entity.User, error) {
	baseUrl, found := ur.config.GetWithContains(USERS_API_BASE_URL)
	if !found {
		return nil, errors.NewSystemError(fmt.Errorf("USERS_API_BASE_URLがありません"), message.E_EX_9001)
	}
	url := fmt.Sprintf("%s/users-api/v1/users/%s", baseUrl, userId)
	ur.log.Debug("url:%s", url)
	// REST APIの呼び出し
	response, err := ur.httpClient.Get(url, nil, nil)
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	// レスポンスデータをアンマーシャル
	var user entity.User
	if err = json.Unmarshal(response.Body, &user); err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	if response.StatusCode != 200 {
		// TODO: 200以外の処理
		return nil, errors.NewBusinessError(message.W_EX_8001, "xxxx")
	}
	return &user, nil
}

// CreateOne implements UserRepository.
func (ur *userRepositoryImplByRestAPI) CreateOne(user *entity.User) (*entity.User, error) {
	baseUrl, found := ur.config.GetWithContains(USERS_API_BASE_URL)
	if !found {
		return nil, errors.NewSystemError(fmt.Errorf("USERS_API_BASE_URLがありません"), message.E_EX_9001)
	}
	url := fmt.Sprintf("%s/users-api/v1/users", baseUrl)
	ur.log.Debug("url:%s", url)
	// リクエストデータをアンマーシャル
	data, err := json.MarshalIndent(user, "", "    ")
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	// REST APIの呼び出し
	response, err := ur.httpClient.Post(url, nil, data)
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	if response.StatusCode != 200 {
		// TODO: 200以外の処理
		return nil, errors.NewBusinessError(message.W_EX_8001, "xxxx")
	}
	// レスポンスデータをアンマーシャル
	var newUser entity.User
	if err = json.Unmarshal(response.Body, &newUser); err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	return &newUser, nil
}
