package lib

import (
	"fmt"
	"gopkg.in/mgo.v2"
)

type MongoDriver struct {
	DbSource   string
	DbName     string
	TableName  string
	mgoSession *mgo.Session
}

func (this *MongoDriver) Init() error {
	if this.DbSource == "" {
		this.DbSource = "default"
	}
	cfg := NewConfigUtil("config.ini")
	mongodb_url, err := cfg.GetString(fmt.Sprintf("mongodb_%s", this.DbSource), "url")
	if err != nil {
		return err
	}
	s, err := mgo.Dial(mongodb_url)
	if err != nil {
		return err
	}
	s.SetMode(mgo.Monotonic, true)
	this.mgoSession = s
	return nil
}

func (this MongoDriver) NewTable() (*mgo.Collection, error) {
	if this.mgoSession == nil {
		this.Init()
	}
	this.mgoSession.Refresh()
	return this.mgoSession.DB(this.DbName).C(this.TableName), nil
}
