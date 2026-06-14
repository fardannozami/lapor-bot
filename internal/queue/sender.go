// Package queue provides a priority-based message sender that serializes
// all SendMessage calls to the whatsmeow client through a single goroutine.
// Scheduled notifications go to a high-priority channel; user reply messages
// go to a normal-priority channel. The worker goroutine always checks the
// high-priority channel first, preventing user activity from delaying
// time-sensitive notifications.
package queue

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

const channelCapacity = 64

type sendRequest struct {
	ctx     context.Context
	target  types.JID
	msg     *waE2E.Message
	errChan chan error
}

type messageClient interface {
	SendMessage(ctx context.Context, to types.JID, msg *waE2E.Message, extra ...whatsmeow.SendRequestExtra) (whatsmeow.SendResponse, error)
}

// MessageSender serializes outgoing WhatsApp messages through a
// priority-aware goroutine so that scheduled notifications are never
// blocked by concurrent user command replies.
type MessageSender struct {
	client         messageClient
	highPriority   chan sendRequest
	normalPriority chan sendRequest
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
}

// NewMessageSender creates a sender for the given client. The sender
// must be started with Start() before use and stopped with Shutdown().
func NewMessageSender(client *whatsmeow.Client, parentCtx context.Context) *MessageSender {
	ctx, cancel := context.WithCancel(parentCtx)
	return &MessageSender{
		client:         client,
		highPriority:   make(chan sendRequest, channelCapacity),
		normalPriority: make(chan sendRequest, channelCapacity),
		ctx:            ctx,
		cancel:         cancel,
	}
}

// Start launches the worker goroutine that drains both priority channels.
func (s *MessageSender) Start() {
	s.wg.Add(1)
	go s.loop()
}

// Shutdown closes the send channels and waits for the worker to drain
// remaining high-priority messages up to the given timeout, then cancels
// the worker context. All remaining messages (including normal-priority)
// are dropped after the timeout.
func (s *MessageSender) Shutdown(timeout time.Duration) {
	close(s.highPriority)
	close(s.normalPriority)

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("[QUEUE] drained all pending messages")
	case <-time.After(timeout):
		log.Println("[QUEUE] drain timeout reached, forcing shutdown")
		s.cancel()
		s.wg.Wait()
	}
}

// SendHighPriority enqueues a message for immediate delivery ahead of
// normal-priority messages. Used for scheduled notifications.
// Returns an error if the queue is full (should not happen with 64 capacity).
func (s *MessageSender) SendHighPriority(ctx context.Context, target types.JID, msg *waE2E.Message) error {
	return s.send(s.highPriority, ctx, target, msg)
}

// SendNormalPriority enqueues a message at normal priority. Used for
// user command replies.
func (s *MessageSender) SendNormalPriority(ctx context.Context, target types.JID, msg *waE2E.Message) error {
	return s.send(s.normalPriority, ctx, target, msg)
}

func (s *MessageSender) send(ch chan sendRequest, ctx context.Context, target types.JID, msg *waE2E.Message) error {
	errChan := make(chan error, 1)
	req := sendRequest{ctx: ctx, target: target, msg: msg, errChan: errChan}

	select {
	case ch <- req:
		select {
		case err := <-errChan:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	case <-s.ctx.Done():
		return fmt.Errorf("sender is shutting down")
	default:
		return fmt.Errorf("send queue is full")
	}
}

func (s *MessageSender) loop() {
	defer s.wg.Done()

	for {
		select {
		case req, ok := <-s.highPriority:
			if !ok {
				// highPriority closed, drain only normalPriority
				s.drainNormal()
				return
			}
			s.doSend(req)
		default:
			select {
			case req, ok := <-s.highPriority:
				if !ok {
					s.drainNormal()
					return
				}
				s.doSend(req)
			case req, ok := <-s.normalPriority:
				if !ok {
					// Drain remaining high-priority before exiting
					s.drainHigh()
					return
				}
				s.doSend(req)
			case <-s.ctx.Done():
				log.Println("[QUEUE] worker cancelled")
				return
			}
		}
	}
}

func (s *MessageSender) drainNormal() {
	for req := range s.normalPriority {
		s.doSend(req)
	}
}

func (s *MessageSender) drainHigh() {
	for req := range s.highPriority {
		s.doSend(req)
	}
}

func (s *MessageSender) doSend(req sendRequest) {
	_, err := s.client.SendMessage(req.ctx, req.target, req.msg)
	select {
	case req.errChan <- err:
	default:
		if err != nil {
			log.Printf("[QUEUE] send error: %v", err)
		}
	}
}

// newTestSender creates a MessageSender with a custom client for testing.
func newTestSender(client messageClient, parentCtx context.Context) *MessageSender {
	ctx, cancel := context.WithCancel(parentCtx)
	return &MessageSender{
		client:         client,
		highPriority:   make(chan sendRequest, channelCapacity),
		normalPriority: make(chan sendRequest, channelCapacity),
		ctx:            ctx,
		cancel:         cancel,
	}
}
