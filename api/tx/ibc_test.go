package tx

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_getIBCSeqFromTx(t *testing.T) {
	// This test assumes tx_response.logs.events is always present,
	// because Tendermint events are formatted this way.
	tests := []struct {
		name    string
		payload string
		want    []string
	}{
		{
			"ibc send payload with a single transfer",
			`{
				"tx_response":{
				   "logs":[
					  {
						 "events":[
							{
							   "type":"send_packet",
							   "attributes":[
								  {
									 "key":"packet_sequence",
									 "value":"42"
								  }
							   ]
							}
						 ]
					  }
				   ]
				}
			 }`,
			[]string{"42"},
		},
		{
			"ibc send payload with multiple transfer",
			`{
				"tx_response":{
				   "logs":[
					  {
						 "events":[
							{
							   "type":"send_packet",
							   "attributes":[
								  {
									 "key":"packet_sequence",
									 "value":"42"
								  }
							   ]
							}
						 ]
					  },
					  {
						"events":[
						   {
							  "type":"send_packet",
							  "attributes":[
								 {
									"key":"packet_sequence",
									"value":"44"
								 }
							  ]
						   }
						]
					 }
				   ]
				}
			 }`,
			[]string{"42", "44"},
		},
		{
			"transaction with no IBC-related tx's",
			`{
				"tx_response":{
				   "logs":[
					  {
						 "events":[
							{
							   "type":"fake_event",
							   "attributes":[
								  {
									 "key":"packet_sequence",
									 "value":"42"
								  }
							   ]
							}
						 ]
					  }
				   ]
				}
			 }`,
			[]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := getIBCSeqFromTx([]byte(tt.payload))
			require.ElementsMatch(t, res, tt.want)
		})
	}
}
