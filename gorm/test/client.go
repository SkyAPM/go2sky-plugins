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

package main

import (
	"fmt"
	"log"
	"net/http"

	gormPlugin "github.com/SkyAPM/go2sky-plugins/gorm"
	httpPlugin "github.com/SkyAPM/go2sky/plugins/http"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/reporter"
)

type testFunc func(*gorm.DB) error

const (
	oap     = "mockoap:19876"
	service = "gorm-client"
	dsn     = "user:password@tcp(mysql:3306)/database"
	addr    = ":8080"
	peer    = "mysql:3306"
)

// DB model
type User struct {
	ID   uint
	Name string
	Age  uint8
}

func main() {
	// init tracer
	re, err := reporter.NewGRPCReporter(oap)
	//re, err := reporter.NewLogReporter()
	if err != nil {
		log.Fatalf("create grpc reporter error: %v \n", err)
	}

	tracer, err := go2sky.NewTracer(service, go2sky.WithReporter(re))
	if err != nil {
		log.Fatalf("crate tracer error: %v \n", err)
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("open db error: %v \n", err)
	}

	db.Use(gormPlugin.New(tracer,
		gormPlugin.WithSqlDBType(gormPlugin.MYSQL),
		gormPlugin.WithQueryReport(),
		gormPlugin.WithParamReport(),
		gormPlugin.WithPeerAddr(peer),
	))

	route := http.NewServeMux()
	route.HandleFunc("/execute", func(res http.ResponseWriter, req *http.Request) {
		tests := []struct {
			name string
			fn   testFunc
		}{
			{"raw", TestRaw},
			{"create", TestCreate},
			{"query", TestQuery},
			{"row", TestRow},
			{"update", TestUpdate},
			{"delete", TestDelete},
		}

		dbWithCtx := db.WithContext(req.Context())
		for _, test := range tests {
			log.Printf("excute test case %s", test.name)
			if err := test.fn(dbWithCtx); err != nil {
				log.Fatalf("test case %s failed: %v", test.name, err)
			}
		}
		_, _ = res.Write([]byte("execute sql success"))
	})

	sm, err := httpPlugin.NewServerMiddleware(tracer)
	if err != nil {
		log.Fatalf("create client error %v \n", err)
	}

	log.Println("start client")
	err = http.ListenAndServe(addr, sm(route))
	if err != nil {
		log.Fatalf("client start error: %v \n", err)
	}
}

func TestRaw(db *gorm.DB) error {
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS users (id char(255), name VARCHAR(255), age INTEGER)`).Error; err != nil {
		return fmt.Errorf("Create error: %s", err.Error())
	}

	return nil
}

func TestCreate(db *gorm.DB) error {
	user := User{Name: "Jinzhu", Age: 18}
	if err := db.Create(&user).Error; err != nil {
		return fmt.Errorf("Create error: %w", err)
	}

	return nil
}

func TestQuery(db *gorm.DB) error {
	var user User
	if err := db.First(&user).Error; err != nil {
		return fmt.Errorf("Query error: %w", err)
	}

	return nil
}

func TestRow(db *gorm.DB) error {
	var name string
	var age uint8
	row := db.Table("users").Where("name = ?", "jinzhu").Select("name", "age").Row()
	row.Scan(&name, &age)

	return nil
}

func TestUpdate(db *gorm.DB) error {
	tx := db.Model(&User{}).Where("name = ?", "jinzhu").Update("name", "hello")
	if err := tx.Error; err != nil {
		return fmt.Errorf("Update error: %w", err)
	}

	return nil
}

func TestDelete(db *gorm.DB) error {
	if err := db.Delete(&User{}, 1).Error; err != nil {
		return fmt.Errorf("Delete error: %w", err)
	}

	return nil
}
