package lib

import (
	"gopkg.in/mgo.v2"
	"goSkylar/server/conf"
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
	mongodbURL := conf.MONGO_URI
	s, err := mgo.Dial(mongodbURL)
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
