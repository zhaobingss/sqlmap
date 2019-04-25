package engine

import "database/sql"

const commitInitStatus = 1

/// session工厂用来创建session
type SessionFactory struct {
	db *sql.DB // 数据库连接
}

/// 创建新的session工厂
func NewSessionFactory(db *sql.DB) *SessionFactory {
	s := &SessionFactory{}
	s.db = db
	return s
}

/// 使用session工厂创建session
func (s *SessionFactory) NewSession() *Session {
	ss := &Session{}
	ss.db = s.db
	return ss
}

/// session回话，管理事务
type Session struct {
	db          *sql.DB // 数据库连接
	tx          *sql.Tx // 事务
	commit      int8    // 本session提交的次数
	canRollback bool    // 标记是否可以回滚
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

/// 执行sql语句
func (s *Session) Exec(sql string, args ...int) (sql.Result, error) {

	return nil, nil
}
