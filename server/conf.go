package server

import (
	"github.com/jinzhu/gorm"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

type Env struct {
	DBDialect, DBOptions string
	Production           bool
}

type ORM struct {
	gorm.DB
}

func (env *Env) InitDB(dialect string, options string) error {
	db, err := gorm.Open(dialect, options)
	if err != nil {
		return err
	}
	db.AutoMigrate(&Quiz{}, &Question{}, &QuestionChoice{}, &QuestionAnswer{}, &User{}, &UserAnswer{})
	if err = db.Close(); err != nil {
		return err
	}
	env.DBDialect = dialect
	env.DBOptions = options
	return nil
}

func (env Env) ApiDefaults(handler httprouter.Handle) httprouter.Handle {
	return func(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
		res.Header().Set("Content-Type", "application/json;charset=utf-8")
		if !env.Production {
			res.Header().Set("Access-Control-Allow-Origin", "*")
		}
		handler(res, req, params)
	}
}

func (env Env) OpenDB() (orm *ORM, err error) {
	orm = new(ORM)
	orm.DB, err = gorm.Open(env.DBDialect, env.DBOptions)
	if !env.Production {
		orm.DB.LogMode(true)
	}
	return orm, err
}
