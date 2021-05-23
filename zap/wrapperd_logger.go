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

package zap

import (
	"context"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type WrapLogger struct {
	base *zap.Logger
	ctx  context.Context
}

// Ctx setting context to logger
func (log *WrapLogger) Ctx(ctx context.Context) *WrapLogger {
	return &WrapLogger{log.base, ctx}
}

// Wrap original zap logger(not sugar)
func Wrap(logger *zap.Logger) *WrapLogger {
	return &WrapLogger{logger, nil}
}

func (log *WrapLogger) Named(s string) *WrapLogger {
	return Wrap(log.base.Named(s))
}

func (log *WrapLogger) WithOptions(opts ...zap.Option) *WrapLogger {
	return Wrap(log.base.WithOptions(opts...))
}

func (log *WrapLogger) With(fields ...zap.Field) *WrapLogger {
	return Wrap(log.base.With(fields...))
}

func (log *WrapLogger) Check(lvl zapcore.Level, msg string) *zapcore.CheckedEntry {
	return log.base.Check(lvl, msg)
}

func (log *WrapLogger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	fields = log.appendContextField(ctx, fields...)
	log.base.Debug(msg, fields...)
}

func (log *WrapLogger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	fields = log.appendContextField(ctx, fields...)
	log.base.Info(msg, fields...)
}

func (log *WrapLogger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	fields = log.appendContextField(ctx, fields...)
	log.base.Warn(msg, fields...)
}

func (log *WrapLogger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	fields = log.appendContextField(ctx, fields...)
	log.base.Error(msg, fields...)
}

func (log *WrapLogger) DPanic(ctx context.Context, msg string, fields ...zap.Field) {
	fields = log.appendContextField(ctx, fields...)
	log.base.DPanic(msg, fields...)
}

func (log *WrapLogger) Panic(ctx context.Context, msg string, fields ...zap.Field) {
	fields = log.appendContextField(ctx, fields...)
	log.base.Panic(msg, fields...)
}

func (log *WrapLogger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	fields = log.appendContextField(ctx, fields...)
	log.base.Fatal(msg, fields...)
}

func (log *WrapLogger) Sync() error {
	return log.base.Sync()
}

func (log *WrapLogger) Core() zapcore.Core {
	return log.base.Core()
}

func (log *WrapLogger) appendContextField(ctx context.Context, fields ...zap.Field) []zap.Field {
	if ctx != nil {
		return append(fields, TraceContext(ctx)...)
	}
	if log.ctx != nil {
		return append(fields, TraceContext(ctx)...)
	}
	return fields
}
