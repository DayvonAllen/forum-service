package repo

import (
	"context"
	"example.com/app/database"
	"example.com/app/domain"
	"example.com/app/helpers"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"log"
	"strconv"
	"time"
)

type ReplyRepoImpl struct {
	Reply     domain.Reply
	ReplyList []domain.Reply
}

func (r ReplyRepoImpl) Create(reply *domain.Reply) error {
	conn := database.MongoConnectionPool.Get().(*database.Connection)
	defer database.MongoConnectionPool.Put(conn)

	replyObj := new(domain.Reply)

	err := conn.PostsCollection.FindOne(context.TODO(), bson.D{{"_id", reply.ResourceId}}).Decode(&replyObj)

	if err != nil {
		return fmt.Errorf("resource not found")
	}

	_, err = conn.RepliesCollection.InsertOne(context.TODO(), &reply)

	if err != nil {
		return err
	}

	go func() {
		event := new(domain.Event)
		event.Action = "reply to forum post"
		event.Target = reply.ResourceId.String()
		event.ResourceId = reply.ResourceId
		event.ActorUsername = reply.AuthorUsername
		event.Message = reply.AuthorUsername + " replied to a post with the ID:" + reply.ResourceId.String()
		err = SendEventMessage(event, 0)
		if err != nil {
			fmt.Println("Error publishing...")
			return
		}
	}()

	return nil
}

func (r ReplyRepoImpl) FindAllRepliesByResourceId(resourceID primitive.ObjectID, username string, page string) (*[]domain.Reply, error) {
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

	cur, err := conn.RepliesCollection.Find(context.TODO(), bson.D{{"resourceId", resourceID}}, &findOptions)

	if err != nil {
		// ErrNoDocuments means that the filter did not match any documents in the collection
		if err == mongo.ErrNoDocuments {
			return nil, err
		}
		return nil, fmt.Errorf("error processing data")
	}

	if err = cur.All(context.TODO(), &r.ReplyList); err != nil {
		log.Fatal(err)
	}

	// Close the cursor once finished
	err = cur.Close(context.TODO())

	if err != nil {
		return nil, fmt.Errorf("error processing data")
	}

	replies := make([]domain.Reply, 0, len(r.ReplyList))
	for _, v := range r.ReplyList {
		v.CurrentUserLiked = helpers.CurrentUserInteraction(v.Likes, username)
		if !v.CurrentUserLiked {
			v.CurrentUserDisLiked = helpers.CurrentUserInteraction(v.Dislikes, username)
		}
		replies = append(replies, v)
	}
	return &replies, nil
}

func (r ReplyRepoImpl) UpdateById(id primitive.ObjectID, newContent string, edited bool, updatedTime time.Time, username string) error {
	conn := database.MongoConnectionPool.Get().(*database.Connection)
	defer database.MongoConnectionPool.Put(conn)

	opts := options.FindOneAndUpdate().SetUpsert(true)
	filter := bson.D{{"_id", id}, {"authorUsername", username}}
	update := bson.D{{"$set", bson.D{{"content", newContent}, {"edited", edited},
		{"updatedTime", updatedTime}}}}

	err := conn.RepliesCollection.FindOneAndUpdate(context.TODO(),
		filter, update, opts).Decode(&r.Reply)

	if err != nil {
		return fmt.Errorf("cannot update comment that you didn't write")
	}

	return nil
}

func (r ReplyRepoImpl) LikeReplyById(replyId primitive.ObjectID, username string) error {
	conn := database.MongoConnectionPool.Get().(*database.Connection)
	defer database.MongoConnectionPool.Put(conn)

	ctx := context.TODO()

	cur, err := conn.RepliesCollection.Find(ctx, bson.D{
		{"_id", replyId}, {"likes", username},
	})

	if err != nil {
		return err
	}

	if cur.Next(ctx) {
		return fmt.Errorf("you've already liked this comment")
	}

	// sets mongo's read and write concerns
	wc := writeconcern.New(writeconcern.WMajority())
	rc := readconcern.Snapshot()
	txnOpts := options.Transaction().SetWriteConcern(wc).SetReadConcern(rc)

	// set up for a transaction
	session, err := conn.StartSession()

	if err != nil {
		panic(err)
	}

	defer session.EndSession(context.Background())

	// execute this code in a logical transaction
	callback := func(sessionContext mongo.SessionContext) (interface{}, error) {

		filter := bson.D{{"_id", replyId}}
		update := bson.M{"$pull": bson.M{"dislikes": username}}

		res, err := conn.RepliesCollection.UpdateOne(context.TODO(), filter, update)

		if err != nil {
			return nil, err
		}

		if res.MatchedCount == 0 {
			return nil, fmt.Errorf("cannot find story")
		}

		err = conn.RepliesCollection.FindOne(context.TODO(),
			filter).Decode(&r.Reply)

		r.Reply.DislikeCount = len(r.Reply.Dislikes)

		update = bson.M{"$push": bson.M{"likes": username}, "$inc": bson.M{"likeCount": 1}, "$set": bson.D{{"dislikeCount", r.Reply.DislikeCount}}}

		filter = bson.D{{"_id", replyId}}

		_, err = conn.RepliesCollection.UpdateOne(context.TODO(),
			filter, update)

		if err != nil {
			return nil, err
		}

		return nil, err
	}

	_, err = session.WithTransaction(context.Background(), callback, txnOpts)

	if err != nil {
		return fmt.Errorf("failed to like comment")
	}

	go func() {
		event := new(domain.Event)
		event.Action = "like reply to post"
		event.Target = replyId.String()
		event.ResourceId = replyId
		event.ActorUsername = username
		event.Message = username + " liked a reply to a post with the ID:" + replyId.String()
		err = SendEventMessage(event, 0)
		if err != nil {
			fmt.Println("Error publishing...")
			return
		}
	}()

	return nil
}

