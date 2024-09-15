// repositoryのパッケージ
package repository

import (
	"app/internal/pkg/message"
	"app/internal/pkg/model"
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
func NewUserRepositoryForRestAPI(httpClient httpclient.HttpClient, logger logging.Logger, config config.Config) UserRepository {
	return &userRepositoryImplByRestAPI{httpClient: httpClient, logger: logger, config: config}
}

type userRepositoryImplByRestAPI struct {
	httpClient httpclient.HttpClient
	logger     logging.Logger
	config     config.Config
}

// FindOne implements UserRepository.
func (ur *userRepositoryImplByRestAPI) FindOne(userId string) (*model.User, error) {
	baseUrl, found := ur.config.GetWithContains(USERS_API_BASE_URL)
	if !found {
		return nil, errors.NewSystemError(fmt.Errorf("USERS_API_BASE_URLがありません"), message.E_EX_9001)
	}
	url := fmt.Sprintf("%s/users-api/v1/users/%s", baseUrl, userId)
	ur.logger.Debug("url:%s", url)
	// REST APIの呼び出し
	response, err := ur.httpClient.Get(url, nil, nil)
	if err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	// レスポンスデータをアンマーシャル
	var user model.User
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
func (ur *userRepositoryImplByRestAPI) CreateOne(user *model.User) (*model.User, error) {
	baseUrl, found := ur.config.GetWithContains(USERS_API_BASE_URL)
	if !found {
		return nil, errors.NewSystemError(fmt.Errorf("USERS_API_BASE_URLがありません"), message.E_EX_9001)
	}
	url := fmt.Sprintf("%s/users-api/v1/users", baseUrl)
	ur.logger.Debug("url:%s", url)
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
	var newUser model.User
	if err = json.Unmarshal(response.Body, &newUser); err != nil {
		return nil, errors.NewSystemError(err, message.E_EX_9001)
	}
	return &newUser, nil
}
