package utils

import (
	"crypto/rand"
	"encoding/binary"
	"sync/atomic"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var objectIDCounter uint32 = 0

func ObjectID() bson.ObjectID {
	counter := objectIDCounter
	objectIDCounter++

	defer func() {
		if counter == 0 {
			objectIDCounter = 0
		}
	}()

	var id bson.ObjectID

	// 4byte timestamp
	binary.BigEndian.PutUint32(id[0:4], uint32(time.Now().Unix()))
	// 5 byte random
	_, _ = rand.Read(id[4:9])
	// 3 byte increment couter
	c := atomic.AddUint32(&objectIDCounter, 1)
	id[9] = byte(c >> 16)
	id[10] = byte(c >> 8)
	id[11] = byte(c)

	return id
}