func (r ReplyRepoImpl) DisLikeReplyById(replyId primitive.ObjectID, username string) error {
	conn := database.MongoConnectionPool.Get().(*database.Connection)
	defer database.MongoConnectionPool.Put(conn)

	ctx := context.TODO()

	cur, err := conn.RepliesCollection.Find(ctx, bson.D{
		{"_id", replyId}, {"dislikes", username},
	})

	if err != nil {
		return err
	}

	if cur.Next(ctx) {
		return fmt.Errorf("you've already disliked this comment")
	}

	// sets mongo's read and write concerns
	wc := writeconcern.New(writeconcern.WMajority())
	rc := readconcern.Snapshot()
	txnOpts := options.Transaction().SetWriteConcern(wc).SetReadConcern(rc)

	// set up for a transaction
	session, err := conn.StartSession()

	if err != nil {
		panic(err)
	}

	defer session.EndSession(context.Background())

	// execute this code in a logical transaction
	callback := func(sessionContext mongo.SessionContext) (interface{}, error) {

		filter := bson.D{{"_id", replyId}}
		update := bson.M{"$pull": bson.M{"likes": username}}

		res, err := conn.RepliesCollection.UpdateOne(context.TODO(), filter, update)

		if err != nil {
			return nil, err
		}

		if res.MatchedCount == 0 {
			return nil, fmt.Errorf("cannot find story")
		}

		err = conn.RepliesCollection.FindOne(context.TODO(),
			filter).Decode(&r.Reply)

		r.Reply.LikeCount = len(r.Reply.Likes)

		update = bson.M{"$push": bson.M{"dislikes": username}, "$inc": bson.M{"dislikeCount": 1}, "$set": bson.D{{"likeCount", r.Reply.LikeCount}}}

		filter = bson.D{{"_id", replyId}}

		_, err = conn.RepliesCollection.UpdateOne(context.TODO(),
			filter, update)

		if err != nil {
			return nil, err
		}

		return nil, err
	}

	_, err = session.WithTransaction(context.Background(), callback, txnOpts)

	if err != nil {
		return fmt.Errorf("failed to dislike comment")
	}

	go func() {
		event := new(domain.Event)
		event.Action = "dislike reply to post"
		event.Target = replyId.String()
		event.ResourceId = replyId
		event.ActorUsername = username
		event.Message = username + " disliked a reply to a post with the ID:" + replyId.String()
		err = SendEventMessage(event, 0)

		if err != nil {
			fmt.Println("Error publishing...")
			return
		}
	}()
	return nil
}

func (r ReplyRepoImpl) DeleteById(id primitive.ObjectID, username string) error {
	conn := database.MongoConnectionPool.Get().(*database.Connection)
	defer database.MongoConnectionPool.Put(conn)

	// sets mongo's read and write concerns
	wc := writeconcern.New(writeconcern.WMajority())
	rc := readconcern.Snapshot()
	txnOpts := options.Transaction().SetWriteConcern(wc).SetReadConcern(rc)

	// set up for a transaction
	session, err := conn.StartSession()

	if err != nil {
		panic(err)
	}

	defer session.EndSession(context.Background())

	// execute this code in a logical transaction
	callback := func(sessionContext mongo.SessionContext) (interface{}, error) {

		res, err := conn.RepliesCollection.DeleteOne(context.TODO(), bson.D{{"_id", id}, {"authorUsername", username}})

		if err != nil {
			panic(err)
		}

		if res.DeletedCount == 0 {
			panic(fmt.Errorf("failed to delete reply"))
		}

		return nil, err
	}

	_, err = session.WithTransaction(context.Background(), callback, txnOpts)

	if err != nil {
		return err
	}

	return nil
}

func NewReplyRepoImpl() ReplyRepoImpl {
	var replyRepoImpl ReplyRepoImpl

	return replyRepoImpl
}
