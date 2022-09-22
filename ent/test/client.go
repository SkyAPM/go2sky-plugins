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
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky-plugins/ent"
	"github.com/SkyAPM/go2sky-plugins/ent/gen/entschema"
	entuser "github.com/SkyAPM/go2sky-plugins/ent/gen/entschema/user"
	httpplugin "github.com/SkyAPM/go2sky/plugins/http"
	"github.com/SkyAPM/go2sky/reporter"
	_ "github.com/go-sql-driver/mysql"
)

type testFunc func(context.Context, *entschema.Client) error

const (
	oap     = "mockoap:19876"
	service = "ent-client"
	dsn     = "user:password@tcp(mysql:3306)/database"
	addr    = ":8080"
)

func main() {
	// init tracer
	re, err := reporter.NewGRPCReporter(oap)
	if err != nil {
		log.Fatalf("create grpc reporter error: %v \n", err)
	}

	tracer, err := go2sky.NewTracer(service, go2sky.WithReporter(re))
	if err != nil {
		log.Fatalf("crate tracer error: %v \n", err)
	}

	client := ent.NewEntClient(dsn, tracer)

	if err != nil {
		log.Fatalf("open db error: %v \n", err)
	}

	route := http.NewServeMux()
	route.HandleFunc("/execute", func(res http.ResponseWriter, req *http.Request) {
		tests := []struct {
			name string
			fn   testFunc
		}{
			{"raw", testRaw},
			{"create", testCreate},
			{"query", testQuery},
			{"update", testUpdate},
			{"delete", testDelete},
			{"tx", testTx},
		}

		for _, test := range tests {
			log.Printf("excute test case %s", test.name)
			if err1 := test.fn(req.Context(), client); err1 != nil {
				log.Fatalf("test case %s failed: %v", test.name, err1)
			}
		}
		_, _ = res.Write([]byte("execute sql success"))
	})

	sm, err := httpplugin.NewServerMiddleware(tracer)
	if err != nil {
		log.Fatalf("create client error %v \n", err)
	}

	log.Println("start client")
	err = http.ListenAndServe(addr, sm(route))
	if err != nil {
		log.Fatalf("client start error: %v \n", err)
	}
}

func testRaw(ctx context.Context, client *entschema.Client) error {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	newClient := entschema.NewClient(entschema.Driver(entsql.OpenDB("mysql", db)))
	if err = newClient.Schema.Create(ctx); err != nil {
		return err
	}
	return nil
}

func testQuery(ctx context.Context, client *entschema.Client) error {
	user, err := client.User.Query().Where(entuser.ID(1)).First(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("%+v \n", user)

	return nil
}

func testCreate(ctx context.Context, client *entschema.Client) error {
	if err := client.User.Create().
		SetName("test02").
		SetAge(12).
		Exec(ctx); err != nil {
		return err
	}
	return nil
}

func testDelete(ctx context.Context, client *entschema.Client) error {
	if _, err := client.User.Delete().
		Where(entuser.Name("test02")).
		Exec(ctx); err != nil {
		return err
	}
	return nil
}

func testUpdate(ctx context.Context, client *entschema.Client) error {
	if err := client.User.Update().
		SetName("test01").
		Where(entuser.ID(1)).
		Exec(ctx); err != nil {
		return err
	}
	return nil
}

func testTx(ctx context.Context, client *entschema.Client) error {
	if err := ent.WithTx(ctx, client, func(tx *entschema.Tx) error {
		if err := tx.User.Create().
			SetName("test02").
			SetAge(12).
			Exec(ctx); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}
