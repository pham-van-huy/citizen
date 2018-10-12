package httpserver

type pingReq struct {
	Value string `json:"value"`
}

type pingRes struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}
