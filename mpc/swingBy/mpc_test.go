/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package swingBy

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/binance-chain/tss-lib/tss"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func (parties parties) init(senders []Sender, threshold int) {
	for i, p := range parties {
		p.Init(parties.numericIDs(), threshold, senders[i])
	}
}

func (parties parties) setShareData(shareData [][]byte) {
	for i, p := range parties {
		p.SetShareData(shareData[i])
	}
}

func (parties parties) sign(msg []byte) ([][]byte, error) {
	var lock sync.Mutex
	var sigs [][]byte
	var threadSafeError atomic.Value

	var wg sync.WaitGroup
	wg.Add(len(parties))

	for _, p := range parties {
		go func(p *party) {
			defer wg.Done()
			sig, err := p.Sign(context.Background(), msg)
			if err != nil {
				threadSafeError.Store(err.Error())
				return
			}

			lock.Lock()
			sigs = append(sigs, sig)
			lock.Unlock()
		}(p)
	}

	wg.Wait()

	err := threadSafeError.Load()
	if err != nil {
		return nil, fmt.Errorf(err.(string))
	}

	return sigs, nil
}

func (parties parties) keygen() ([][]byte, error) {
	var lock sync.Mutex
	shares := make([][]byte, len(parties))
	var threadSafeError atomic.Value

	var wg sync.WaitGroup
	wg.Add(len(parties))

	for i, p := range parties {
		go func(p *party, i int) {
			defer wg.Done()
			share, err := p.KeyGen(context.Background())
			if err != nil {
				threadSafeError.Store(err.Error())
				return
			}

			lock.Lock()
			shares[i] = share
			lock.Unlock()
		}(p, i)
	}

	wg.Wait()

	err := threadSafeError.Load()
	if err != nil {
		return nil, fmt.Errorf(err.(string))
	}

	return shares, nil
}

func (parties parties) Mapping() map[string]*tss.PartyID {
	partyIDMap := make(map[string]*tss.PartyID)
	for _, id := range parties {
		partyIDMap[id.id.Id] = id.id
	}
	return partyIDMap
}

/*func TestTSS(t *testing.T) {
	pA := NewParty(1, logger("pA", t.Name()))
	pB := NewParty(2, logger("pB", t.Name()))
	pC := NewParty(3, logger("pC", t.Name()))

	t.Logf("Created parties")

	parties := parties{pA, pB, pC}
	parties.init(senders(parties))

	t.Logf("Running DKG")

	t1 := time.Now()
	shares, err := parties.keygen()
	assert.NoError(t, err)
	t.Logf("DKG elapsed %s", time.Since(t1))

	parties.init(senders(parties))

	parties.setShareData(shares)
	t.Logf("Signing")

	msgToSign := []byte("bla bla")

	t.Logf("Signing message")
	t1 = time.Now()
	sigs, err := parties.sign(digest(msgToSign))
	assert.NoError(t, err)
	t.Logf("Signing completed in %v", time.Since(t1))

	sigSet := make(map[string]struct{})
	for _, s := range sigs {
		sigSet[string(s)] = struct{}{}
	}
	assert.Len(t, sigSet, 1)

	pk, err := parties[0].TPubKey()
	assert.NoError(t, err)

	assert.True(t, ecdsa.VerifyASN1(pk, digest(msgToSign), sigs[0]))
}*/

func senders(parties parties) []Sender {
	var senders []Sender
	for _, src := range parties {
		src := src
		sender := func(msgBytes []byte, broadcast bool, to uint16) {
			messageSource := uint16(big.NewInt(0).SetBytes(src.id.Key).Uint64())
			if broadcast {
				for _, dst := range parties {
					if dst.id == src.id {
						continue
					}
					dst.OnMsg(msgBytes, messageSource, broadcast)
				}
			} else {
				for _, dst := range parties {
					if to != uint16(big.NewInt(0).SetBytes(dst.id.Key).Uint64()) {
						continue
					}
					dst.OnMsg(msgBytes, messageSource, broadcast)
				}
			}
		}
		senders = append(senders, sender)
	}
	return senders
}

func logger(id string, testName string) Logger {
	logConfig := zap.NewDevelopmentConfig()
	logger, _ := logConfig.Build()
	logger = logger.With(zap.String("t", testName)).With(zap.String("id", id))
	return logger.Sugar()
}

