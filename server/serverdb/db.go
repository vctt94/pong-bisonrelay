package serverdb

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"time"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	bolt "go.etcd.io/bbolt"
)

var (
	tipsBucket = []byte("receivedTips")
)

// BoltDB implements the ServerDB interface using bbolt as the backend.
type boltDB struct {
	db *bolt.DB
}

var _ ServerDB = (*boltDB)(nil)

// NewBoltDB initializes the BoltDB instance with a bbolt database.
func NewBoltDB(dbPath string) (ServerDB, error) {
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}
	// Initialize the main and status buckets on the first run.
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(tipsBucket)
		return err
	})
	if err != nil {
		db.Close()
		return nil, err
	}
	return &boltDB{db: db}, nil
}

// StoreUnprocessedTip stores a tip under the Uid sub-bucket in the main tips bucket.
func (b *boltDB) StoreUnprocessedTip(ctx context.Context, tip *types.ReceivedTip) error {
	payload := ReceivedTipWrapper{
		Tip:    tip,
		Status: StatusUnprocessed,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return b.db.Update(func(tx *bolt.Tx) error {
		mainBucket := tx.Bucket(tipsBucket)
		if mainBucket == nil {
			return errors.New("main bucket not found")
		}

		userBucket, err := mainBucket.CreateBucketIfNotExists(payload.Tip.Uid)
		if err != nil {
			return err
		}

		sequenceId := make([]byte, 8)
		binary.BigEndian.PutUint64(sequenceId, tip.SequenceId)
		if userBucket.Get(sequenceId) != nil {
			return ErrAlreadyStoredRV
		}

		return userBucket.Put(sequenceId, data)
	})
}

// MarkTipAsSending updates the status of a specific tip to "sending".
func (b *boltDB) UpdateTipStatus(ctx context.Context, uid []byte, tipID []byte, status TipStatus) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		mainBucket := tx.Bucket(tipsBucket)
		if mainBucket == nil {
			return errors.New("main bucket not found")
		}

		// Locate the sub-bucket for the specified uid.
		userBucket := mainBucket.Bucket(uid)
		if userBucket == nil {
			return errors.New("user bucket not found for the given UID")
		}

		// Retrieve the tip data by tipID within the user bucket.
		data := userBucket.Get(tipID)
		if data == nil {
			return errors.New("tip not found")
		}

		// Decode the data into ReceivedTipWrapper
		wrapper := &ReceivedTipWrapper{}
		if err := json.Unmarshal(data, wrapper); err != nil {
			return err
		}

		// Update the status
		wrapper.Status = status

		// Re-encode the updated wrapper and store it back in the database.
		updatedData, err := json.Marshal(wrapper)
		if err != nil {
			return err
		}

		return userBucket.Put(tipID, updatedData)
	})
}

// Close closes the bbolt database.
func (b *boltDB) Close() error {
	return b.db.Close()
}

// FetchReceivedTipsByUID retrieves all tips with a specific status for a specific user by Uid.
func (b *boltDB) FetchReceivedTipsByUID(ctx context.Context, uid zkidentity.ShortID, status TipStatus) ([]*types.ReceivedTip, error) {
	var unprocessedTips []*types.ReceivedTip

	err := b.db.View(func(tx *bolt.Tx) error {
		mainBucket := tx.Bucket(tipsBucket)
		if mainBucket == nil {
			return errors.New("tip bucket not found")
		}

		// Retrieve the sub-bucket for the specified Uid
		userBucket := mainBucket.Bucket(uid[:])
		if userBucket == nil {
			// No tips found for this Uid
			return nil
		}

		// Iterate through each tip in the Uid's sub-bucket
		return userBucket.ForEach(func(tipID, v []byte) error {

			// Decode the protobuf payload into a ReceivedTip struct
			tip := &ReceivedTipWrapper{}
			if err := json.Unmarshal(v, tip); err != nil {
				return err
			}

			// Filter based on the specified status
			if tip.Status == status {
				// Add the decoded tip to the unprocessed tips slice
				unprocessedTips = append(unprocessedTips, tip.Tip)
			}
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return unprocessedTips, nil
}

// FetchAllReceivedTipsByUID retrieves all tips for a specific user by Uid, regardless of status.
func (b *boltDB) FetchAllReceivedTipsByUID(ctx context.Context, uid zkidentity.ShortID) ([]ReceivedTipWrapper, error) {
	var allTips []ReceivedTipWrapper

	err := b.db.View(func(tx *bolt.Tx) error {
		mainBucket := tx.Bucket(tipsBucket)
		if mainBucket == nil {
			return errors.New("tip bucket not found")
		}

		// Retrieve the sub-bucket for the specified Uid
		userBucket := mainBucket.Bucket(uid[:])
		if userBucket == nil {
			// No tips found for this Uid
			return nil
		}

		// Iterate through each tip in the Uid's sub-bucket
		return userBucket.ForEach(func(tipID, v []byte) error {
			// Decode the JSON payload into a ReceivedTipWrapper struct
			tip := &ReceivedTipWrapper{}
			if err := json.Unmarshal(v, tip); err != nil {
				return err
			}

			// Add the decoded tip to the allTips slice
			allTips = append(allTips, *tip)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return allTips, nil
}

// FetchUnprocessedTips retrieves all unprocessed tips for all users.
func (b *boltDB) FetchUnprocessedTips(ctx context.Context) (map[zkidentity.ShortID][]*types.ReceivedTip, error) {
	unprocessedTips := make(map[zkidentity.ShortID][]*types.ReceivedTip)

	err := b.db.View(func(tx *bolt.Tx) error {
		mainBucket := tx.Bucket(tipsBucket)
		if mainBucket == nil {
			return errors.New("main bucket not found")
		}

		// Iterate over each UID's sub-bucket within the main bucket
		return mainBucket.ForEach(func(uid, _ []byte) error {
			userBucket := mainBucket.Bucket(uid)
			if userBucket == nil {
				// Skip if there's no bucket for this UID
				return nil
			}

			// Convert uid bytes to ShortID
			var userID zkidentity.ShortID
			copy(userID[:], uid)

			// Iterate through each tip in the user's sub-bucket
			return userBucket.ForEach(func(_, v []byte) error {
				// Decode the JSON-encoded `ReceivedTipWrapper` object
				var wrapper ReceivedTipWrapper
				if err := json.Unmarshal(v, &wrapper); err != nil {
					return err
				}

				// Only append tips with StatusUnprocessed
				if wrapper.Status == StatusUnprocessed {
					unprocessedTips[userID] = append(unprocessedTips[userID], wrapper.Tip)
				}
				return nil
			})
		})
	})
	if err != nil {
		return nil, err
	}

	return unprocessedTips, nil
}
