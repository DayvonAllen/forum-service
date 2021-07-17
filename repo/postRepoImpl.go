package repo

import (
	"context"
	"example.com/app/database"
	"example.com/app/domain"
	"example.com/app/helper"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"log"
	"sync"
	"time"
)

type PostRepoImpl struct {
	Post  domain.Post
	Reply domain.Reply
	Posts []domain.Post
}

func (p PostRepoImpl) Create(post *domain.Post) error {
	conn := database.MongoConnectionPool.Get().(*database.Connection)
	defer database.MongoConnectionPool.Put(conn)

	thread := new(domain.Thread)

	err := conn.ThreadsCollection.FindOne(context.TODO(), bson.D{{"_id", post.ResourceId}}).Decode(&thread)

	if err != nil {
		return fmt.Errorf("resource not found")
	}

	_, err = conn.PostsCollection.InsertOne(context.TODO(), &post)

	if err != nil {
		return err
	}

	go func() {
		event := new(domain.Event)
		event.Action = "post on thread"
		event.Target = post.ResourceId.String()
		event.ResourceId = post.ResourceId
		event.ActorUsername = post.AuthorUsername
		event.Message = post.AuthorUsername + " created a post"
		err = SendEventMessage(event, 0)
		if err != nil {
			fmt.Println("Error publishing...")
			return
		}
	}()

	return nil
}

func (p PostRepoImpl) FindAllPostsByResourceId(id primitive.ObjectID, username string) (*[]domain.Post, error) {
	conn := database.MongoConnectionPool.Get().(*database.Connection)
	defer database.MongoConnectionPool.Put(conn)

	cur, err := conn.ThreadsCollection.Find(context.TODO(), bson.D{{"resourceId", id}})

	if err != nil {
		// ErrNoDocuments means that the filter did not match any documents in the collection
		if err == mongo.ErrNoDocuments {
			return nil, err
		}
		return nil, fmt.Errorf("error processing data")
	}

	if err = cur.All(context.TODO(), &p.Posts); err != nil {
		log.Fatal(err)
	}

	// Close the cursor once finished
	err = cur.Close(context.TODO())

	if err != nil {
		return nil, fmt.Errorf("error processing data")
	}

	comments := make([]domain.Post, 0, len(p.Posts))
	var wg sync.WaitGroup
	for _, v := range p.Posts {
		wg.Add(2)

		go func() {
			defer wg.Done()

			v.CurrentUserLiked = helper.CurrentUserInteraction(v.Likes, username)
			if !v.CurrentUserLiked {
				v.CurrentUserDisLiked = helper.CurrentUserInteraction(v.Dislikes, username)
			}

			return
		}()

		go func() {
			defer wg.Done()

			replies, err := ReplyRepoImpl{}.FindAllRepliesByResourceId(v.Id, username)

			v.Replies = replies

			if err != nil {
				panic(fmt.Errorf("error fetching data..."))
			}
			return
		}()

		wg.Wait()

		comments = append(comments, v)
	}
	return &comments, nil
}

func (p PostRepoImpl) UpdateById(id primitive.ObjectID, newContent string, edited bool, updatedTime time.Time, username string) error {
	conn := database.MongoConnectionPool.Get().(*database.Connection)
	defer database.MongoConnectionPool.Put(conn)

	opts := options.FindOneAndUpdate().SetUpsert(true)
	filter := bson.D{{"_id", id}, {"authorUsername", username}}
	update := bson.D{{"$set", bson.D{{"content", newContent}, {"edited", edited},
		{"updatedTime", updatedTime}}}}

	err := conn.ThreadsCollection.FindOneAndUpdate(context.TODO(),
		filter, update, opts).Decode(&p.Post)

	if err != nil {
		return fmt.Errorf("cannot update comment that you didn't write")
	}

	return nil
}

