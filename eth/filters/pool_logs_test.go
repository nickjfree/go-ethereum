package filters

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

var (
	bancorPool      = common.HexToAddress("0xe331821bc94187c2649e932810a60204699d45cb")
	bancorPoolToken = common.HexToAddress("0xb1cd6e4153b2a390cf00a6556b0fc1458c4a5533")
	bancorNative    = common.HexToAddress("0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE")
	bnt             = common.HexToAddress("0x1F573D6Fb3F13d689FF844B4cE37794d79a7FF1C")
)

func rateLog(token1, token2 common.Address, index uint) *types.Log {
	return &types.Log{
		Address:     bancorPool,
		Topics:      []common.Hash{BancorTokenRateUpdateTopic, common.BytesToHash(token1.Bytes()), common.BytesToHash(token2.Bytes())},
		BlockNumber: 26356517,
		Index:       index,
	}
}

func TestBancorTopic(t *testing.T) {
	want := common.HexToHash("0x77f29993cf2c084e726f7e802da0719d6a0ade3e204badc7a3ffd57ecb768c24")
	if BancorTokenRateUpdateTopic != want {
		t.Fatalf("topic %x, want %x", BancorTokenRateUpdateTopic, want)
	}
}

// The rate logs of tx 0xebd597d4d489a5fcbb66ab7ac3aef2478e926155c7c759f737c27156184dba2c (0.3 ETH to BNT): the pair of
// reserves first, then one per reserve and pool token. All three pass the pending feed's filter, and poolLogs keeps
// each pair's latest, so the pair of reserves is not lost to the pool token's logs.
func TestPoolLogsKeepEachBancorPair(t *testing.T) {
	logs := []*types.Log{
		rateLog(bancorNative, bnt, 809),
		rateLog(bancorPoolToken, bancorNative, 810),
		rateLog(bancorPoolToken, bnt, 811),
	}
	if got := filterRelevantLogs(logs); len(got) != 3 {
		t.Fatalf("pending filter kept %d of 3 rate logs", len(got))
	}
	if got := latestPoolLogs(logs); len(got) != 3 {
		t.Fatalf("poolLogs kept %d of 3 rate logs", len(got))
	}

	// a later trade the other way round names the pair the other way round: it replaces the first pair log
	back := rateLog(bnt, bancorNative, 900)
	got := latestPoolLogs(append(logs, back))
	if len(got) != 3 {
		t.Fatalf("poolLogs kept %d logs, want 3", len(got))
	}
	for _, lg := range got {
		if lg.Index == 809 {
			t.Fatal("the older pair log survived the later one")
		}
	}
}

func TestPoolLogsKeepTheLatestSync(t *testing.T) {
	pair := common.HexToAddress("0x0d4a11d5EEaaC28EC3F61d100daF4d40471f1852")
	sync := func(index uint) *types.Log {
		return &types.Log{Address: pair, Topics: []common.Hash{UniswapV2SyncTopic}, BlockNumber: 1, Index: index}
	}
	got := latestPoolLogs([]*types.Log{sync(5), sync(2)})
	if len(got) != 1 || got[0].Index != 5 {
		t.Fatalf("got %d logs, want the one at index 5", len(got))
	}
}
