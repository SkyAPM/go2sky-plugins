// Licensed to SkyAPM org under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. SkyAPM org licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

// Package micro (sw_micro) is a plugin that can be used to trace Go-micro framework.

module go_micro

go 1.14

require (
	github.com/SkyAPM/go2sky v0.5.0
	github.com/micro/go-micro v1.18.0
)

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0
