package user

import (
	"github.com/MG-RAST/Shock/shock-server/conf"
	"github.com/MG-RAST/Shock/shock-server/db"
	"github.com/MG-RAST/golib/go-uuid/uuid"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Array of User
type Users []User

// User struct
type User struct {
	Uuid     string `bson:"uuid" json:"uuid"`
	Username string `bson:"username" json:"username"`
	Fullname string `bson:"fullname" json:"fullname"`
	Email    string `bson:"email" json:"email"`
	Password string `bson:"password" json:"-"`
	Admin    bool   `bson:"shock_admin" json:"shock_admin"`
}

// Initialize creates a copy of the mongodb connection and then uses that connection to
// create the Users collection in mongodb. Then, it ensures that there is a unique index
// on the uuid key and the username key in this collection, creating the indexes if necessary.
func Initialize() (err error) {
	session := db.Connection.Session.Copy()
	defer session.Close()
	c := session.DB(conf.MONGODB_DATABASE).C("Users")
	if err = c.EnsureIndex(mgo.Index{Key: []string{"uuid"}, Unique: true}); err != nil {
		return err
	}
	if err = c.EnsureIndex(mgo.Index{Key: []string{"username"}, Unique: true}); err != nil {
		return err
	}

	// Setting admin users based on config file.  First, set all users to Admin = false
	if _, err = c.UpdateAll(bson.M{}, bson.M{"$set": bson.M{"shock_admin": false}}); err != nil {
		return err
	}

	// process list of amin users from config, create those that are missing
	for _, v := range conf.AdminUsers {
		info, err := c.UpdateAll(bson.M{"username": v}, bson.M{"$set": bson.M{"shock_admin": true}})
		if err != nil {
			return err
		}
		if info.Updated == 0 {
			if _, err := New(v, "", true); err != nil {
				return err
			}
		}
	}
	return
}

func New(username string, password string, isAdmin bool) (u *User, err error) {
	u = &User{Uuid: uuid.New(), Username: username, Password: password, Admin: isAdmin}
	err = u.Save()
	if err != nil {
		u = nil
	}
	return
}

func FindByUuid(uuid string) (u *User, err error) {
	session := db.Connection.Session.Copy()
	defer session.Close()
	c := session.DB(conf.MONGODB_DATABASE).C("Users")
	u = &User{Uuid: uuid}
	if err = c.Find(bson.M{"uuid": u.Uuid}).One(&u); err != nil {
		return nil, err
	}
	return
}

func FindByUsernamePassword(username string, password string) (u *User, err error) {
	session := db.Connection.Session.Copy()
	defer session.Close()
	c := session.DB(conf.MONGODB_DATABASE).C("Users")
	u = &User{}
	if err = c.Find(bson.M{"username": username, "password": password}).One(&u); err != nil {
		return nil, err
	}
	return
}

func AdminGet(u *Users) (err error) {
	session := db.Connection.Session.Copy()
	defer session.Close()
	c := session.DB(conf.MONGODB_DATABASE).C("Users")
	err = c.Find(nil).All(u)
	return
}

func (u *User) SetMongoInfo() (err error) {
	if uu, admin, err := dbGetInfo(u.Username); err == nil {
		u.Uuid = uu
		u.Admin = admin
		return nil
	} else {
		// this is a new user
		u.Uuid = uuid.New()
		// check if user is on admin list, if so set as true
		for _, v := range conf.AdminUsers {
			if v == u.Username {
				u.Admin = true
				break
			}
		}
		if err := u.Save(); err != nil {
			return err
		}
	}
	return
}

func dbGetInfo(username string) (uuid string, admin bool, err error) {
	session := db.Connection.Session.Copy()
	defer session.Close()
	c := session.DB(conf.MONGODB_DATABASE).C("Users")
	u := User{}
	if err = c.Find(bson.M{"username": username}).One(&u); err != nil {
		return "", false, err
	}
	return u.Uuid, u.Admin, nil
}

func (u *User) Save() (err error) {
	session := db.Connection.Session.Copy()
	defer session.Close()
	c := session.DB(conf.MONGODB_DATABASE).C("Users")
	_, err = c.Upsert(bson.M{"uuid": u.Uuid}, &u)
	return
}
