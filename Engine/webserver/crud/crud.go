package crud

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ICrud interface {
	getCollection() *mongo.Collection
	Create() (ICrud, error)
	Read(key string) error
	Update(newValue ICrud) error
	Delete() error
	Find(key string) ICrud
}

type DatabaseConnection struct {
	ctx context.Context
	DB  *mongo.Database
}

var (
	DbConnection *DatabaseConnection
)

func (d *DatabaseConnection) Collection(name string) (result *mongo.Collection) {
	result = d.DB.Collection(name)
	return result
}

func init() {
	DbConnection = &DatabaseConnection{}

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017/")
	DbConnection.ctx = context.TODO()
	client, err := mongo.Connect(DbConnection.ctx, clientOptions)

	if err != nil {
		panic(err)
	}

	DbConnection.DB = client.Database("HSNBlockchain")

}
