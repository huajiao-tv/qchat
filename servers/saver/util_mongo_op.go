// provides mongo message pure db operations
package main

import (
	"errors"
	"strings"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

/*
 * pure mongo find function, retrieve mongo records from mongo collection
 * @param query is query condition, support bson.M and bson.D
 * @param count is max return record count
 * @oaram sessions is mongo session group
 * @param db is database which data stored
 * @param collection is collection which stores data actually
 * @param sortFields is sorting fields
 *
 * @return (records, nil) if succeeded, otherwise (nil, error) will be returned
 */
func find(query interface{}, count int, sessions []*mgo.Session,
	db, collection string, sortFields ...string) ([]bson.M, error) {

	var errInfo string
	for _, session := range sessions {
		q := session.DB(db).C(collection).Find(query).Limit(count)
		if len(sortFields) > 0 {
			q = q.Sort(sortFields...)
		}

		var result []bson.M
		if err := q.All(&result); err != nil {
			if err == mgo.ErrNotFound {
				return nil, err
			}

			errInfo = strings.Join([]string{errInfo, err.Error()}, "; ")
			continue
		}

		return result, nil
	}

	return nil, errors.New(errInfo)
}

/*
 * pure mongo findOne function, retrieve one mongo record from mongo collection
 * @param query is query condition, support bson.M and bson.D
 * @oaram sessions is mongo session group
 * @param db is database which data stored
 * @param collection is collection which stores data actually
 * @param sortFields is sorting fields
 *
 * @return (record, nil) if succeeded, otherwise (nil, error) will be returned
 */
func findOne(query interface{}, sessions []*mgo.Session,
	db, collection string, sortFields ...string) (*bson.M, error) {
	var errInfo string
	for _, session := range sessions {
		q := session.DB(db).C(collection).Find(query)
		if len(sortFields) > 0 {
			q = q.Sort(sortFields...)
		}

		var result bson.M
		if err := q.One(&result); err != nil {
			if err == mgo.ErrNotFound {
				return nil, err
			}

			errInfo = strings.Join([]string{errInfo, err.Error()}, "; ")
			continue
		}

		return &result, nil
	}

	return nil, errors.New(errInfo)
}

/*
 * pure mongo findAndModify function, find and then modify record if found
 * @oaram sessiond is mongo session group
 * @param db is database which data stored
 * @param collection is collection which stores data actually
 * @param query is query condition, support bson.M and bson.D
 * @param update is update statement
 * @param upsert indicates whether insert new record if not found
 * @param remove indicates whether remove the record if found
 * @param updated is used to return new record if need
 *
 * @return (nil) if succeeded, otherwise (error) will be returned
 */
func findAndModify(sessions []*mgo.Session, db, collection string,
	query, update interface{}, upsert, remove bool, updated interface{}) error {
	change := mgo.Change{
		Update:    update,
		ReturnNew: true,
		Upsert:    upsert,
		Remove:    remove,
	}

	var errInfo string
	for _, session := range sessions {
		if _, err := session.DB(db).C(collection).Find(query).Apply(change, updated); err != nil {
			if err == mgo.ErrNotFound {
				return err
			}

			errInfo = strings.Join([]string{errInfo, err.Error()}, "; ")
			continue
		}
		return nil
	}

	return errors.New(errInfo)
}
