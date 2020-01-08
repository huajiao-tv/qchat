package idgen_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/huajiao-tv/qchat/utility/idgen"
)

func TestGenId(t *testing.T) {
	t.Log(time.Now().UnixNano()/1000000 - 86400*2*1000)
	t.Log(time.Unix(1388506077506/1000, 1388506077506%1000))
	version := 3
	shardId := 1000
	if (idgen.SetVersion(version) != nil) || (idgen.SetShardId(shardId) != nil) {
		t.Error("set version | set shardId failed")
	}
	id := idgen.GenId()
	if idgen.GetVersion(id) != 3 || idgen.GetShardId(id) != shardId {
		t.Error("id generation error", idgen.GetVersion(id), idgen.GetShardId(id))
	}
	hash := map[uint64]int{}
	st := time.Now()
	count := 250000
	for i := 0; i < count; i++ {
		id := idgen.GenId()
		if _, ok := hash[id]; ok {
			hash[id]++
		} else {
			hash[id] = 1
		}
		if i == 1000 {
			t.Log("id is", id)
		}
	}
	t.Log("gen id amount: "+strconv.Itoa(count)+" timecost: ", time.Now().Sub(st))
	for k, v := range hash {
		if v > 1 {
			t.Error("idgen reduplicate id ", k, v, idgen.GetSequence(k), idgen.GetTimeUnixNano(k))
			break
		}
	}
	t.Log("success")
}

var testId uint64 = 3459127098900473889

func TestGetVersion(t *testing.T) {
	version := idgen.GetVersion(testId)
	if version != 3 {
		t.Error("getVersion failed", version)
	}
	t.Log("get version success", version)
}

func TestGetShardId(t *testing.T) {
	shardId := idgen.GetShardId(testId)
	if shardId != 1000 {
		t.Error("get shardid failed")
	}
	t.Log("get shardId success", shardId)
}

func TestGetTime(t *testing.T) {
	time := idgen.GetTime(testId)
	if time.UnixNano() != 1388678971545000000 {
		t.Error("get time failed", time)
	}
	t.Log("get time success", time)
}

func TestGetTimeUnixNano(t *testing.T) {
	timeUnixNano := idgen.GetTimeUnixNano(testId)
	if timeUnixNano != 1388678971545000000 {
		t.Error("get Time Unix Nano failed", timeUnixNano)
	}
	t.Log("get time unix success", timeUnixNano)
}

func TestGetSequence(t *testing.T) {
	sequence := idgen.GetSequence(testId)
	if sequence != 33 {
		t.Error("get sequence error", sequence)
	}
}
