package dbs

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"sync"
	"time"
)

const (
	mysqlEnvDev = "dev"
)

var (
	_mysqlDB     *MysqlDB
	_mysqlDbOnce sync.Once
)

type MySqlLogger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

type IMySqlDB interface {
	MainDB() *gorm.DB
	OtherDB(name string) *gorm.DB
	ClearAllData(db *gorm.DB, tables ...string)
}

type MySqlConfig struct {
	Env         string            `json:"env,optional"`
	OtherDns    map[string]string `json:"otherDns,optional" yaml:"otherDns"`
	MainDns     string            `json:"mainDns" yaml:"mainDns"`
	MaxIdleConn int               `json:"maxIdleConn,optional" yaml:"maxIdleConn"`
	MaxOpenConn int               `json:"maxOpenConn,optional" yaml:"maxOpenConn"`
	LogLevel    string            `json:"logLevel,optional" yaml:"logLevel"`
	SlowLogTm   int               `json:"slowLogTm,optional" yaml:"slowLogTm"`
}

func (sel *MySqlConfig) GetEnv() string {
	if sel.Env == "" {
		sel.Env = mysqlEnvDev
	}
	return sel.Env
}

func (sel *MySqlConfig) GetMaxIdleConn() int {
	return getIntWithDefault(sel.MaxIdleConn, 5)
}

func (sel *MySqlConfig) GetMaxOpenConn() int {
	return getIntWithDefault(sel.MaxOpenConn, 20)
}

func (sel *MySqlConfig) GetSlowLogTm() time.Duration {
	if sel.SlowLogTm == 0 {
		sel.SlowLogTm = 2
	}
	return time.Duration(sel.SlowLogTm) * time.Second
}

func (sel *MySqlConfig) GetLogLevel() logger.LogLevel {
	switch sel.LogLevel {
	case "info":
		return logger.Info
	case "warn":
		return logger.Warn
	case "error":
		return logger.Error
	default:
		return logger.Silent
	}
}

func getIntWithDefault(n int, d int) int {
	if n == 0 {
		return d
	}
	return n
}

func logLevel(level string) logger.LogLevel {
	switch level {
	case "info":
		return logger.Info
	case "warn":
		return logger.Warn
	case "error":
		return logger.Error
	default:
		return logger.Silent
	}
}

type MysqlDB struct {
	env      string
	mainDb   *gorm.DB
	otherDbs map[string]*gorm.DB
}

func InitMysqlDB(cfg MySqlConfig, l ...MySqlLogger) IMySqlDB {
	_mysqlDbOnce.Do(func() {
		var newLogger logger.Interface
		if len(l) > 0 {
			newLogger = newDefaultLogger(&cfg, l[0])
		} else {
			newLogger = logger.New(
				log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
				logger.Config{
					SlowThreshold:             cfg.GetSlowLogTm(), // Slow SQL threshold
					LogLevel:                  cfg.GetLogLevel(),  // Log level
					IgnoreRecordNotFoundError: true,               // Ignore ErrRecordNotFound error for logger
					ParameterizedQueries:      false,              // Don't include params in the SQL log
					Colorful:                  false,              // Disable color
				},
			)
		}
		_mysqlDB = new(MysqlDB)
		_mysqlDB.env = cfg.Env
		_mysqlDB.otherDbs = make(map[string]*gorm.DB)
		_mysqlDB.mainDb = openGormDB(cfg.MainDns, cfg.GetMaxIdleConn(), cfg.GetMaxOpenConn(), newLogger)
		if len(cfg.OtherDns) > 0 {
			for k, v := range cfg.OtherDns {
				_mysqlDB.otherDbs[k] = openGormDB(v, cfg.MaxIdleConn, cfg.MaxOpenConn, newLogger)
			}
		}

	})

	return _mysqlDB
}

func (sel *MysqlDB) MainDB() *gorm.DB {
	return sel.mainDb
}

func (sel *MysqlDB) OtherDB(name string) *gorm.DB {
	if len(sel.otherDbs) == 0 {
		return sel.mainDb
	}
	return sel.otherDbs[name]
}

func (sel *MysqlDB) ClearAllData(db *gorm.DB, tables ...string) {
	if sel.env == mysqlEnvDev {
		if len(tables) <= 0 {
			tables, _ = db.Migrator().GetTables()
		}
		for _, table := range tables {
			_ = db.Table(table).Where("1 = 1").Delete(nil).Error
		}
	}
}

func openGormDB(dns string, maxIdleConns int, maxOpenConns int, l logger.Interface) *gorm.DB {
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                           dns,
		DefaultStringSize:             255,
		DefaultDatetimePrecision:      nil,
		DisableDatetimePrecision:      true,
		DontSupportNullAsDefaultValue: true,
		DontSupportRenameColumn:       true,
		DontSupportRenameColumnUnique: true,
	}), &gorm.Config{
		Logger: l,
	})
	if err != nil {
		panic(err)
	}
	sqlDb, err := db.DB()
	if err != nil {
		panic(err)
	}
	if err = sqlDb.Ping(); err != nil {
		panic(err)
	}
	sqlDb.SetMaxIdleConns(maxIdleConns)
	sqlDb.SetMaxOpenConns(maxOpenConns)
	sqlDb.SetConnMaxLifetime(time.Hour)
	return db
}

type defaultLogger struct {
	level         logger.LogLevel
	log           MySqlLogger
	SlowThreshold time.Duration
	traceStr      string
	traceWarnStr  string
	traceErrStr   string
}

func newDefaultLogger(cfg *MySqlConfig, l MySqlLogger) logger.Interface {
	return &defaultLogger{
		level:         logLevel(cfg.LogLevel),
		log:           l,
		SlowThreshold: time.Second * 2,
		traceStr:      "[%.3fms rows:%v] %s",
		traceWarnStr:  "%s [%.3fms rows:%v] %s",
		traceErrStr:   "%s [%.3fms rows:%v] %s",
	}
}

func (sel *defaultLogger) LogMode(l logger.LogLevel) logger.Interface {
	sel.level = l
	return sel
}

func (sel *defaultLogger) Info(ctx context.Context, msg string, arg ...interface{}) {
	if logger.Info <= sel.level {
		sel.log.Infof(msg, arg...)
	}
}

func (sel *defaultLogger) Warn(ctx context.Context, msg string, arg ...interface{}) {
	if logger.Warn <= sel.level {
		sel.log.Warnf(msg, arg...)
	}
}

func (sel *defaultLogger) Error(ctx context.Context, msg string, arg ...interface{}) {
	if logger.Error <= sel.level {
		sel.log.Errorf(msg, arg...)
	}
}

func (sel *defaultLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	switch {
	case err != nil && sel.level >= logger.Error && !errors.Is(err, logger.ErrRecordNotFound):
		if rows == -1 {
			sel.log.Errorf(sel.traceErrStr, err.Error(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
		}
		sel.log.Errorf(sel.traceErrStr, err.Error(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
	case elapsed > sel.SlowThreshold && sel.SlowThreshold != 0 && sel.level >= logger.Warn:
		slowLog := fmt.Sprintf("SLOW SQL >= %v", sel.SlowThreshold)
		if rows == -1 {
			sel.log.Warnf(sel.traceWarnStr, slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		}
		sel.log.Warnf(sel.traceWarnStr, slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
	case sel.level >= logger.Info:
		if rows == -1 {
			sel.log.Infof(sel.traceStr, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		}
		sel.log.Infof(sel.traceStr, float64(elapsed.Nanoseconds())/1e6, rows, sql)
	}
}
