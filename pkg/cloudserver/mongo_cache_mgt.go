/*
 * Copyright (c) The One True Way 2020. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package cloudserver

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func DoMongoStuff() error {
	uid := "admin"
	pwd := "NetApp1!!"
	host := "localhost"
	urlToUse := fmt.Sprintf("mongodb://%s:%s@%s:27017", uid, pwd, host)
	client, err := mongo.NewClient(options.Client().ApplyURI(urlToUse))
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		return err
	}
	defer client.Disconnect(ctx)
	collection := client.Database("natssync").Collection("messages.client1")
	m := CachedMsg{ClientID: "client1", Data: "some encryppted data", Timestamp: time.Now()}
	ictx := context.Background()
	one, err := collection.InsertOne(ictx, &m)
	fmt.Println(one)
	findOptions := options.Find()
	// Sort by `price` field descending
	findOptions.SetSort(bson.D{{"$natural", 1}})
	cur, err := collection.Find(nil, bson.D{}, findOptions)
	if err != nil {
		return err
	}
	for cur.Next(ictx) {
		var m1 CachedMsg
		cur.Decode(&m1)
		fmt.Printf("%s %+v", m1.Data, m1.Timestamp)
	}

	return err
}
