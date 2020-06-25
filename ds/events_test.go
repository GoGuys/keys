package ds_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/keys-pub/keys/ds"
	"github.com/keys-pub/keys/tsutil"
	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v4"
)

func TestMemEventLog(t *testing.T) {
	var err error

	// keys.SetLogger(keys.NewLogger(keys.DebugLevel))
	eds := ds.NewMem()
	clock := tsutil.NewClock()
	eds.SetTimeNow(clock.Now)

	ctx := context.TODO()
	path := ds.Path("test", "eds")

	length := 40
	values := [][]byte{}
	strs := []string{}
	for i := 0; i < length; i++ {
		str := fmt.Sprintf("value%d", i)
		values = append(values, []byte(str))
		strs = append(strs, str)
	}
	out, err := eds.EventsAdd(ctx, path, values)
	require.NoError(t, err)
	require.Equal(t, 40, len(out))
	for i, event := range out {
		require.False(t, event.Timestamp.IsZero())
		require.Equal(t, int64(i+1), event.Index)
	}

	// Events (limit=10, asc)
	iter, err := eds.Events(ctx, path, 0, 10, ds.Ascending)
	require.NoError(t, err)
	events, index, err := ds.EventsFromIterator(iter, 0)
	require.NoError(t, err)
	iter.Release()
	require.Equal(t, 10, len(events))
	eventsValues := []string{}
	for i, event := range events {
		require.False(t, event.Timestamp.IsZero())
		require.Equal(t, int64(i+1), event.Index)
		eventsValues = append(eventsValues, string(event.Data))
	}
	require.Equal(t, strs[0:10], eventsValues)

	// Events (index, asc)
	iter, err = eds.Events(ctx, path, index, 10, ds.Ascending)
	require.NoError(t, err)
	events, index, err = ds.EventsFromIterator(iter, index)
	require.NoError(t, err)
	iter.Release()
	require.Equal(t, int64(20), index)
	require.Equal(t, 10, len(events))
	eventsValues = []string{}
	for _, event := range events {
		eventsValues = append(eventsValues, string(event.Data))
	}
	require.Equal(t, strs[10:20], eventsValues)

	// Events (large index)
	large := int64(1000000000)
	iter, err = eds.Events(ctx, path, large, 100, ds.Ascending)
	require.NoError(t, err)
	events, index, err = ds.EventsFromIterator(iter, large)
	require.NoError(t, err)
	iter.Release()
	require.Equal(t, 0, len(events))
	require.Equal(t, large, index)

	// Descending
	revs := reverseCopy(strs)

	// Events (limit=10, desc)
	iter, err = eds.Events(ctx, path, 0, 10, ds.Descending)
	require.NoError(t, err)
	events, index, err = ds.EventsFromIterator(iter, 0)
	require.NoError(t, err)
	iter.Release()
	require.Equal(t, 10, len(events))
	require.Equal(t, int64(31), index)
	eventsValues = []string{}
	for _, event := range events {
		eventsValues = append(eventsValues, string(event.Data))
	}
	require.Equal(t, revs[0:10], eventsValues)

	// Events (limit=5, index, desc)
	iter, err = eds.Events(ctx, path, index, 5, ds.Descending)
	require.NoError(t, err)
	events, index, err = ds.EventsFromIterator(iter, index)
	require.NoError(t, err)
	iter.Release()
	require.Equal(t, 5, len(events))
	require.Equal(t, int64(26), index)
	eventsValues = []string{}
	for _, event := range events {
		eventsValues = append(eventsValues, string(event.Data))
	}
	require.Equal(t, revs[10:15], eventsValues)
}

func reverseCopy(s []string) []string {
	a := make([]string, len(s))
	for i, j := 0, len(s)-1; i < len(s); i++ {
		a[i] = s[j]
		j--
	}
	return a
}

func TestEventMarshal(t *testing.T) {
	clock := tsutil.NewClock()
	event := ds.Event{
		Data:      []byte{0x01, 0x02, 0x03},
		Index:     123,
		Timestamp: clock.Now(),
	}
	out, err := msgpack.Marshal(event)
	require.NoError(t, err)
	expected := `([]uint8) (len=36 cap=64) {
 00000000  83 a3 64 61 74 c4 03 01  02 03 a3 69 64 78 d3 00  |..dat......idx..|
 00000010  00 00 00 00 00 00 7b a2  74 73 d7 ff 00 3d 09 00  |......{.ts...=..|
 00000020  49 96 02 d2                                       |I...|
}
`
	require.Equal(t, expected, spew.Sdump(out))

	out, err = json.Marshal(event)
	require.NoError(t, err)
	expected = `([]uint8) (len=57 cap=64) {
 00000000  7b 22 64 61 74 61 22 3a  22 41 51 49 44 22 2c 22  |{"data":"AQID","|
 00000010  69 64 78 22 3a 31 32 33  2c 22 74 73 22 3a 22 32  |idx":123,"ts":"2|
 00000020  30 30 39 2d 30 32 2d 31  33 54 32 33 3a 33 31 3a  |009-02-13T23:31:|
 00000030  33 30 2e 30 30 31 5a 22  7d                       |30.001Z"}|
}
`
	require.Equal(t, expected, spew.Sdump(out))
}