/*func TestBenchmarkTss(t *testing.T) {
	pA := NewParty(1, logger("pA", t.Name()))
	pB := NewParty(2, logger("pB", t.Name()))
	pC := NewParty(3, logger("pC", t.Name()))
	pD := NewParty(4, logger("pD", t.Name()))
	pE := NewParty(5, logger("pE", t.Name()))
	pF := NewParty(6, logger("pF", t.Name()))
	pG := NewParty(7, logger("pG", t.Name()))
	pH := NewParty(8, logger("pH", t.Name()))
	pI := NewParty(9, logger("pI", t.Name()))
	pJ := NewParty(10, logger("pJ", t.Name()))
	pK := NewParty(11, logger("pK", t.Name()))
	pL := NewParty(12, logger("pL", t.Name()))
	pM := NewParty(13, logger("pM", t.Name()))
	pN := NewParty(14, logger("pN", t.Name()))
	pO := NewParty(15, logger("pO", t.Name()))
	pP := NewParty(16, logger("pP", t.Name()))
	pQ := NewParty(17, logger("pQ", t.Name()))
	pR := NewParty(18, logger("pR", t.Name()))
	pS := NewParty(19, logger("pS", t.Name()))
	pT := NewParty(20, logger("pT", t.Name()))

	threshold := 2

	t.Logf("Created parties")

	parties := parties{pA, pB, pC, pD, pE, pF, pG, pH, pI, pJ, pK, pL, pM, pN, pO, pP, pQ, pR, pS, pT}
	//parties := parties{pA, pB, pC, pD, pE, pF, pG, pH, pI}
	parties.init(senders(parties), threshold)

	t.Logf("Running DKG")

	t1 := time.Now()
	shares, err := parties.keygen()
	assert.NoError(t, err)
	t.Logf("DKG elapsed %s", time.Since(t1))

	parties.init(senders(parties), threshold)

	parties.setShareData(shares)
	t.Logf("Signing")

	msgToSign := []byte("bla bla")

	t.Logf("Signing message")
	t1 = time.Now()
	sigs, err := parties.sign(digest(msgToSign))
	assert.NoError(t, err)
	t.Logf("Signing completed in %v", time.Since(t1))

	sigSet := make(map[string]struct{})
	for _, s := range sigs {
		sigSet[string(s)] = struct{}{}
	}
	assert.Len(t, sigSet, 1)

	pk, err := parties[0].TPubKey()
	assert.NoError(t, err)

	assert.True(t, ecdsa.VerifyASN1(pk, digest(msgToSign), sigs[0]))
}*/

func TestBenchmarkTss(t *testing.T) {
	allParties := []*party{
		NewParty(1, logger("pA", t.Name())),
		NewParty(2, logger("pB", t.Name())),
		NewParty(3, logger("pC", t.Name())),
		NewParty(4, logger("pD", t.Name())),
		NewParty(5, logger("pE", t.Name())),
		NewParty(6, logger("pF", t.Name())),
		NewParty(7, logger("pG", t.Name())),
		NewParty(8, logger("pH", t.Name())),
		NewParty(9, logger("pI", t.Name())),
		NewParty(10, logger("pJ", t.Name())),
		NewParty(11, logger("pK", t.Name())),
		NewParty(12, logger("pL", t.Name())),
		NewParty(13, logger("pM", t.Name())),
		NewParty(14, logger("pN", t.Name())),
		NewParty(15, logger("pO", t.Name())),
		NewParty(16, logger("pP", t.Name())),
		NewParty(17, logger("pQ", t.Name())),
		NewParty(18, logger("pR", t.Name())),
		NewParty(19, logger("pS", t.Name())),
		NewParty(20, logger("pT", t.Name())),
	}

	benchmarks := []struct {
		threshold  int
		numParties int
	}{
		{2, 3}, /*, {2, 4}, {3, 4}, {2, 5}, {3, 5}, {4, 5},
		{2, 6}, {3, 6}, {4, 6}, {5, 6}, {2, 7}, {14, 20},*/
	}

	numRuns := 1

	for _, bm := range benchmarks {
		t.Run(fmt.Sprintf("Threshold:%d/Parties:%d", bm.threshold, bm.numParties), func(t *testing.T) {
			var totalDKGTime time.Duration
			var totalSigningTime time.Duration

			for i := 0; i < numRuns; i++ {
				parties := parties(allParties[:bm.numParties])
				threshold := bm.threshold

				parties.init(senders(parties), threshold)

				// DKG
				t1 := time.Now()
				shares, err := parties.keygen()
				assert.NoError(t, err)
				dkgTime := time.Since(t1)
				totalDKGTime += dkgTime

				parties.init(senders(parties), threshold)
				parties.setShareData(shares)

				// Signing
				msgToSign := []byte("bla bla")
				t1 = time.Now()
				sigs, err := parties.sign(digest(msgToSign))
				assert.NoError(t, err)
				signingTime := time.Since(t1)
				totalSigningTime += signingTime

				// Verification (only done once per benchmark for simplicity)
				if i == 0 {
					sigSet := make(map[string]struct{})
					for _, s := range sigs {
						sigSet[string(s)] = struct{}{}
					}
					assert.Len(t, sigSet, 1)

					pk, err := parties[0].TPubKey()
					assert.NoError(t, err)
					assert.True(t, ecdsa.VerifyASN1(pk, digest(msgToSign), sigs[0]))
				}
			}

			meanDKGTime := totalDKGTime / time.Duration(numRuns)
			meanSigningTime := totalSigningTime / time.Duration(numRuns)

			t.Logf("Parties: %d, Threshold: %d", bm.numParties, bm.threshold)
			t.Logf("Mean DKG time: %v", meanDKGTime)
			t.Logf("Mean Signing time: %v", meanSigningTime)
		})
	}
}
