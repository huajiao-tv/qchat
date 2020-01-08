package main

import (
	"flag"
	"fmt"
	"sync"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	mongoDBs = map[string]string{
		//#"mongodb://user@password@oldMongoAddr/?authMechanism=MONGODB-CR": "mongodb://user@password@newMongoAddr/?authMechanism=MONGODB-CR",

	}

	dbTables = map[string]string{
		"msg_im_inboxes":  "msg_im_inbox",
		"msg_im_lastr":    "im_last_reads",
		"msg_im_outboxes": "msg_im_outbox",
		"msg_imid":        "im_latest_id",
		"msg_lastr":       "last_reads",
		"msg_messageid":   "latest_id",
		"msg_messages":    "message_data",
		"msg_publics":     "public_message_data",
		"msg_pushes":      "push_message_data",
		"msg_pushid":      "push_latest_id",
	}

	modifyFields = map[string]string{
		"msg_im_inboxes":  "modified",
		"msg_im_lastr":    "latest_modified",
		"msg_im_outboxes": "modified",
		"msg_imid":        "latest_modified",
		"msg_lastr":       "latest_modified",
		"msg_messageid":   "latest_modified",
		"msg_messages":    "modified",
		"msg_publics":     "",
		"msg_pushes":      "modified",
		"msg_pushid":      "latest_modified",
	}

	uniqueKeys = map[string][]string{
		"msg_im_inboxes":  {"owner", "msg_id"},
		"msg_im_lastr":    {"jid"},
		"msg_im_outboxes": {"owner", "msg_id"},
		"msg_imid":        {"jid"},
		"msg_lastr":       {"jid"},
		"msg_messageid":   {"jid"},
		"msg_messages":    {"jid", "msg_id"},
		"msg_publics":     {"msg_id"},
		"msg_pushes":      {"jid"},
		"msg_pushid":      {"jid"},
	}

	beginTs time.Time
)

func init() {
	//t, _ := time.Parse(time.RFC3339, "2018-12-01T07:00:00Z")

	// TODO
	t, _ := time.Parse(time.RFC3339, "2018-12-19T00:00:00Z")
	beginTs = time.Unix(0, t.UnixNano()/1e6*1e6)
}

func main() {
	//// TODO
	//for _, newMongo := range mongoDBs {
	//	fmt.Println("WARNING: cleaning db:", newMongo)
	//	cleanNew(newMongo)
	//}
	//return

	emptyDB := false
	flag.BoolVar(&emptyDB, "empty", false, "group members")
	flag.Parse()

	if emptyDB {
		fmt.Println("create collection copy from old mongo db")
		createCollectionData()
	} else {
		fmt.Println("porting data from old mongo db")
		portCollection()
	}
}

//func cleanNew(newMongo string) {
//	for db, collection := range dbTables {
//		// connect new
//		newSession, err := mgo.DialWithTimeout(newMongo, 5*time.Second)
//		if err != nil {
//			panic(err)
//		}
//		defer newSession.Close()
//		fmt.Println(newSession.DB(db).C(collection).RemoveAll(nil))
//	}
//}

func createCollectionData() {
	start := time.Now()

	wg := sync.WaitGroup{}
	for oldMongo, newMongo := range mongoDBs {
		for db, collection := range dbTables {
			wg.Add(1)

			go func(oldMongo, newMongo, db, collection string) {
				if db == "msg_publics" {
					portPublicCollection(oldMongo, newMongo, db, collection)
				} else {
					createNormalCollection(oldMongo, newMongo, db, collection)
				}
				wg.Done()
			}(oldMongo, newMongo, db, collection)
		}
	}

	wg.Wait()
	fmt.Println("done..., cost:", time.Now().Sub(start))
	fmt.Println()
	fmt.Println()
}

func createNormalCollection(oldMongo, newMongo, db, collection string) {
	// connect
	oldSession, err := mgo.DialWithTimeout(oldMongo, 5*time.Second)
	if err != nil {
		fmt.Println("conn failed", err, oldMongo, db, collection)
		panic(err)
	}
	defer oldSession.Close()
	oldSession.SetSyncTimeout(time.Minute)
	oldSession.SetSocketTimeout(time.Minute)
	oldSession.SetCursorTimeout(0)

	oldCollection := oldSession.DB(db).C(collection)
	c, err := oldCollection.Count()
	if err == nil && c == 0 {
		fmt.Println("empty collection", oldMongo, db, collection)
		return
	} else if err != nil {
		fmt.Println("get count failed", oldMongo, db, collection)
		panic(err)
	}

	field, ok := modifyFields[db]
	if !ok || field == "" {
		panic("find db field failed:" + db)
	}

	query := []bson.M{{"$match": bson.M{field: bson.M{"$lte": beginTs}}}}
	pipe := oldCollection.Pipe(query)
	iter := pipe.Iter()

	// connect new
	newSession, err := mgo.DialWithTimeout(newMongo, 5*time.Second)
	if err != nil {
		fmt.Println("conn new failed", err, newMongo, db, collection)
		panic(err)
	}
	defer newSession.Close()
	newSession.SetSyncTimeout(time.Minute)
	newSession.SetSocketTimeout(time.Minute)

	newCollection := newSession.DB(db).C(collection)

	// copy
	count := 0
	complete := false
	for {
		batch := make([]interface{}, 0, 500)
		failed := 0
		for i := 0; i < 500; i++ {
			var result bson.M
			if !iter.Next(&result) {
				if iter.Err() == nil {
					failed = 0
					complete = true
					fmt.Println("complete", iter.Err(), oldMongo, db, collection)
					break
				} else {
					failed++
					if failed > 10 {
						fmt.Println("iteration failed", iter.Err())
						panic(iter.Err())
					} else {
						fmt.Println("iteration error, sleep & retry...", iter.Err())
						time.Sleep(time.Second)
						continue
					}
				}
			}
			delete(result, "_id")
			batch = append(batch, &result)
		}
		if len(batch) == 0 {
			break
		}

		count += len(batch)
		err = newCollection.Insert(batch...)
		if err != nil {
			fmt.Println("write to new failed", err, newMongo, db, collection)
			panic(err)
		}

		if complete {
			break
		}
	}

	fmt.Println("porting count", oldMongo, db, collection, count)
}

