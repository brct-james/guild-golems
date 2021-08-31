package db

import (
	"context"
	"errors"
	"fmt"
	"log"

	goredis "github.com/go-redis/redis/v8"
	rejson "github.com/nitishm/go-rejson/v4"
)

var verbose = false

var ErrNil = errors.New("db.go: nil returned")

// Copied from github.com/redigo/redis
// Bytes is a helper that converts a command reply to a slice of bytes. If err
// is not equal to nil, then Bytes returns nil, err. Otherwise Bytes converts
// the reply to a slice of bytes as follows:
//
//  Reply type      Result
//  bulk string     reply, nil
//  simple string   []byte(reply), nil
//  nil             nil, ErrNil
//  other           nil, error
func Bytes(reply interface{}, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}
	switch reply := reply.(type) {
	case []byte:
		return reply, nil
	case string:
		return []byte(reply), nil
	case nil:
		return nil, ErrNil
	case error:
		return nil, reply
	}
	return nil, fmt.Errorf("db.go: unexpected type for Bytes, got type %T", reply)
}

//Generic set data function
func JsonSetData(rh *rejson.Handler, key string, data interface{}) {
	if verbose {
		log.Printf("New attempt JsonSetData")
		log.Printf("Key: '%s', Data: '%s'", key, data)
	}
	res, err := rh.JSONSet(key, ".", data)
	if err != nil {
		log.Printf("Failed to JSONSet, reason: '%s'", err.Error())
		return
	}
	if res.(string) == "OK" {
		if verbose {
			fmt.Printf("JsonSetData Success: %s\n", res)
		}
	} else {
		fmt.Println("JsonSetData Failed to Set: " + res.(string))
	}
}

//Generic get data function
//Returns marshalled json so make sure to unmarshall externally
func JsonGetData(rh *rejson.Handler, key string) []uint8 {
	if verbose {
		log.Printf("New attempt JsonGetData. Key: '%s'", key)
	}
	dataJSON, err := Bytes(rh.JSONGet(key, "."))
	if err != nil {
		log.Printf("Failed to JSONGet, reason: '%s'", err.Error())
		return nil
	}
	return dataJSON
}

//Database struct for holding the handler/client
type Database struct {
	Rejson *rejson.Handler
	Goredis *goredis.Client
}

//Create new database and return populated struct
func NewDatabase(redisAddr string, dbNum int) Database {
	rh := rejson.NewReJSONHandler()

	//GoRedis Client
	cli := goredis.NewClient(&goredis.Options{Addr: redisAddr, DB: dbNum})
	rh.SetGoRedisClient(cli)
	db := Database{
		Rejson: rh,
		Goredis: cli,
	}
	return db
}

//Close passed database
func CloseDB(db Database) {
	if err := db.Goredis.FlushAll(context.Background()).Err(); err != nil {
		log.Fatalf("goredis - failed to flush: %v", err)
	}
	if err := db.Goredis.Close(); err != nil {
		log.Fatalf("goredis - failed to communicate to redis-server: %v", err)
	}
}