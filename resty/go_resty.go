//
// Copyright 2022 SkyAPM org
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package resty

import (
	"log"

	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/plugins/http"
	"github.com/go-resty/resty/v2"
)

// NewGoResty returns a resty Client with tracer
func NewGoResty(tracer *go2sky.Tracer, options ...http.ClientOption) *resty.Client {
	hc, err := http.NewClient(tracer, options...)
	if err != nil {
		log.Fatalf("create client error %v \n", err)
	}

	return resty.NewWithClient(hc)
}
