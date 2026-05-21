package email

import "context"

type Sender interface {
	// Send(to []string, subject, body string) error
	Send(ctx context.Context, fromAddress string, to []string, subject, body string, isHTML bool) error
}
