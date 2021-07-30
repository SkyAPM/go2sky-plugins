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

package sql

import (
	"database/sql/driver"

	"github.com/SkyAPM/go2sky"
)

// tx is a wrapper for driver.Tx
type tx struct {
	tx   driver.Tx
	span go2sky.Span
}

func (t *tx) Commit() error {
	if t.span != nil {
		t.span.Tag(tagDbStatement, "commit")
		defer t.span.End()
	}
	return t.tx.Commit()
}

func (t *tx) Rollback() error {
	if t.span != nil {
		t.span.Tag(tagDbStatement, "rollback")
		defer t.span.End()
	}
	return t.tx.Rollback()
}
