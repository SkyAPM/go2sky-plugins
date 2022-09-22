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

package ent

import (
	"context"

	"github.com/SkyAPM/go2sky-plugins/ent/gen/entschema"
	"github.com/pkg/errors"
)

// WithTx 事务快捷方式.
// https://entgo.io/docs/transactions/#best-practices
func WithTx(ctx context.Context, client *entschema.Client, fn func(tx *entschema.Tx) error) error {
	// not use. throw error panic: interface conversion: sql.ExecQuerier is *sql.DB, not *sql.DB
	// tx, err := client.Tx(ctx, nil)
	tx, err := client.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if v := recover(); v != nil {
			err = tx.Rollback()
			panic(v)
		}
	}()
	if err = fn(tx); err != nil {
		if r := tx.Rollback(); r != nil {
			err = errors.Wrapf(err, "rolling back transaction: %v", r)
		}
		return err
	}
	if err = tx.Commit(); err != nil {
		return errors.Wrapf(err, "committing transaction: %v", err)
	}
	return nil
}
