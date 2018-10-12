// Copyright 2018 https://gophersland.com
// All rights reserved.
// Use of this source code is governed by an Apache License that can be found in the LICENSE file.
package httpserver

type pingReq struct {
	Value string `json:"value"`
}

type pingRes struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}
