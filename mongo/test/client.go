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
	"log"
	"net/http"

	"github.com/SkyAPM/go2sky"
	httpPlugin "github.com/SkyAPM/go2sky/plugins/http"
	"github.com/SkyAPM/go2sky/reporter"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	mongoPlugin "github.com/SkyAPM/go2sky-plugins/mongo"
)

type testFunc func(ctx context.Context, client *mongo.Client) error

const (
	oap     = "mockoap:19876"
	service = "mongo-client"
	dsn     = "mongodb://user:password@mongo:27017"
	addr    = ":8080"
	db      = "database"
	peer    = "mongo:27017"
)

// User model.
type User struct {
	ID   primitive.ObjectID `bson:"_id,omitempty"`
	Name string             `bson:"name"`
	Age  int                `bson:"age"`
}

func main() {
	re, err := reporter.NewGRPCReporter(oap)
	if err != nil {
		log.Fatalf("create grpc reporter error: %v \n", err)
	}

	tracer, err := go2sky.NewTracer(service, go2sky.WithReporter(re))
	if err != nil {
		log.Fatalf("crate tracer error: %v \n", err)
	}

	ctx := context.Background()
	middleware := mongoPlugin.Middleware(tracer, peer)
	/* middleware := mongoPlugin.Middleware(tracer, func(span go2sky.Span, peer, evt *event.CommandStartedEvent) {
		span.Tag(go2sky.TagDBStatement, evt.Command.String())
	}) */
	// init connect mongodb.
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dsn).SetMonitor(middleware))
	if err != nil {
		log.Fatalf("connect mongodb error %v \n", err)
	}

	route := http.NewServeMux()
	route.HandleFunc("/execute", func(res http.ResponseWriter, req *http.Request) {
		tests := []struct {
			name string
			fn   testFunc
		}{
			{"create_collection", TestCreateCollection},
			{"create", TestCreate},
			{"query", TestQuery},
			{"update", TestUpdate},
			{"delete", TestDelete},
		}

		for _, test := range tests {
			log.Printf("excute test case %s", test.name)
			if err = test.fn(req.Context(), client); err != nil {
				log.Fatalf("test case %s failed: %v", test.name, err)
			}
		}
		_, _ = res.Write([]byte("execute success"))
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

// TestCreateCollection create collection.
func TestCreateCollection(ctx context.Context, client *mongo.Client) error {
	return client.Database(db).CreateCollection(ctx, "users")
}

// TestCreate create model.
func TestCreate(ctx context.Context, client *mongo.Client) error {
	collection := client.Database(db).Collection("users")
	objectID, err := primitive.ObjectIDFromHex("637334579a3d0cf34c31d08f")
	if err != nil {
		return err
	}
	_, err = collection.InsertOne(ctx, &User{
		ID:   objectID,
		Name: "Elza2",
		Age:  18,
	})
	return err
}

// TestQuery query model.
func TestQuery(ctx context.Context, client *mongo.Client) error {
	collection := client.Database(db).Collection("users")
	var user User
	err := collection.FindOne(ctx, bson.D{
		{Key: "name", Value: "Elza2"},
	}).Decode(&user)

	return err
}

// TestUpdate update model.
func TestUpdate(ctx context.Context, client *mongo.Client) error {
	collection := client.Database(db).Collection("users")

	var user User
	err := collection.FindOne(ctx, bson.D{
		{Key: "name", Value: "Elza2"},
	}).Decode(&user)
	if err != nil {
		return err
	}

	_, err = collection.UpdateByID(ctx, user.ID, primitive.D{{
		Key: "$set", Value: primitive.D{
			{Key: "age", Value: 22},
		},
	}})
	return err
}

// TestDelete delete model.
func TestDelete(ctx context.Context, client *mongo.Client) error {
	collection := client.Database(db).Collection("users")

	_, err := collection.DeleteOne(ctx, primitive.D{{Key: "name", Value: "Elza2"}})
	return err
}
