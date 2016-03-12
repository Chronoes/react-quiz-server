package server

import (
	"github.com/jinzhu/gorm"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

// Env defines required environment variables for operations performed in the API
type Env struct {
	DBDialect, DBOptions string
	Production           bool
}

// InitDB synchronizes database tables and sets database options for further use
func (env *Env) InitDB(dialect string, options string) error {
	db, err := gorm.Open(dialect, options)
	if err != nil {
		return err
	}
	db.AutoMigrate(&Quiz{}, &Question{}, &QuestionChoice{}, &User{}, &UserTextAnswer{}, &UserChoiceAnswer{})
	if err = db.Close(); err != nil {
		return err
	}
	env.DBDialect = dialect
	env.DBOptions = options
	return nil
}

// APIDefaults is a required wrapper for httprouter.Handle functions that define defaults used for the API
func (env Env) APIDefaults(handler httprouter.Handle) httprouter.Handle {
	return func(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
		res.Header().Set("Accept", "application/json")
		res.Header().Set("Content-Type", "application/json;charset=utf-8")
		if !env.Production {
			res.Header().Set("Access-Control-Allow-Origin", "*")
		}
		handler(res, req, params)
	}
}

func (env Env) openDB() (db gorm.DB, err error) {
	db, err = gorm.Open(env.DBDialect, env.DBOptions)
	if !env.Production {
		db.LogMode(true)
	}
	return db, err
}
