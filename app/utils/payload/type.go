package payload

type ResponsePayLoad struct {
	Code   int         `json:"code"`
	Desc   string      `json:"desc"`
	Values interface{} `json:"values"`
}

func NewResponsePayLoad(err error, obj interface{}) *ResponsePayLoad {
	if err == nil {
		return &ResponsePayLoad{
			Code:   1,
			Desc:   "ok",
			Values: obj,
		}
	}
	return &ResponsePayLoad{
		Code:   0,
		Desc:   err.Error(),
		Values: obj,
	}
}
