package adminJson

import (
	"encoding/json"
)

// var MarshalErrJson

type Result struct {
	ErrNo int         `json:"errno"`
	Err   string      `json:"errmsg"`
	Data  interface{} `json:"data"`
}

func FmtJson(errno int, err string, data interface{}) string {
	r := Result{
		ErrNo: errno,
		Err:   err,
		Data:  data,
	}
	v, e := json.Marshal(r)
	if e != nil {
		return `{"ErrNo":500,"Err":"Marshal failed ` + e.Error() + `","Data":""}`
	}
	return string(v)
}
