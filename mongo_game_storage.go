package main

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	. "go.mongodb.org/mongo-driver/bson"
	mgo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"time"
)

type mongoSessionStorage struct {
	mClient *mgo.Client
	mColl   *mgo.Collection
}

func NewMongoSessionStorage() *mongoSessionStorage {
	mUser, mPass, mHost := os.Getenv("MONGO_USER"), os.Getenv("MONGO_PASS"), os.Getenv("MONGO_HOST")
	mDBName := os.Getenv("MONGO_DBNAME")
	mConnUrl := fmt.Sprintf(
		"mongodb+srv://%s:%s@%s/%s?retryWrites=true&w=majority",
		mUser,
		mPass,
		mHost,
		mDBName,
	)

	mConnOpts := options.Client()
	mConnOpts.ApplyURI(mConnUrl)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mClient, err := mgo.Connect(ctx, mConnOpts)
	if err != nil {
		log.Fatal(err)
	}

	if err = mClient.Ping(ctx, nil); err != nil {
		log.Fatal("failed to connect to mongodb: ", err)
	}

	return &mongoSessionStorage{
		mClient: mClient,
		mColl:   ensureCollectionAndIndexes(mClient),
	}
}

func (i *mongoSessionStorage) StoreSession(session *GameInstance) error {
	gameId, err := uuid.Parse(session.GameId.Value)
	if err != nil {
		return err
	}

	_, err = i.mColl.UpdateOne(
		context.Background(),
		M{"game_id": M{"eq": gameId}}, //Update only game session with matching uuid
		session,
		options.Update().SetUpsert(true),
	)
	if err != nil {
		log.Println("failed to store session in mongo: ", err)
		return err
	}

	return nil
}

func (i *mongoSessionStorage) GetSession(id uuid.UUID) (*GameInstance, error) {
	singleRes := i.mColl.FindOne(
		context.Background(),
		M{"game_id": M{"eq": id.String()}},
	)

	if err := singleRes.Err(); err != nil {
		log.Println("failed to fetch session data from mongo: ", err)
		return nil, err
	}
}

func (i *mongoSessionStorage) CloseSession(id uuid.UUID) error {

}

func (i *mongoSessionStorage) CheckExistance(id uuid.UUID) (bool, error) {

}

func (i *mongoSessionStorage) NumberOfGames() (uint, error) {

}

func ensureCollectionAndIndexes(mClient *mgo.Client) *mgo.Collection {
	mDB := mClient.Database(os.Getenv("MONGO_DBNAME"))
	err := mDB.CreateCollection(context.Background(), "avalonGames")
	if err != nil {
		//Collection already exists?
		if _, exist := err.(mgo.CommandError); exist {
			//Yes, no need to create one
			return
		} else {
			//No, its generic error
			log.Fatal(err)
		}
	}

	mColl := mDB.Collection("avalonGames")
	_, err = mColl.Indexes().CreateMany(
		context.Background(),
		[]mgo.IndexModel{
			mgo.IndexModel{
				Keys: "uuid",
				Options: options.Index().
					SetUnique(true),
			},
		},
	)

	if err != nil {
		log.Fatal("failed to create mongo indexes: ", err)
	}

	return mColl
}
