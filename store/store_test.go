package store

import (
	"crypto/rand"
	"fmt"
	"testing"

	"code.aliyun.com/chain33/chain33/common"
	"code.aliyun.com/chain33/chain33/common/config"
	"code.aliyun.com/chain33/chain33/queue"
	"code.aliyun.com/chain33/chain33/types"
)

func init() {
	common.SetLogLevel("debug")
}

func initEnv() (*queue.Queue, *Store) {
	var q = queue.New("channel")
	cfg := config.InitCfg("chain33.toml")
	s := New(cfg.Store)
	s.SetQueueClient(q.NewClient())
	return q, s
}

func set(qclient queue.Client, hash, key, value []byte) ([]byte, error) {
	kv := &types.KeyValue{key, value}
	set := &types.StoreSet{}
	set.StateHash = hash
	set.KV = append(set.KV, kv)

	msg := qclient.NewMessage("store", types.EventStoreSet, set)
	qclient.Send(msg, true)
	msg, err := qclient.Wait(msg)
	if err != nil {
		return nil, err
	}
	return msg.GetData().(*types.ReplyHash).GetHash(), nil
}

func setmem(qclient queue.Client, hash, key, value []byte) ([]byte, error) {
	kv := &types.KeyValue{key, value}
	set := &types.StoreSet{}
	set.StateHash = hash
	set.KV = append(set.KV, kv)

	msg := qclient.NewMessage("store", types.EventStoreMemSet, set)
	qclient.Send(msg, true)
	msg, err := qclient.Wait(msg)
	if err != nil {
		return nil, err
	}
	return msg.GetData().(*types.ReplyHash).GetHash(), nil
}

func get(qclient queue.Client, hash, key []byte) ([]byte, error) {
	query := &types.StoreGet{hash, [][]byte{key}}
	msg := qclient.NewMessage("store", types.EventStoreGet, query)
	qclient.Send(msg, true)
	msg, err := qclient.Wait(msg)
	if err != nil {
		return nil, err
	}
	values := msg.GetData().(*types.StoreReplyValue).GetValues()
	return values[0], nil
}

func commit(qclient queue.Client, hash []byte) ([]byte, error) {
	req := &types.ReqHash{hash}
	msg := qclient.NewMessage("store", types.EventStoreCommit, req)
	qclient.Send(msg, true)
	msg, err := qclient.Wait(msg)
	if err != nil {
		return nil, err
	}
	hash = msg.GetData().(*types.ReplyHash).GetHash()
	return hash, nil
}

func rollback(qclient queue.Client, hash []byte) ([]byte, error) {
	req := &types.ReqHash{hash}
	msg := qclient.NewMessage("store", types.EventStoreRollback, req)
	qclient.Send(msg, true)
	msg, err := qclient.Wait(msg)
	if err != nil {
		return nil, err
	}
	hash = msg.GetData().(*types.ReplyHash).GetHash()
	return hash, nil
}

func TestGetAndSet(t *testing.T) {
	q, s := initEnv()
	qclient := q.NewClient()
	var stateHash [32]byte
	//先set一个数
	key := []byte("hello")
	value := []byte("world")

	hash, err := set(qclient, stateHash[:], key, value)
	if err != nil {
		t.Error(err)
		return
	}

	value2, err := get(qclient, hash, key)
	if err != nil {
		t.Error(err)
		return
	}
	if string(value2) != string(value) {
		t.Errorf("values not match")
		return
	}
	s.Close()
}

func randstr() string {
	var hash [16]byte
	_, err := rand.Read(hash[:])
	if err != nil {
		panic(err)
	}
	return common.ToHex(hash[:])
}

func TestGetAndSetCommitAndRollback(t *testing.T) {
	q, s := initEnv()
	qclient := q.NewClient()
	var stateHash [32]byte
	//先set一个数
	key := []byte("hello" + randstr())
	value := []byte("world")

	hash, err := setmem(qclient, stateHash[:], key, value)
	if err != nil {
		t.Error(err)
		return
	}

	value2, err := get(qclient, hash, key)
	if err != nil {
		t.Error(err)
		return
	}
	if string(value2) != string(value) {
		t.Errorf("values not match %s %s %x", string(value2), string(value), hash)
		return
	}

	rollback(qclient, hash)
	value2, err = get(qclient, hash, key)
	if err != nil {
		t.Error(err)
		return
	}
	if len(value2) != 0 {
		t.Error(err)
		return
	}

	hash, err = setmem(qclient, stateHash[:], key, value)
	if err != nil {
		t.Error(err)
		return
	}

	commit(qclient, hash)

	value2, err = get(qclient, hash, key)
	if err != nil {
		t.Error(err)
		return
	}
	if string(value2) != string(value) {
		t.Errorf("values not match [%s] [%s] %x", string(value2), string(value), hash)
		return
	}

	s.Close()
}

func BenchmarkGetKey(b *testing.B) {
	q, s := initEnv()
	qclient := q.NewClient()
	var stateHash [32]byte
	hash := stateHash[:]
	var err error
	for i := 0; i < 1000; i++ {
		key := []byte(fmt.Sprintf("%020d", i))
		value := []byte(fmt.Sprintf("%020d", i))
		hash, err = set(qclient, hash, key, value)
		if err != nil {
			b.Error(err)
			return
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := []byte(fmt.Sprintf("%020d", i%1000))
		value := fmt.Sprintf("%020d", i%1000)
		value2, err := get(qclient, hash, key)
		if err != nil {
			b.Error(err)
			return
		}
		if string(value2) != value {
			b.Error(err)
			return
		}
	}
	s.Close()
}

func BenchmarkSetKeyOneByOne(b *testing.B) {
	q, s := initEnv()
	qclient := q.NewClient()
	var stateHash [32]byte
	hash := stateHash[:]
	var err error
	for i := 0; i < b.N; i++ {
		key := []byte(fmt.Sprintf("%020d", i))
		value := []byte(fmt.Sprintf("%020d", i))
		hash, err = set(qclient, hash, key, value)
		if err != nil {
			b.Error(err)
			return
		}
	}
	s.Close()
}

func BenchmarkSetKey1000(b *testing.B) {
	q, s := initEnv()
	qclient := q.NewClient()
	var stateHash [32]byte
	hash := stateHash[:]
	set := &types.StoreSet{}

	for i := 0; i < b.N; i++ {
		key := []byte(fmt.Sprintf("%020d", i))
		value := []byte(fmt.Sprintf("%020d", i))
		kv := &types.KeyValue{key, value}
		if i%1000 == 0 {
			set = &types.StoreSet{}
			set.StateHash = hash
		}
		set.KV = append(set.KV, kv)

		if i > 0 && i%1000 == 0 {
			msg := qclient.NewMessage("store", types.EventStoreSet, set)
			qclient.Send(msg, true)
			msg, err := qclient.Wait(msg)
			if err != nil {
				b.Error(err)
				return
			}
			hash = msg.GetData().(*types.ReplyHash).GetHash()
		}
	}
	s.Close()
}
