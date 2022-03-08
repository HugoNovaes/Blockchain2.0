package crud

import (
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type User struct {
	ID         primitive.ObjectID `bson:"_id,ommitempty"`
	Login      string             `bson:"login"`
	FullName   string             `bson:"full_name"`
	CreateTime string             `bson:"create_time"`
}

func (u *User) getCollection() (result *mongo.Collection) {
	result = DbConnection.DB.Collection("users")
	return result
}

func (u *User) Create() (result ICrud, err error) {
	col := u.getCollection()
	u.ID = primitive.NewObjectID()
	doc, err := col.InsertOne(DbConnection.ctx, u)

	log.Printf("%v\r\n", doc)
	result = u
	return result, err
}

func (u *User) Read(key string) (err error) {
	return err
}

func (u *User) Update(newValue ICrud) (err error) {
	return err

}

func (u *User) Delete() (err error) {
	return err
}

func (u *User) Find(email string) (result ICrud) {

	col := u.getCollection()
	mr := col.FindOne(DbConnection.ctx, bson.M{email: email})

	fmt.Println(mr)

	return result
}
