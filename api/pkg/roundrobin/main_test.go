package roundrobin_test

import (
	"leadsextractor/pkg/roundrobin"
	"testing"

	"github.com/stretchr/testify/assert"
)

var rr *roundrobin.RoundRobin[string]

func TestMain(t *testing.M) {
    parcipants := []string{"AA", "BB", "CC", "DD"}
    rr = roundrobin.New(parcipants)

    t.Run()
}

func TestNext(t *testing.T) {
    bb := rr.Next()
    assert.Equal(t, bb, "BB")
    cc := rr.Next()
    assert.Equal(t, cc, "CC")
    dd := rr.Next()
    assert.Equal(t, dd, "DD")
    aa := rr.Next()
    assert.Equal(t, aa, "AA")
}

func TestRestart(t *testing.T) {
    rr.Next()
    rr.Restart()
    bb := rr.Next()
    assert.Equal(t, bb, "BB")
}

func TestReasign(t *testing.T) {
    parcipants := []string{"aa", "bb", "cc", "dd"}
    rr.Reasign(parcipants)
    bb := rr.Next()
    assert.Equal(t, bb, "bb")
}
