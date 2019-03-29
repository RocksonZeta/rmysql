package rmysql

import (
	"database/sql"
	stdlog "log"
	"os"

	"github.com/go-gorp/gorp"
)

func NewMysqlService(url, prefix string, trace bool) *MysqlService {

	config := MysqlConfig{
		Url:       url,
		LogPrefix: prefix,
		Trace:     trace,
	}
	return NewMysqlServiceWithConfig(config)
}

type MysqlConfig struct {
	Url         string
	LogPrefix   string
	Trace       bool
	TraceWriter gorp.GorpLogger
	Engine      string
	Encoding    string
}

func NewMysqlServiceWithConfig(config MysqlConfig) *MysqlService {
	if config.Encoding == "" {
		config.Encoding = "UTF8MB4"
	}
	if config.Engine == "" {
		config.Engine = "InnoDB"
	}
	if config.TraceWriter == nil && config.Trace {
		config.TraceWriter = stdlog.New(os.Stdout, "", stdlog.Lmicroseconds)
	}
	db, err := sql.Open("mysql", config.Url)
	if err != nil {
		panic(err)
	}
	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{config.Engine, config.Encoding}}
	if config.Trace {
		dbmap.TraceOn(config.LogPrefix, config.TraceWriter)
	} else {
		dbmap.TraceOff()
	}
	return &MysqlService{MysqlWrapper: Wrap(dbmap), DbMap: dbmap}
}
