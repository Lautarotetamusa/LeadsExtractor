package whatsapp

import (
	"fmt"
	"testing"
)

const phone = "+5493415854220"

func TestGenerateLink(t *testing.T) {
    params := "s=rebora&m=ashe&c=test"
    link := GenerateWppLink("5213328092850", params, "Hola!")
    fmt.Println(link)
    res := "https://wa.me/5213328092850?text=%E2%80%8C%E2%80%8E%E2%80%8D%E2%80%8A%E2%80%8C%E2%80%8C%E2%80%8C%E2%80%8B%E2%80%8C%E2%80%8E%E2%80%8C%E2%80%8E%E2%80%8C%E2%80%8E%E2%80%8A%E2%80%8B%E2%80%8C%E2%80%8D%E2%80%8E%E2%80%8D%E2%80%8C%E2%80%8E%E2%80%8C%E2%80%8B%E2%80%8C%E2%80%8E%E2%80%8C%E2%80%8E%E2%80%8C%E2%80%8D%E2%80%8E%E2%80%8C%E2%80%8C%E2%80%8B%E2%80%8C%E2%80%8D%E2%80%8C%E2%80%8E%E2%80%8B%E2%80%8E%E2%80%8C%E2%80%8C%E2%80%8C%E2%80%8B%E2%80%8C%E2%80%8D%E2%80%8E%E2%80%8C%E2%80%8C%E2%80%8E%E2%80%8D%E2%80%8A%E2%80%8C%E2%80%8E%E2%80%8A%E2%80%8E%E2%80%8C%E2%80%8E%E2%80%8A%E2%80%8B%E2%80%8C%E2%80%8B%E2%80%8C%E2%80%8D%E2%80%8C%E2%80%8D%E2%80%8E%E2%80%8E%E2%80%8C%E2%80%8C%E2%80%8C%E2%80%8B%E2%80%8C%E2%80%8E%E2%80%8D%E2%80%8B%E2%80%8C%E2%80%8E%E2%80%8A%E2%80%8B%E2%80%8C%E2%80%8E%E2%80%8D%E2%80%8A%E2%80%8C%E2%80%8E%E2%80%8D%E2%80%8BHola!"
    if link != res {
        t.Error("no match")
    }
}

func NotCompleteUtm(t *testing.T) {
    params := "s=rebora&m=ashe"
    link := GenerateWppLink("5213328092850", params, "Hola!")
    fmt.Println(link)
}
