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

package zap

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// WrapLogger is wrap the zap logger, also contains all method at zap logger, correlate the context before logging
type WrapLogger struct {
	base *zap.Logger
}

// WrapWithContext original zap logger(not sugar)
func WrapWithContext(logger *zap.Logger) *WrapLogger {
	return &WrapLogger{logger}
}

// Named wrap the zap logging function
func (log *WrapLogger) Named(s string) *WrapLogger {
	return WrapWithContext(log.base.Named(s))
}

// WithOptions wrap the zap logging function
func (log *WrapLogger) WithOptions(opts ...zap.Option) *WrapLogger {
	return WrapWithContext(log.base.WithOptions(opts...))
}

// With wrap the zap logging function
func (log *WrapLogger) With(fields ...zap.Field) *WrapLogger {
	return WrapWithContext(log.base.With(fields...))
}

// Check wrap the zap logging function
func (log *WrapLogger) Check(lvl zapcore.Level, msg string) *zapcore.CheckedEntry {
	return log.base.Check(lvl, msg)
}

// Debug wrap the zap logging function and relate the context
func (log *WrapLogger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	log.appendContextField(ctx, &fields)
	log.base.Debug(msg, fields...)
}

// Info wrap the zap logging function and relate the context
func (log *WrapLogger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	log.appendContextField(ctx, &fields)
	log.base.Info(msg, fields...)
}

// Warn wrap the zap logging function and relate the context
func (log *WrapLogger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	log.appendContextField(ctx, &fields)
	log.base.Warn(msg, fields...)
}

// Error wrap the zap logging function and relate the context
func (log *WrapLogger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	log.appendContextField(ctx, &fields)
	log.base.Error(msg, fields...)
}

// DPanic wrap the zap logging function and relate the context
func (log *WrapLogger) DPanic(ctx context.Context, msg string, fields ...zap.Field) {
	log.appendContextField(ctx, &fields)
	log.base.DPanic(msg, fields...)
}

// Panic wrap the zap logging function and relate the context
func (log *WrapLogger) Panic(ctx context.Context, msg string, fields ...zap.Field) {
	log.appendContextField(ctx, &fields)
	log.base.Panic(msg, fields...)
}

// Fatal wrap the zap logging function and relate the context
func (log *WrapLogger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	log.appendContextField(ctx, &fields)
	log.base.Fatal(msg, fields...)
}

// Sync wrap the zap logging function
func (log *WrapLogger) Sync() error {
	return log.base.Sync()
}

// Core wrap the zap logging function
func (log *WrapLogger) Core() zapcore.Core {
	return log.base.Core()
}

func (log *WrapLogger) appendContextField(ctx context.Context, fields *[]zap.Field) {
	if ctx != nil {
		*fields = append(*fields, TraceContext(ctx)...)
	}
}
