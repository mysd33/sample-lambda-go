package service_test

import (
	"app/internal/app/user/service"
	"app/internal/pkg/entity"
	"app/internal/pkg/repository"
	"testing"

	"example.com/appbase/pkg/config"
	"example.com/appbase/pkg/id"
	"example.com/appbase/pkg/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	//testfyによるMockオブジェクト化
	mock.Mock
}

func (d *MockUserRepository) GetUser(userId string) (*entity.User, error) {
	//インタフェースメソッド実装時、Mock.Calledメソッド呼び出し
	args := d.Called(userId)
	return args.Get(0).(*entity.User), args.Error(1)
}

func (d *MockUserRepository) PutUser(user *entity.User) (*entity.User, error) {
	//インタフェースメソッド実装時、Mock.Calledメソッド呼び出し
	args := d.Called(user)
	return args.Get(0).(*entity.User), args.Error(1)
}

func TestRegist(t *testing.T) {
	//入力値
	inputUserName := "fuga"
	//期待値
	expectedName := "fuga"
	log := logging.NewLogger()
	cfg := &config.Config{Hoge: config.Hoge{Name: "hoge"}}

	//Mockへの入力値と戻り値の設定
	mock := new(MockUserRepository)
	mockInputValue := entity.User{Name: inputUserName}
	mockReturnValue := entity.User{ID: id.GenerateId(), Name: expectedName}
	mock.On("PutUser", &mockInputValue).Return(&mockReturnValue, nil)
	var repository repository.UserRepository = mock
	sut := service.NewUserService(log, cfg, &repository)
	//テスト対象メソッドの呼び出し
	actual, _ := sut.Regist(inputUserName)
	println(actual)
	//テスト対象メソッドのAssert
	assert.Equal(t, expectedName, actual.Name)
	//Mockで定義した入力が呼ばれたことのAssert
	mock.AssertExpectations(t)

}
