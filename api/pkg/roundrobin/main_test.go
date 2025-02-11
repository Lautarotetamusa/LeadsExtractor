package roundrobin_test

import (
	"leadsextractor/pkg/roundrobin"
	"testing"

	"github.com/stretchr/testify/assert"
)

var rr *roundrobin.RoundRobin[m]

type m struct {
    f string
}

func M(f string) *m{
    return &m{
        f: f,
    }
}

func TestMain(t *testing.M) {
    parcipants := []*m{M("AA"), M("BB"), M("CC"), M("DD")}
    rr = roundrobin.New(parcipants)

    t.Run()
}

func TestNext(t *testing.T) {
    bb := rr.Next()
    assert.Equal(t, bb.f, "BB")
    cc := rr.Next()
    assert.Equal(t, cc.f, "CC")
    dd := rr.Next()
    assert.Equal(t, dd.f, "DD")
    aa := rr.Next()
    assert.Equal(t, aa.f, "AA")
}

func TestRestart(t *testing.T) {
    rr.Next()
    rr.Restart()
    bb := rr.Next()
    assert.Equal(t, bb.f, "BB")
}

func TestReasign(t *testing.T) {
    parcipants := []*m{M("aa"), M("bb"), M("cc"), M("dd")}
    rr.Reasign(parcipants)
    bb := rr.Next()
    assert.Equal(t, bb.f, "bb")
}

func TestContains(t *testing.T) {
    assert.True(t, rr.Contains(M("aa")))
    assert.False(t, rr.Contains(M("ee")))
}

func TestAdd(t *testing.T) {
    rr.Add(M("EE"))
    assert.True(t, rr.Contains(M("EE")))
}
