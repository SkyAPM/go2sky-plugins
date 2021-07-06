package sql

import (
	"errors"
	"regexp"
)

const (
	TagDbType          = "db.type"
	TagDbInstance      = "db.instance"
	TagDbStatement     = "db.statement"
	TagDbSqlParameters = "db.sql.parameters"
)

var ErrUnsupportedOp = errors.New("operation unsupported by the underlying driver")

// emptyInjectFunc defines a empty injector for propagation.Injector function
func emptyInjectFunc(key, value string) error { return nil }

// parseAddr parse dsn to a endpoint addr string (host:port)
func parseAddr(dsn string, dbType DBType) (addr string) {
	switch dbType {
	case MYSQL:
		// [user[:password]@][net[(addr)]]/dbname[?param1=value1&paramN=valueN]
		re := regexp.MustCompile(`\(.+\)`)
		addr = re.FindString(dsn)
		addr = addr[1 : len(addr)-1]
	case IPV4:
		// ipv4 addr
		re := regexp.MustCompile(`((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})(\.((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})){3}:\d{1,5}`)
		addr = re.FindString(dsn)
	}
	return
}

func genOpName(dbType DBType, opName string) string {
	switch dbType {
	case MYSQL:
		return "Mysql/Go2Sky/" + opName
	default:
		return "Sql/Go2Sky/" + opName
	}
}