func portCollection() {
	for {
		start := time.Now()
		ts := bson.Now()

		wg := sync.WaitGroup{}
		for oldMongo, newMongo := range mongoDBs {
			for db, collection := range dbTables {
				wg.Add(1)

				go func(oldMongo, newMongo, db, collection string) {
					if db == "msg_publics" {
						portPublicCollection(oldMongo, newMongo, db, collection)
					} else {
						portNormalCollection(oldMongo, newMongo, db, collection, beginTs)
					}
					wg.Done()
				}(oldMongo, newMongo, db, collection)
			}
		}

		wg.Wait()
		beginTs = ts

		cost := time.Now().Sub(start)
		fmt.Println("done..., cost:", cost)

		if cost < time.Millisecond*500 {
			time.Sleep(time.Millisecond * 500)
		}
		fmt.Println()
		fmt.Println()
	}
}

func portNormalCollection(oldMongo, newMongo, db, collection string, ts time.Time) {
	// connect
	oldSession, err := mgo.DialWithTimeout(oldMongo, 5*time.Second)
	if err != nil {
		fmt.Println("conn failed", err, oldMongo, db, collection)
		panic(err)
	}
	defer oldSession.Close()
	oldSession.SetSyncTimeout(time.Minute)
	oldSession.SetSocketTimeout(time.Minute)
	oldSession.SetCursorTimeout(0)

	oldCollection := oldSession.DB(db).C(collection)
	c, err := oldCollection.Count()
	if err == nil && c == 0 {
		fmt.Println("empty collection", oldMongo, db, collection)
		return
	} else if err != nil {
		fmt.Println("get count failed", oldMongo, db, collection)
		panic(err)
	}

	field, ok := modifyFields[db]
	if !ok || field == "" {
		panic("find db field failed:" + db)
	}

	query := []bson.M{{"$match": bson.M{field: bson.M{"$gte": ts}}}}
	pipe := oldCollection.Pipe(query)
	iter := pipe.Iter()

	// connect new
	newSession, err := mgo.DialWithTimeout(newMongo, 5*time.Second)
	if err != nil {
		fmt.Println("conn new failed", err, newMongo, db, collection)
		panic(err)
	}
	defer newSession.Close()
	newSession.SetSyncTimeout(time.Minute)
	newSession.SetSocketTimeout(time.Minute)

	newCollection := newSession.DB(db).C(collection)

	uniqueKey := uniqueKeys[db]
	selector := bson.M{}

	// copy
	count := 0
	var result bson.M
	for {
		if !iter.Next(&result) {
			if iter.Err() == nil {
				fmt.Println("complete", iter.Err(), oldMongo, db, collection)
				break
			} else {
				fmt.Println("iteration failed", iter.Err())
				panic(iter.Err())
			}
		}
		count += 1

		for _, key := range uniqueKey {
			selector[key] = result[key]
		}

		delete(result, "_id")
		_, err = newCollection.Upsert(selector, &result)
		if err != nil {
			fmt.Println("write to new failed", err, newMongo, db, collection)
			panic(err)
		}
	}

	fmt.Println("porting count", oldMongo, db, collection, count)
}

func portPublicCollection(oldMongo, newMongo, db, collection string) {
	// connect
	oldSession, err := mgo.DialWithTimeout(oldMongo, 5*time.Second)
	if err != nil {
		fmt.Println("conn failed", err, oldMongo, db, collection)
		panic(err)
	}
	defer oldSession.Close()
	oldSession.SetSyncTimeout(time.Minute)
	oldSession.SetSocketTimeout(time.Minute)
	oldSession.SetCursorTimeout(0)

	oldCollection := oldSession.DB(db).C(collection)
	c, err := oldCollection.Count()
	if err == nil && c == 0 {
		fmt.Println("empty collection", oldMongo, db, collection)
		return
	} else if err != nil {
		fmt.Println("get count failed", oldMongo, db, collection)
		panic(err)
	}

	// query all
	query := []bson.M{{"$match": bson.M{"msg_id": bson.M{"$gte": 0}}}}
	pipe := oldCollection.Pipe(query)
	iter := pipe.Iter()

	// connect new
	newSession, err := mgo.DialWithTimeout(newMongo, 5*time.Second)
	if err != nil {
		fmt.Println("conn new failed", err, newMongo, db, collection)
		panic(err)
	}
	defer newSession.Close()
	newSession.SetSyncTimeout(time.Minute)
	newSession.SetSocketTimeout(time.Minute)

	newCollection := newSession.DB(db).C(collection)

	// copy
	count := 0
	var result bson.M
	for {
		if !iter.Next(&result) {
			if iter.Err() == nil {
				fmt.Println("complete", iter.Err(), oldMongo, db, collection)
				break
			} else {
				fmt.Println("iteration failed", iter.Err())
				panic(iter.Err())
			}
		}
		count += 1

		delete(result, "_id")
		_, err = newCollection.Upsert(bson.M{"msg_id": result["msg_id"]}, &result)
		if err != nil {
			fmt.Println("write to new failed", err, newMongo, db, collection)
			panic(err)
		}
	}

	fmt.Println("porting count", oldMongo, db, collection, count)
}
