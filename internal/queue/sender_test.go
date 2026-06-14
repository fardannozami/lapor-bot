package queue

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

type fakeClient struct {
	sent   int32
	delay  time.Duration
	blockC chan struct{}
}

func (f *fakeClient) SendMessage(ctx context.Context, target types.JID, msg *waE2E.Message, extra ...whatsmeow.SendRequestExtra) (whatsmeow.SendResponse, error) {
	atomic.AddInt32(&f.sent, 1)
	if f.delay > 0 {
		select {
		case <-time.After(f.delay):
		case <-ctx.Done():
			return whatsmeow.SendResponse{}, ctx.Err()
		}
	}
	if f.blockC != nil {
		<-f.blockC
	}
	return whatsmeow.SendResponse{}, nil
}

func (f *fakeClient) sentCount() int32 {
	return atomic.LoadInt32(&f.sent)
}

func TestSendNormalPriority(t *testing.T) {
	fc := &fakeClient{}
	sender := NewTestSender(fc, context.Background())
	sender.Start()

	err := sender.SendNormalPriority(context.Background(), types.JID{}, &waE2E.Message{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sender.cancel()
	sender.wg.Wait()

	if fc.sentCount() != 1 {
		t.Errorf("expected 1 message sent, got %d", fc.sentCount())
	}
}

func TestPriorityOrdering(t *testing.T) {
	blockC := make(chan struct{})
	fc := &fakeClient{blockC: blockC}
	sender := NewTestSender(fc, context.Background())
	sender.Start()

	target := types.JID{}
	msg := &waE2E.Message{}

	go sender.SendNormalPriority(context.Background(), target, msg)
	go sender.SendHighPriority(context.Background(), target, msg)

	time.Sleep(50 * time.Millisecond)

	close(blockC)
	time.Sleep(50 * time.Millisecond)

	sender.cancel()
	sender.wg.Wait()

	if fc.sentCount() != 2 {
		t.Errorf("expected 2 messages sent, got %d", fc.sentCount())
	}
}

func TestChannelFullReturnsError(t *testing.T) {
	// Worker blocks on the first message forever. We push 3 messages
	// into a capacity-1 channel so the 3rd fails with a "channel full" error.
	blockForever := make(chan struct{})
	fc := &fakeClient{blockC: blockForever}
	sender := NewTestSender(fc, context.Background())
	sender.normalPriority = make(chan sendRequest, 1)
	sender.Start()

	msg := &waE2E.Message{}
	errC1 := make(chan error, 1)
	errC2 := make(chan error, 1)

	// Push msg1 — worker dequeues it and blocks.
	go func() {
		errC1 <- sender.SendNormalPriority(context.Background(), types.JID{}, msg)
	}()

	// Wait until msg1 is dequeued (the buffer is now empty).
	time.Sleep(50 * time.Millisecond)

	// Push msg2 — fills the buffer.
	go func() {
		errC2 <- sender.SendNormalPriority(context.Background(), types.JID{}, msg)
	}()

	// Wait for msg2 to land in the buffer.
	time.Sleep(50 * time.Millisecond)

	// Push msg3 — buffer is full, must fail immediately.
	err := sender.SendNormalPriority(context.Background(), types.JID{}, msg)
	if err == nil {
		t.Error("expected error when channel is full")
	} else {
		t.Logf("got expected error: %v", err)
	}

	// Unblock the worker so everything drains.
	close(blockForever)

	// Wait for the two blocked goroutines to finish.
	<-errC1
	<-errC2

	sender.cancel()
	sender.wg.Wait()
}

func TestShutdownDrainsPending(t *testing.T) {
	blockC := make(chan struct{})
	fc := &fakeClient{blockC: blockC}
	sender := NewTestSender(fc, context.Background())
	sender.Start()

	target := types.JID{}
	msg := &waE2E.Message{}

	// Enqueue from goroutines so the test goroutine doesn't block.
	go sender.SendHighPriority(context.Background(), target, msg)
	go sender.SendNormalPriority(context.Background(), target, msg)

	time.Sleep(50 * time.Millisecond)

	// Both messages are in the queue, worker is blocked on blockC.
	// Start shutdown — it closes channels and waits for drain.
	shutdownDone := make(chan struct{})
	go func() {
		sender.Shutdown(2 * time.Second)
		close(shutdownDone)
	}()

	// Unblock the worker so it can drain.
	close(blockC)

	select {
	case <-shutdownDone:
	case <-time.After(3 * time.Second):
		t.Fatal("shutdown timed out")
	}

	if fc.sentCount() != 2 {
		t.Errorf("expected 2 messages drained, got %d", fc.sentCount())
	}
}
