package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/SunMaungOo/task-runner-auth/internal/model"
	"github.com/SunMaungOo/task-runner-auth/internal/repo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type document struct {
	ID           bson.ObjectID `bson:"_id,omitempty"`
	Email        string        `bson:"email"`
	PasswordHash string        `bson:"passwordHash"`
	CreatedAt    time.Time     `bson:"createdAt"`
}

func (doc document) toModel() model.User {
	return model.User{
		ID:           doc.ID.Hex(),
		Email:        doc.Email,
		PasswordHash: doc.PasswordHash,
		CreatedAt:    doc.CreatedAt,
	}
}

type UserRespository struct {
	collection *mongo.Collection
}

func New(context context.Context, uri string, database string) (*UserRespository, error) {

	client, err := mongo.Connect(options.Client().ApplyURI(uri))

	if err != nil {
		return nil, err
	}

	if err := client.Ping(context, nil); err != nil {
		return nil, err
	}

	collection := client.Database(database).Collection("users")

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	if _, err := collection.Indexes().CreateOne(context, indexModel); err != nil {
		return nil, err
	}

	return &UserRespository{collection: collection}, nil

}

func (userRepo UserRespository) Create(context context.Context, email string, passwordHash string) (model.User, error) {

	doc := document{
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now().UTC(),
	}

	result, err := userRepo.collection.InsertOne(context, doc)

	if err != nil {

		if mongo.IsDuplicateKeyError(err) {
			return model.User{}, repo.ErrorDuplicateEmail
		}

		return model.User{}, err
	}

	oid, ok := result.InsertedID.(bson.ObjectID)

	if !ok {
		return model.User{}, fmt.Errorf("expected ObjectID, got %T", result.InsertedID)
	}

	doc.ID = oid

	return doc.toModel(), nil
}

func (userRepo UserRespository) GetByEmail(context context.Context, email string) (model.User, error) {

	var doc document

	err := userRepo.collection.FindOne(context, bson.D{{Key: "email", Value: email}}).Decode(&doc)

	if err != nil {

		if errors.Is(err, mongo.ErrNoDocuments) {
			return model.User{}, repo.ErrorNotFound
		}

		return model.User{}, err
	}

	return doc.toModel(), nil

}

func (repo UserRespository) Ping(context context.Context) error {
	return repo.collection.Database().Client().Ping(context, nil)
}
