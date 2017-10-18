package rabbitroutine

import (
	"context"
	"testing"
	"time"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
)

var testCfg = Config{
	Host:     "127.0.0.1",
	Port:     5672,
	Username: "guest",
	Password: "guest",
	// Max reconnect attempts
	Attempts: 20,
	// How long wait between reconnect
	Wait: 5 * time.Second,
}

func TestContextDoneIsCorrectAndNotBlocking(t *testing.T) {
	defer time.AfterFunc(1*time.Second, func() { panic("contextDone deadlock") }).Stop()

	tests := []struct {
		ctx      func() context.Context
		expected bool
	}{
		{
			func() context.Context {
				return context.Background()
			},
			false,
		},
		{
			func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			true,
		},
		{
			// nolint: vet
			func() context.Context {
				ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
				return ctx
			},
			false,
		},
	}

	for _, test := range tests {
		actual := contextDone(test.ctx())

		assert.Equal(t, test.expected, actual)
	}
}

func TestDoNotBlocking(t *testing.T) {
	defer time.AfterFunc(1*time.Second, func() { panic("Do deadlock") }).Stop()

	Do(context.Background(), testCfg)
}

func TestDoWaitIsBlocking(t *testing.T) {
	conn := Do(context.Background(), testCfg)

	// nolint: errcheck
	go func() {
		conn.Wait()

		panic("Wait is not blocking")
	}()

	<-time.After(5 * time.Millisecond)
}

func TestStartIsBlocking(t *testing.T) {
	c := &Connector{
		cfg:    testCfg,
		connCh: make(chan *amqp.Connection),
	}

	// nolint: errcheck
	go func() {
		c.start(context.Background())

		panic("start is not blocking")
	}()

	<-time.After(5 * time.Millisecond)
}
