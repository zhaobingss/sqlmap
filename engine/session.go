package engine

import "database/sql"

const commitInitStatus = 1

/// session工厂用来创建session
type sessionFactory struct {
	db *sql.DB // 数据库连接
}

/// 创建新的session工厂
func newSessionFactory(db *sql.DB) *sessionFactory {
	s := &sessionFactory{}
	s.db = db
	return s
}

/// 使用session工厂创建session
func (s *sessionFactory) newSession(tplType string) *Session {
	ss := &Session{}
	ss.db = s.db
	ss.tplType = tplType
	return ss
}

/// session回话，管理事务
type Session struct {
	db          *sql.DB // 数据库连接
	tx          *sql.Tx // 事务
	commit      int8    // 本session提交的次数
	canRollback bool    // 标记是否可以回滚
	tplType     string  // 模板leixing
}

/// 开始事务
func (s *Session) BeginTx() error {
	s.canRollback = true
	if s.tx != nil {
		s.commit++
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	s.tx = tx
	s.commit = commitInitStatus
	return nil
}

/// 回滚事务
func (s *Session) Rollback() error {
	if s.tx == nil || !s.canRollback {
		return nil
	}

	err := s.tx.Rollback()
	if err != nil {
		return err
	}

	s.tx = nil
	return nil
}

/// 提交事务
func (s *Session) Commit() error {
	s.canRollback = false
	if s.tx == nil {
		return nil
	}

	if s.commit != commitInitStatus {
		s.commit--
		return nil
	}

	err := s.tx.Commit()
	if err != nil {
		return err
	}

	s.tx = nil
	return nil
}

/// 非查询语句
func (s *Session) Exec(key string, data interface{}) (sql.Result, error) {
	if s.tx == nil {
		return exec(key, s.tplType, data, s.db.Exec)
	} else {
		return exec(key, s.tplType, data, s.tx.Exec)
	}
}

/// 查询语句
func (s *Session) Query(key string, data interface{}) ([]map[string]string, error) {
	if s.tx == nil {
		return query(key, s.tplType, data, s.db.Prepare)
	} else {
		return query(key, s.tplType, data, s.tx.Prepare)
	}
}
