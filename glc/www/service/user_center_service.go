package service

import (
	"encoding/json"
	"fmt"
	"github.com/gotoeasy/glang/cmn"
	"glc/conf"
	"glc/www/model"
)

type Service interface {
	GetCenterToken() (*model.RespStrMessage, error)
	GetCenterUser(apiToken string, userToken string) (*model.RespMessage, error)
}

func NewService() Service {
	return &service{}
}

type service struct {
}

func commonResponse(body []byte) (*model.RespMessage, error) {
	var result *model.RespMessage
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result, nil
}
func commonStrResponse(body []byte) (*model.RespStrMessage, error) {
	var result *model.RespStrMessage
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result, nil
}
func (h *service) GetCenterToken() (*model.RespStrMessage, error) {
	url := fmt.Sprintf("/data-service/api/token/generate?appKey=%s&appSecret=%s", conf.GetUcAppKey(), conf.GetUcAppSecret())
	body, err1 := cmn.HttpGetJson(conf.GetUcAppUrl() + url)
	if err1 != nil {
		return nil, err1
	}
	return commonStrResponse(body)
}
func (h *service) GetCenterUser(apiToken string, userToken string) (*model.RespMessage, error) {
	url := fmt.Sprintf("/data-service/api/sys/user/userinfo")
	cmn.Info("url:" + conf.GetUcAppUrl() + url)
	cmn.Info("apiToken:" + apiToken + "||Authorization:" + userToken)
	body, err := cmn.HttpPostJson(conf.GetUcAppUrl()+url, "", "Authorization:"+userToken, "Apitoken:"+apiToken)
	if err != nil {
		return nil, err
	}
	return commonResponse(body)
}
