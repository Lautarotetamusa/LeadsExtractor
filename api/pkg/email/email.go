package email

import "context"

type Sender interface {
	Send(ctx context.Context, to []string, subject, body string, isHTML bool) error
}
