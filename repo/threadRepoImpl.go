package repo

import (
	"context"
	"example.com/app/database"
	"example.com/app/domain"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"strconv"
)

type ThreadRepoImpl struct {
	forum             domain.Forum
	thread            domain.Thread
	threadList        []domain.Thread
	threadPreviewList []domain.ThreadPreview
}

func (t ThreadRepoImpl) FindAll(page string, ctx context.Context) (*[]domain.ThreadPreview, error) {
	conn := database.MongoConnectionPool.Get().(*database.Connection)
	defer database.MongoConnectionPool.Put(conn)

	findOptions := options.FindOptions{}
	perPage := 10
	pageNumber, err := strconv.Atoi(page)

	if err != nil {
		return nil, fmt.Errorf("page must be a number")
	}
	findOptions.SetSkip((int64(pageNumber) - 1) * int64(perPage))
	findOptions.SetLimit(int64(perPage))

	// Get all users
	cur, err := conn.ThreadsCollection.Find(ctx, bson.M{}, &findOptions)

	if err != nil {
		return nil, err
	}

	if err = cur.All(ctx, &t.threadPreviewList); err != nil {
		log.Fatal(err)
	}

	return &t.threadPreviewList, nil
}

func (t ThreadRepoImpl) FindByName(threadName string, username string) (*domain.Thread, error) {
	conn := database.MongoConnectionPool.Get().(*database.Connection)
	defer database.MongoConnectionPool.Put(conn)

	err := conn.ThreadsCollection.FindOne(context.TODO(), bson.D{{"name", threadName}}).Decode(&t.thread)

	if err != nil {
		return nil, err
	}

	go func() {
		event := new(domain.Event)
		event.Action = "view thread"
		event.Target = t.thread.Id.String()
		event.ResourceId = t.thread.Id
		event.ActorUsername = username
		event.Message = username + " viewed a thread"
		err = SendEventMessage(event, 0)
		if err != nil {
			fmt.Println("Error publishing...")
			return
		}
	}()

	return &t.thread, nil
}

func (t ThreadRepoImpl) Create(thread *domain.Thread) error {
	conn := database.MongoConnectionPool.Get().(*database.Connection)
	defer database.MongoConnectionPool.Put(conn)

	cur, err := conn.ThreadsCollection.Find(context.TODO(), bson.D{{"name", thread.Name}})

	if err != nil {
		return fmt.Errorf("error processing data")
	}

	if !cur.Next(context.TODO()) {
		_, err = conn.ThreadsCollection.InsertOne(context.TODO(), thread)

		if err != nil {
			return fmt.Errorf("error processing data")
		}

		go func() {
			event := new(domain.Event)
			event.Action = "create thread"
			event.Target = thread.Id.String()
			event.ResourceId = thread.Id
			event.ActorUsername = thread.OwnerUsername
			event.Message = thread.OwnerUsername + " created a thread"
			err = SendEventMessage(event, 0)
			if err != nil {
				fmt.Println("Error publishing...")
				return
			}
		}()

		return nil
	}

	return fmt.Errorf("thread with that name already exists, thread names must be unique")
}

func (t ThreadRepoImpl) DeleteByID(id primitive.ObjectID, username string) error {
	conn := database.MongoConnectionPool.Get().(*database.Connection)
	defer database.MongoConnectionPool.Put(conn)

	_, err := conn.ThreadsCollection.DeleteOne(context.TODO(), bson.D{{"_id", id}, {"ownerUsername", username}})

	if err != nil {
		return err
	}

	go func() {
		event := new(domain.Event)
		event.Action = "delete thread"
		event.Target = id.String()
		event.ResourceId = id
		event.ActorUsername = username
		event.Message = username + " deleted a thread"
		err = SendEventMessage(event, 0)
		if err != nil {
			fmt.Println("Error publishing...")
			return
		}
	}()

	return nil
}

func NewThreadRepoImpl() ThreadRepoImpl {
	var threadRepoImpl ThreadRepoImpl

	return threadRepoImpl
}
