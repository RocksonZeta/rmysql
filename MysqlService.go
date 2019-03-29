package rmysql

import (
	"database/sql"
	"reflect"
	"strconv"

	"github.com/go-gorp/gorp"
	_ "github.com/go-sql-driver/mysql"
	"github.com/guregu/null"
)

type MysqlService struct {
	*MysqlWrapper
	DbMap           *gorp.DbMap
	Transaction     *gorp.Transaction
	Tx              *MysqlWrapper
	TmpMysqlWrapper *MysqlWrapper
}

func (s *MysqlService) Close() {
	err := s.DbMap.Db.Close()
	s.Transaction = nil
	s.Tx = nil
	s.DbMap = nil
	s.MysqlWrapper = nil
	s.TmpMysqlWrapper = nil
	CheckError(err)
}

//Begin transaction , in transation we should use `mysql.Tx.Select`
func (s *MysqlService) Begin() {
	tx, err := s.DbMap.Begin()
	CheckError(err)
	s.Transaction = tx
	s.Tx = Wrap(s.Transaction)
	s.TmpMysqlWrapper = s.MysqlWrapper
	s.MysqlWrapper = s.Tx
}
func (s *MysqlService) Commit() error {
	s.TmpMysqlWrapper = s.Tx
	s.MysqlWrapper = s.MysqlWrapper
	return s.Transaction.Commit()
}
func (s *MysqlService) Rollback() error {
	s.TmpMysqlWrapper = s.Tx
	s.MysqlWrapper = s.MysqlWrapper
	return s.Transaction.Rollback()
}
func (s *MysqlService) AddTable(i interface{}) {
	s.DbMap.AddTable(i).SetKeys(true, "Id")
}
func (s *MysqlService) Add(i interface{}) {
	s.AddTable(reflect.ValueOf(i).Elem().Interface())
	s.Insert(i)
}

type MysqlWrapper struct {
	Executor gorp.SqlExecutor
}

func Wrap(executor gorp.SqlExecutor) *MysqlWrapper {
	return &MysqlWrapper{Executor: executor}
}

func (m *MysqlWrapper) Select(i interface{}, query string, args ...interface{}) {
	_, err := m.Executor.Select(i, query, args...)
	CheckError(err)
}
func (m *MysqlWrapper) SelectOne(holder interface{}, query string, args ...interface{}) {
	err := m.Executor.SelectOne(holder, query, args...)
	if err != nil {
		if err.Error() != "sql: no rows in result set" {
			CheckError(err)
		}
	}
}
func (m *MysqlWrapper) SelectFloat(query string, args ...interface{}) float64 {
	r, err := m.Executor.SelectFloat(query, args...)
	CheckError(err)
	return r
}
func (m *MysqlWrapper) SelectNullFloat(query string, args ...interface{}) null.Float {
	r, err := m.Executor.SelectNullFloat(query, args...)
	CheckError(err)
	return null.Float{r}
}
func (m *MysqlWrapper) SelectInt64(query string, args ...interface{}) int64 {
	r, err := m.Executor.SelectInt(query, args...)
	CheckError(err)
	return r
}
func (m *MysqlWrapper) SelectInt(query string, args ...interface{}) int {
	r, err := m.Executor.SelectInt(query, args...)
	CheckError(err)
	return int(r)
}
func (m *MysqlWrapper) SelectNullInt(query string, args ...interface{}) null.Int {
	r, err := m.Executor.SelectNullInt(query, args...)
	CheckError(err)
	return null.Int{r}
}
func (m *MysqlWrapper) SelectStr(query string, args ...interface{}) string {
	r, err := m.Executor.SelectStr(query, args...)
	CheckError(err)
	return r
}
func (m *MysqlWrapper) SelectNullStr(query string, args ...interface{}) null.String {
	r, err := m.Executor.SelectNullStr(query, args...)
	CheckError(err)
	return null.String{r}
}
func (m *MysqlWrapper) Exec(query string, args ...interface{}) sql.Result {
	r, err := m.Executor.Exec(query, args...)
	CheckError(err)
	return r
}
func (m *MysqlWrapper) Insert(list ...interface{}) {
	err := m.Executor.Insert(list...)
	CheckError(err)
}
func (m *MysqlWrapper) Update(list ...interface{}) int64 {
	r, err := m.Executor.Update(list...)
	CheckError(err)
	return r
}
func (m *MysqlWrapper) GetByKey(i interface{}, kvs ...interface{}) {
	if len(kvs) <= 0 {
		return
	}
	structName := reflect.TypeOf(i).Elem().Name()
	ssql := "select * from `" + structName + "` where "
	args := make([]interface{}, len(kvs)/2)
	for i := 0; i < len(kvs); i += 2 {
		ssql += kvs[i].(string) + "=? and "
		args[i/2] = kvs[i+1]
	}
	ssql = ssql[0:len(ssql)-5] + " limit 1"
	m.SelectOne(i, ssql, args...)
}
func (m *MysqlWrapper) Get(i interface{}, id interface{}) {
	m.SelectOne(i, "select * from `"+reflect.TypeOf(i).Elem().Name()+"` where id=? limit 1", id)
}
func (m *MysqlWrapper) GetObject(i interface{}, id interface{}, table string, fields []string) {
	m.SelectOne(i, "select "+JoinFields("", fields...)+" from `"+table+"` where id=? limit 1", id)
}
func (m *MysqlWrapper) Gets(i interface{}, ids ...interface{}) {
	if len(ids) == 0 {
		return
	}
	structName := reflect.TypeOf(i).Elem().Elem().Name()
	ssql := "select * from `" + structName + "` where id in ("
	for range ids {
		ssql += "?,"
	}
	ssql = ssql[0 : len(ssql)-1]
	ssql += ")"
	m.Select(i, ssql, ids...)
}
func (m *MysqlWrapper) List(i interface{}, ids []int) {
	if len(ids) == 0 {
		return
	}
	idMap := make(map[int]bool, len(ids))
	for _, id := range ids {
		idMap[id] = true
	}
	structName := reflect.TypeOf(i).Elem().Elem().Name()
	ssql := "select * from `" + structName + "` where id in ("
	for id := range idMap {
		ssql += strconv.Itoa(id) + ","
	}
	ssql = ssql[0 : len(ssql)-1]
	ssql += ")"
	m.Select(i, ssql)
}

