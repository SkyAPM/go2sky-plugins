package sql

import (
	"database/sql/driver"

	"github.com/SkyAPM/go2sky"
)

// tx is a wrapper for driver.Tx
type tx struct {
	tx driver.Tx
	tracer *go2sky.Tracer
	span go2sky.Span
}

func (t *tx) Commit() error {
	if t.span != nil {
		defer t.span.End()
	}
	return t.tx.Commit()
}

func (t *tx) Rollback() error {
	if t.span != nil {
		defer t.span.End()
	}
	return t.tx.Rollback()
}