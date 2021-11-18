//
// Copyright 2021 SkyAPM org
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

package gorm

import (
	"fmt"
	"time"

	"github.com/SkyAPM/go2sky"
	"gorm.io/gorm"
	agentv3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
)

var (
	_ gorm.Plugin = &SkyWalking{}
)

const (
	componentIDUnknown = 0
	componentIDMySQL   = 5012
)

type DBType string

const (
	UNKNOWN DBType = "unknown"
	MYSQL   DBType = "mysql"
)

const spanKey = "spanKey"

type SkyWalking struct {
	tracer  *go2sky.Tracer
	peer    string
	sqlType DBType
}

func New(tracer *go2sky.Tracer, peer string, sqlType DBType) *SkyWalking {
	return &SkyWalking{tracer, peer, sqlType}
}

func (s *SkyWalking) Name() string {
	return "gorm:skyWalking"
}

func (s *SkyWalking) Initialize(db *gorm.DB) (err error) {
	// before database operation
	db.Callback().Create().Before("gorm:create").Register("sky_create_span", s.BeforeCallback("create"))
	db.Callback().Query().Before("gorm:query").Register("sky_create_span", s.BeforeCallback("query"))
	db.Callback().Update().Before("gorm:update").Register("sky_create_span", s.BeforeCallback("update"))
	db.Callback().Delete().Before("gorm:delete").Register("sky_create_span", s.BeforeCallback("delete"))
	db.Callback().Row().Before("gorm:row").Register("sky_create_span", s.BeforeCallback("row"))
	db.Callback().Raw().Before("gorm:raw").Register("sky_create_span", s.BeforeCallback("raw"))

	// after database operation
	db.Callback().Create().After("gorm:create").Register("sky_end_span", s.AfterCallback())
	db.Callback().Query().After("gorm:query").Register("sky_end_span", s.AfterCallback())
	db.Callback().Update().After("gorm:update").Register("sky_end_span", s.AfterCallback())
	db.Callback().Delete().After("gorm:delete").Register("sky_end_span", s.AfterCallback())
	db.Callback().Row().After("gorm:row").Register("sky_end_span", s.AfterCallback())
	db.Callback().Raw().After("gorm:raw").Register("sky_end_span", s.AfterCallback())

	return
}

func (s *SkyWalking) BeforeCallback(operation string) func(db *gorm.DB) {
	tracer := s.tracer
	peer := s.peer

	if tracer == nil {
		return func(db *gorm.DB) {}
	}

	return func(db *gorm.DB) {
		tableName := db.Statement.Table
		operation := fmt.Sprintf("%s/%s", tableName, operation)

		span, err := tracer.CreateExitSpan(db.Statement.Context, operation, peer, func(key, value string) error {
			return nil
		})
		if err != nil {
			db.Logger.Error(db.Statement.Context, "gorm:skyWalking failed to create exit span, got error: %v", err)
		}

		// set span from db instance's context to pass span
		db.Set(spanKey, span)
	}
}

func (s *SkyWalking) AfterCallback() func(db *gorm.DB) {
	tracer := s.tracer
	if tracer == nil {
		return func(db *gorm.DB) {}
	}

	return func(db *gorm.DB) {
		// get span from db instance's context
		spanInterface, _ := db.Get(spanKey)
		span := spanInterface.(go2sky.Span)

		defer span.End()

		sql := db.Statement.SQL.String()
		vars := db.Statement.Vars
		err := db.Statement.Error

		span.SetComponent(s.getComponent())
		span.SetSpanLayer(agentv3.SpanLayer_Database)
		span.Tag(go2sky.TagDBType, string(s.sqlType))
		span.Tag(go2sky.TagDBStatement, sql)

		if len(vars) != 0 {
			varsStr := fmt.Sprintf("%v", vars)
			span.Tag(go2sky.TagDBBindVariables, varsStr)
		}

		if err != nil {
			span.Error(time.Now(), err.Error())
		}
	}
}

func (s *SkyWalking) getComponent() int32 {
	switch s.sqlType {
	case MYSQL:
		return componentIDMySQL
	default:
		return componentIDUnknown
	}
}
