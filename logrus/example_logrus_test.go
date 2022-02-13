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

package logrus

import (
	"context"
	"os"

	"github.com/SkyAPM/go2sky"
	"github.com/sirupsen/logrus"
)

func ExampleWrapFormat() {
	context := context.Background()

	logrus.SetOutput(os.Stdout)

	// init tracer
	_, err := go2sky.NewTracer("example")
	if err != nil {
		logrus.Fatalf("create tracer error %v \n", err)
	}

	// json format
	logrus.SetFormatter(Wrap(&logrus.JSONFormatter{
		DisableTimestamp: true,
	}, "SW_CTX"))
	logrus.WithContext(context).Info("test1")

	// test format
	logrus.SetFormatter(Wrap(&logrus.TextFormatter{
		DisableTimestamp: true,
	}, "SW_CTX"))
	logrus.WithContext(context).Info("test2")

	// Output:
	// {"SW_CTX":"[,,N/A,N/A,-1]","level":"info","msg":"test1"}
	// level=info msg=test2 SW_CTX="[,,N/A,N/A,-1]"
}
