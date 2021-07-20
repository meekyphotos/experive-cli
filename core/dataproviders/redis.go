package dataproviders

import (
	"encoding/binary"
	"math"

	"github.com/dgraph-io/badger/v3"
	"github.com/valyala/fastjson"
)

type Store struct {
	db   *badger.DB
	pool *fastjson.ParserPool
}

func (s *Store) Open(path string) {
	db, err := badger.Open(
		badger.DefaultOptions(path),
	)
	if err != nil {
		panic(err)
	}
	s.db = db
	s.pool = &fastjson.ParserPool{}
}
func (s *Store) Close() error {
	return s.db.Close()
}

type Mapper func(*fastjson.Value) map[string]interface{}

func (s *Store) Stream(mapper Mapper) chan map[string]interface{} {
	out := make(chan map[string]interface{}, 10000)
	go func() {
		err := s.db.View(func(txn *badger.Txn) error {
			itr := txn.NewIterator(badger.IteratorOptions{
				PrefetchValues: true,
				PrefetchSize:   10000,
				AllVersions:    false,
			})
			defer itr.Close()
			for itr.Rewind(); itr.Valid(); itr.Next() {
				item := itr.Item()
				item.Value(func(val []byte) error {
					parser := s.pool.Get()
					defer s.pool.Put(parser)

					json, err := parser.ParseBytes(val)
					if err != nil {
						return err
					}
					out <- mapper(json)
					return nil
				})
			}
			close(out)
			return nil
		})
		if err != nil {
			panic(err)
		}
	}()

	return out
}

func Uint64ToBytes(i uint64) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], i)
	return buf[:]
}

func Float64ToBytes(f float64) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], math.Float64bits(f))
	return buf[:]
}

func (s *Store) Save(node *INode) {
	err := s.db.Update(func(txn *badger.Txn) error {
		return saveInTransaction(txn, node)
	})
	if err != nil {
		panic(err)
	}
}

type INode struct {
	Id      int64
	Content []byte
}

func saveInTransaction(txn *badger.Txn, n *INode) error {
	err := txn.Set(Uint64ToBytes(uint64(n.Id)), n.Content)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) SaveMany(nodes ...*INode) {
	err := s.db.Update(func(txn *badger.Txn) error {
		for _, n := range nodes {
			err := saveInTransaction(txn, n)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func (s *Store) FindOne(id int64) *fastjson.Value {
	var json *fastjson.Value

	err := s.db.View(func(txn *badger.Txn) error {
		var err error
		json, err = s.readKey(txn, id)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		panic(err)
	}
	return json
}

func (s *Store) readKey(txn *badger.Txn, id int64) (*fastjson.Value, error) {
	var json *fastjson.Value
	item, err := txn.Get(Uint64ToBytes(uint64(id)))
	if err == badger.ErrKeyNotFound {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	if item != nil {
		err = item.Value(func(val []byte) error {
			parser := s.pool.Get()
			defer s.pool.Put(parser)

			json, err = parser.ParseBytes(val)
			if err != nil {
				return err
			}
			return nil
		})
		return json, err
	} else {
		return nil, nil
	}
}

func (s *Store) FindMany(ids ...int64) map[int64]*fastjson.Value {
	var out = make(map[int64]*fastjson.Value, len(ids))

	err := s.db.View(func(txn *badger.Txn) error {
		for _, id := range ids {
			json, err := s.readKey(txn, id)
			if err != nil {
				return err
			}
			if json != nil {
				out[id] = json
			}
		}
		return nil
	})

	if err != nil {
		panic(err)
	}
	return out
}
