package krpc

import (
	"testing"

	"github.com/sauromates/leech/dht/query"
	"github.com/stretchr/testify/assert"
)

func TestSerializeResponse(t *testing.T) {
	body := struct{ id string }{"mnopqrstuvwxyz123456"}
	expected := "d1:rd2:id20:mnopqrstuvwxyz123456e1:t2:aa1:y1:re"

	msg, err := Serialize(*NewResponseMessage(body))

	assert.Nil(t, err)
	assert.Equal(t, expected, string(msg))
}

func TestSerializeError(t *testing.T) {
	body := []any{201, "A Generic Error Ocurred"}
	expected := "d1:eli201e23:A Generic Error Ocurrede1:t2:aa1:y1:ee"

	msg, err := Serialize(*NewErrorMessage(body))

	assert.Nil(t, err)
	assert.Equal(t, expected, string(msg))
}

func TestSerializeQuery(t *testing.T) {
	type testCase struct {
		queryType  query.QueryType
		body       interface{}
		output     string
		shouldFail bool
	}

	tt := map[string]testCase{
		"ping query": {
			queryType:  query.Ping,
			body:       struct{ id string }{id: "abcdefghij0123456789"},
			output:     "d1:ad2:id20:abcdefghij0123456789e1:q4:ping1:t2:aa1:y1:qe",
			shouldFail: false,
		},
		"find node query": {
			queryType: query.FindNode,
			body: struct {
				id     string
				target string
			}{"abcdefghij0123456789", "mnopqrstuvwxyz123456"},
			output:     "d1:ad2:id20:abcdefghij01234567896:target20:mnopqrstuvwxyz123456e1:q9:find_node1:t2:aa1:y1:qe",
			shouldFail: false,
		},
		"get peers query": {
			queryType: query.GetPeers,
			body: struct {
				id        string
				info_hash string
			}{"abcdefghij0123456789", "mnopqrstuvwxyz123456"},
			output:     "d1:ad2:id20:abcdefghij01234567899:info_hash20:mnopqrstuvwxyz123456e1:q9:get_peers1:t2:aa1:y1:qe",
			shouldFail: false,
		},
		"announce peer query": {
			queryType: query.Announce,
			body: struct {
				id           string
				implied_port int
				info_hash    string
				port         int
				token        string
			}{"abcdefghij0123456789", 1, "mnopqrstuvwxyz123456", 6881, "aoeusnth"},
			output:     "d1:ad2:id20:abcdefghij012345678912:implied_porti1e9:info_hash20:mnopqrstuvwxyz1234564:porti6881e5:token8:aoeusnthe1:q13:announce_peer1:t2:aa1:y1:qe",
			shouldFail: false,
		},
		"unknown query": {
			queryType:  query.Unknown,
			body:       nil,
			output:     string([]byte{}),
			shouldFail: true,
		},
	}

	for name, tc := range tt {
		msg, err := NewQueryMessage(tc.queryType, tc.body)

		if tc.shouldFail {
			assert.NotNil(t, err, name)
		} else {
			serialized, err := Serialize(*msg)
			if err != nil {
				t.Error(err, name)
			}

			assert.Nil(t, err, name)
			assert.Equal(t, tc.output, string(serialized), name)
		}
	}
}
