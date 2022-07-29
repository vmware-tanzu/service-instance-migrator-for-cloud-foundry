/* 
 *  Copyright 2022 VMware, Inc.
 *  
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *  http://www.apache.org/licenses/LICENSE-2.0
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package cli_test

import (
	"net/http"

	"github.com/onsi/gomega/ghttp"
)

func ConfigureTaskResult(firstHandler http.HandlerFunc, result string, server *ghttp.Server) {
	redirectHeader := http.Header{}
	redirectHeader.Add("Location", "/tasks/123")

	server.AppendHandlers(
		ghttp.CombineHandlers(
			firstHandler,
			ghttp.RespondWith(http.StatusFound, nil, redirectHeader),
		),
		ghttp.CombineHandlers(
			ghttp.VerifyRequest("GET", "/tasks/123"),
			ghttp.VerifyBasicAuth("username", "password"),
			ghttp.RespondWith(http.StatusOK, `{"id":123, "state":"done"}`),
		),
		ghttp.CombineHandlers(
			ghttp.VerifyRequest("GET", "/tasks/123"),
			ghttp.VerifyBasicAuth("username", "password"),
			ghttp.RespondWith(http.StatusOK, `{"id":123, "state":"done"}`),
		),
		ghttp.CombineHandlers(
			ghttp.VerifyRequest("GET", "/tasks/123/output", "type=event"),
			ghttp.VerifyBasicAuth("username", "password"),
			ghttp.RespondWith(http.StatusOK, ``),
		),
		ghttp.CombineHandlers(
			ghttp.VerifyRequest("GET", "/tasks/123/output", "type=result"),
			ghttp.VerifyBasicAuth("username", "password"),
			ghttp.RespondWith(http.StatusOK, result),
		),
	)
}

func AppendBadRequest(firstHandler http.HandlerFunc, server *ghttp.Server) {
	server.AppendHandlers(
		ghttp.CombineHandlers(
			firstHandler,
			ghttp.RespondWith(http.StatusBadRequest, ""),
		),
	)
}
