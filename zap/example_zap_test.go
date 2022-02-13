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
package zap_test

import (
	"context"

	zapplugin "github.com/SkyAPM/go2sky-plugins/zap"
	"go.uber.org/zap"
)

func ExampleTraceContext() {
	ctx := context.Background()

	logger := zap.NewExample()
	logger.With(zapplugin.TraceContext(ctx)...).Info("test")
	// Output:
	// {"level":"info","msg":"test","SW_CTX":"[,,N/A,N/A,-1]"}
}

func ExampleWrapWithContext() {
	ctx := context.Background()

	logger := zapplugin.WrapWithContext(zap.NewExample())
	logger.Info(ctx, "test")
	// Output:
	// {"level":"info","msg":"test","SW_CTX":"[,,N/A,N/A,-1]"}
}
