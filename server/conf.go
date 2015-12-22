package server

import "github.com/jinzhu/gorm"

type Env struct {
	DB *gorm.DB
}

func (env *Env) InitDB(dialect string, options string) error {
	db, err := gorm.Open(dialect, options)
	db.CreateTable(&Choice{})
	db.CreateTable(&Question{})
	db.CreateTable(&Quiz{})
	env.DB = &db
	return err
}
