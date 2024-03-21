package model

type RespStrMessage struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data string `json:"data"`
}
type RespMessage struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data *Data  `json:"data"`
}
type Data struct {
	ID         int         `json:"id"`
	Username   string      `json:"username"`
	RealName   string      `json:"realName"`
	Avatar     string      `json:"avatar"`
	Gender     int         `json:"gender"`
	Email      string      `json:"email"`
	Mobile     string      `json:"mobile"`
	OrgID      int         `json:"orgId"`
	Status     int         `json:"status"`
	SuperAdmin interface{} `json:"superAdmin"`
	ProjectIds []int       `json:"projectIds"`
}