func (p PostRepoImpl) LikePostById(postId primitive.ObjectID, username string) error {
	conn := database.MongoConnectionPool.Get().(*database.Connection)
	defer database.MongoConnectionPool.Put(conn)

	ctx := context.TODO()

	cur, err := conn.PostsCollection.Find(ctx, bson.D{
		{"_id", postId}, {"likes", username},
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

		filter := bson.D{{"_id", postId}}
		update := bson.M{"$pull": bson.M{"dislikes": username}}

		res, err := conn.PostsCollection.UpdateOne(context.TODO(), filter, update)

		if err != nil {
			return nil, err
		}

		if res.MatchedCount == 0 {
			return nil, fmt.Errorf("cannot find story")
		}

		err = conn.PostsCollection.FindOne(context.TODO(),
			filter).Decode(&p.Post)

		p.Post.DislikeCount = len(p.Post.Dislikes)

		update = bson.M{"$push": bson.M{"likes": username}, "$inc": bson.M{"likeCount": 1}, "$set": bson.D{{"dislikeCount", p.Post.DislikeCount}}}

		filter = bson.D{{"_id", postId}}

		_, err = conn.PostsCollection.UpdateOne(context.TODO(),
			filter, update)

		if err != nil {
			return nil, err
		}

		go func() {
			event := new(domain.Event)
			event.Action = "like post"
			event.Target = postId.String()
			event.ResourceId = postId
			event.ActorUsername = username
			event.Message = username + " liked a post with the ID:" + postId.String()
			err = SendEventMessage(event, 0)
			if err != nil {
				fmt.Println("Error publishing...")
				return
			}
		}()

		return nil, err
	}

	_, err = session.WithTransaction(context.Background(), callback, txnOpts)

	if err != nil {
		return fmt.Errorf("failed to like comment")
	}

	return nil
}

func (p PostRepoImpl) DisLikePostById(postId primitive.ObjectID, username string) error {
	conn := database.MongoConnectionPool.Get().(*database.Connection)
	defer database.MongoConnectionPool.Put(conn)

	ctx := context.TODO()

	cur, err := conn.PostsCollection.Find(ctx, bson.D{
		{"_id", postId}, {"dislikes", username},
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

		filter := bson.D{{"_id", postId}}
		update := bson.M{"$pull": bson.M{"likes": username}}

		res, err := conn.PostsCollection.UpdateOne(context.TODO(), filter, update)

		if err != nil {
			return nil, err
		}

		if res.MatchedCount == 0 {
			return nil, fmt.Errorf("cannot find story")
		}

		err = conn.PostsCollection.FindOne(context.TODO(),
			filter).Decode(&p.Post)

		p.Post.LikeCount = len(p.Post.Likes)

		update = bson.M{"$push": bson.M{"dislikes": username}, "$inc": bson.M{"dislikeCount": 1}, "$set": bson.D{{"likeCount", p.Post.LikeCount}}}

		filter = bson.D{{"_id", postId}}

		_, err = conn.PostsCollection.UpdateOne(context.TODO(),
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
		event.Action = "dislike post"
		event.Target = postId.String()
		event.ResourceId = postId
		event.ActorUsername = username
		event.Message = username + " disliked a post with the ID:" + postId.String()
		err = SendEventMessage(event, 0)

		if err != nil {
			fmt.Println("Error publishing...")
			return
		}
	}()

	return nil
}


func (p PostRepoImpl) DeleteById(id primitive.ObjectID, username string) error {
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
		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()

			res, err := conn.PostsCollection.DeleteOne(context.TODO(), bson.D{{"_id", id}, {"authorUsername", username}})

			if err != nil {
				panic(err)
			}

			if res.DeletedCount == 0 {
				panic(fmt.Errorf("you can't delete a comment that you didn't create"))
			}

			return
		}()

		go func() {
			defer wg.Done()

			_, err = conn.RepliesCollection.DeleteMany(context.TODO(), bson.D{{"resourceId", id}})
			if err != nil {
				panic(err)
			}

			return
		}()

		wg.Wait()

		return nil, err
	}

	_, err = session.WithTransaction(context.Background(), callback, txnOpts)

	if err != nil {
		return fmt.Errorf("failed to delete reply")
	}

	return nil
}

func (p PostRepoImpl) DeleteManyById(id primitive.ObjectID, username string) error {
	conn := database.MongoConnectionPool.Get().(*database.Connection)
	defer database.MongoConnectionPool.Put(conn)

	err := conn.PostsCollection.FindOne(context.TODO(), bson.D{{"resourceId", id}}).Decode(&p.Post)

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
		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()

			res, err := conn.PostsCollection.DeleteMany(context.TODO(), bson.D{{"resourceId", id}, {"authorUsername", username}})
			if err != nil {
				panic(err)
			}

			if res.DeletedCount == 0 {
				panic(err)
			}

			return
		}()

		go func() {
			defer wg.Done()

			_, err = conn.RepliesCollection.DeleteMany(context.TODO(), bson.D{{"resourceId", p.Post.Id}})
			if err != nil {
				panic(err)
			}

			return
		}()

		wg.Wait()
		return nil, err
	}

	_, err = session.WithTransaction(context.Background(), callback, txnOpts)

	if err != nil {
		return err
	}

	return nil
}

func NewPostRepoImpl() PostRepoImpl {
	var commentRepoImpl PostRepoImpl

	return commentRepoImpl
}