func (m *MysqlWrapper) ListObject(r interface{}, ids []int, table string, fields []string) {
	if len(ids) == 0 {
		return
	}
	idMap := make(map[int]bool, len(ids))
	for _, id := range ids {
		idMap[id] = true
	}
	fieldSql := "*"
	if len(fields) > 0 {
		fieldSql = JoinFields(",", fields...)
	}
	ssql := "select " + fieldSql + " from `" + table + "` where id in ("
	for id := range idMap {
		ssql += strconv.Itoa(id) + ","
	}
	ssql = ssql[0 : len(ssql)-1]
	ssql += ")"
	m.Select(r, ssql)
}

func (m *MysqlWrapper) CountByKeys(table, appendSql string, kvs ...interface{}) int64 {
	// structName := reflect.TypeOf(r).Elem().Elem().Name()
	ssql := "select count(*) from `" + table + "` where "
	args := make([]interface{}, len(kvs)/2)
	for i := 0; i < len(kvs); i += 2 {
		ssql += kvs[i].(string) + "=? and "
		args[i/2] = kvs[i+1]
	}
	ssql = ssql[0 : len(ssql)-5]
	if appendSql != "" {
		ssql += " " + appendSql
	}
	return m.SelectInt64(ssql, args...)
}

//GetByKeys
// eg. var r []UserMajor
// GetByKeys(&r ,"order by id", "uid",1)
func (m *MysqlWrapper) ListByKeys(r interface{}, appendSql string, kvs ...interface{}) {
	if len(kvs) <= 0 {
		return
	}
	structName := reflect.TypeOf(r).Elem().Elem().Name()
	ssql := "select * from `" + structName + "` where "
	args := make([]interface{}, len(kvs)/2)
	for i := 0; i < len(kvs); i += 2 {
		ssql += kvs[i].(string) + "=? and "
		args[i/2] = kvs[i+1]
	}
	ssql = ssql[0 : len(ssql)-5]
	if appendSql != "" {
		ssql += " " + appendSql
	}
	m.Select(r, ssql, args...)
}
func (m *MysqlWrapper) DeleteList(list ...interface{}) interface{} {
	r, err := m.Executor.Delete(list...)
	CheckError(err)
	return r
}
func (m *MysqlWrapper) Delete(table string, id interface{}) sql.Result {
	return m.Exec("delete from `"+table+"` where id=?", id)
}
