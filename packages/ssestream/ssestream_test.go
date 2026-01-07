package ssestream_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Nordlys-Labs/anthropic-sdk-go/packages/ssestream"
)

type mockDecoder struct {
	events []ssestream.Event
	index  int
	err    error
}

func (m *mockDecoder) Next() bool {
	if m.err != nil {
		return false
	}
	if m.index >= len(m.events) {
		return false
	}
	m.index++
	return true
}

func (m *mockDecoder) Event() ssestream.Event {
	if m.index == 0 {
		return ssestream.Event{}
	}
	return m.events[m.index-1]
}

func (m *mockDecoder) Close() error { return nil }

func (m *mockDecoder) Err() error { return m.err }

type testStruct struct {
	ID   string `json:"id"`
	Data string `json:"data"`
}

func TestStreamNormalOperation(t *testing.T) {
	events := []ssestream.Event{
		{Type: "message_start", Data: []byte(`{"id":"1","data":"test1"}`)},
		{Type: "message_delta", Data: []byte(`{"id":"2","data":"test2"}`)},
		{Type: "content_block_delta", Data: []byte(`{"id":"3","data":"test3"}`)},
	}
	decoder := &mockDecoder{events: events}
	stream := ssestream.NewStream[testStruct](decoder, nil)

	var ids []string
	for stream.Next() {
		ids = append(ids, stream.Current().ID)
	}

	expected := []string{"1", "2", "3"}
	if len(ids) != len(expected) {
		t.Fatalf("expected %d items, got %d", len(expected), len(ids))
	}
	for i := range expected {
		if ids[i] != expected[i] {
			t.Fatalf("ids[%d]: expected %q, got %q", i, expected[i], ids[i])
		}
	}
	if stream.Err() != nil {
		t.Fatalf("unexpected error: %v", stream.Err())
	}
}

func TestStreamErrorHandling(t *testing.T) {
	t.Run("error event sets error state", func(t *testing.T) {
		events := []ssestream.Event{
			{Type: "message_start", Data: []byte(`{"id":"1","data":"ok"}`)},
			{Type: "error", Data: []byte(`{"error":"bad"}`)},
		}
		decoder := &mockDecoder{events: events}
		stream := ssestream.NewStream[testStruct](decoder, nil)

		if !stream.Next() {
			t.Fatal("expected first event")
		}
		if stream.Current().ID != "1" {
			t.Fatalf("expected id 1, got %q", stream.Current().ID)
		}
		if stream.Next() {
			t.Fatal("expected stream to stop on error")
		}
		if stream.Err() == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("initial error prevents stream", func(t *testing.T) {
		decoder := &mockDecoder{err: errors.New("decoder error")}
		stream := ssestream.NewStream[testStruct](decoder, errors.New("initial error"))
		if stream.Next() {
			t.Fatal("expected false")
		}
		if stream.Err() == nil {
			t.Fatal("expected error")
		}
	})
}

func TestPeekBasic(t *testing.T) {
	events := []ssestream.Event{
		{Type: "message_start", Data: []byte(`{"id":"1","data":"first"}`)},
		{Type: "message_delta", Data: []byte(`{"id":"2","data":"second"}`)},
		{Type: "content_block_delta", Data: []byte(`{"id":"3","data":"third"}`)},
	}
	decoder := &mockDecoder{events: events}
	stream := ssestream.NewStream[testStruct](decoder, nil)

	peeked, ok := stream.Peek()
	if !ok {
		t.Fatal("expected peek ok")
	}
	if peeked.ID != "1" {
		t.Fatalf("expected peek id 1, got %q", peeked.ID)
	}

	peeked2, ok := stream.Peek()
	if !ok {
		t.Fatal("expected second peek ok")
	}
	if peeked2.ID != "1" {
		t.Fatalf("expected second peek id 1, got %q", peeked2.ID)
	}

	if !stream.Next() {
		t.Fatal("expected next")
	}
	if stream.Current().ID != "1" {
		t.Fatalf("expected current id 1, got %q", stream.Current().ID)
	}

	if !stream.Next() {
		t.Fatal("expected next")
	}
	if stream.Current().ID != "2" {
		t.Fatalf("expected current id 2, got %q", stream.Current().ID)
	}

	peeked3, ok := stream.Peek()
	if !ok {
		t.Fatal("expected third peek ok")
	}
	if peeked3.ID != "3" {
		t.Fatalf("expected peek id 3, got %q", peeked3.ID)
	}

	if stream.Err() != nil {
		t.Fatalf("unexpected error: %v", stream.Err())
	}
}

func TestPeekEmptyStream(t *testing.T) {
	decoder := &mockDecoder{events: []ssestream.Event{}}
	stream := ssestream.NewStream[testStruct](decoder, nil)

	peeked, ok := stream.Peek()
	if ok {
		t.Fatal("expected peek false")
	}
	var zero testStruct
	if peeked != zero {
		t.Fatal("expected zero value")
	}
}

func TestPeekAfterNext(t *testing.T) {
	events := []ssestream.Event{
		{Type: "message_start", Data: []byte(`{"id":"1","data":"first"}`)},
		{Type: "message_delta", Data: []byte(`{"id":"2","data":"second"}`)},
	}
	decoder := &mockDecoder{events: events}
	stream := ssestream.NewStream[testStruct](decoder, nil)

	if !stream.Next() {
		t.Fatal("expected next")
	}

	peeked, ok := stream.Peek()
	if !ok {
		t.Fatal("expected peek ok")
	}
	if peeked.ID != "2" {
		t.Fatalf("expected peek id 2, got %q", peeked.ID)
	}
}

func TestPeekAfterError(t *testing.T) {
	decoder := &mockDecoder{err: errors.New("decoder error")}
	stream := ssestream.NewStream[testStruct](decoder, errors.New("initial error"))

	peeked, ok := stream.Peek()
	if ok {
		t.Fatal("expected peek false")
	}
	var zero testStruct
	if peeked != zero {
		t.Fatal("expected zero value")
	}
}

func TestPeekN(t *testing.T) {
	events := []ssestream.Event{
		{Type: "message_start", Data: []byte(`{"id":"1","data":"first"}`)},
		{Type: "message_delta", Data: []byte(`{"id":"2","data":"second"}`)},
		{Type: "content_block_delta", Data: []byte(`{"id":"3","data":"third"}`)},
		{Type: "content_block_delta", Data: []byte(`{"id":"4","data":"fourth"}`)},
	}
	decoder := &mockDecoder{events: events}
	stream := ssestream.NewStream[testStruct](decoder, nil)

	peeks := stream.PeekN(3)
	if len(peeks) != 3 {
		t.Fatalf("expected 3, got %d", len(peeks))
	}
	for i, p := range peeks {
		expected := fmt.Sprintf("%d", i+1)
		if p.ID != expected {
			t.Fatalf("peeks[%d]: expected %q, got %q", i, expected, p.ID)
		}
	}

	if !stream.Next() || stream.Current().ID != "1" {
		t.Fatal("expected next=1")
	}
	if !stream.Next() || stream.Current().ID != "2" {
		t.Fatal("expected next=2")
	}

	peeks = stream.PeekN(2)
	if len(peeks) != 2 {
		t.Fatalf("expected 2, got %d", len(peeks))
	}
	if peeks[0].ID != "3" || peeks[1].ID != "4" {
		t.Fatalf("expected 3 and 4, got %q and %q", peeks[0].ID, peeks[1].ID)
	}

	if stream.Err() != nil {
		t.Fatalf("unexpected error: %v", stream.Err())
	}
}

func TestPeekNWithInsufficientEvents(t *testing.T) {
	events := []ssestream.Event{
		{Type: "message_start", Data: []byte(`{"id":"1","data":"first"}`)},
		{Type: "message_delta", Data: []byte(`{"id":"2","data":"second"}`)},
	}
	decoder := &mockDecoder{events: events}
	stream := ssestream.NewStream[testStruct](decoder, nil)

	peeks := stream.PeekN(5)
	if len(peeks) != 2 {
		t.Fatalf("expected 2, got %d", len(peeks))
	}

	if !stream.Next() || stream.Current().ID != "1" {
		t.Fatal("expected next=1")
	}

	peeks = stream.PeekN(5)
	if len(peeks) != 1 {
		t.Fatalf("expected 1, got %d", len(peeks))
	}
	if peeks[0].ID != "2" {
		t.Fatalf("expected 2, got %q", peeks[0].ID)
	}
}

func TestPeekNZeroOrNegative(t *testing.T) {
	decoder := &mockDecoder{events: []ssestream.Event{{Type: "message_start", Data: []byte(`{"id":"1","data":"x"}`)}}}
	stream := ssestream.NewStream[testStruct](decoder, nil)

	if stream.PeekN(0) != nil {
		t.Fatal("expected nil")
	}
	if stream.PeekN(-1) != nil {
		t.Fatal("expected nil")
	}
}

func TestPeekMixedWithNext(t *testing.T) {
	events := []ssestream.Event{
		{Type: "message_start", Data: []byte(`{"id":"1","data":"first"}`)},
		{Type: "message_delta", Data: []byte(`{"id":"2","data":"second"}`)},
		{Type: "content_block_delta", Data: []byte(`{"id":"3","data":"third"}`)},
		{Type: "content_block_delta", Data: []byte(`{"id":"4","data":"fourth"}`)},
	}
	decoder := &mockDecoder{events: events}
	stream := ssestream.NewStream[testStruct](decoder, nil)

	peek1, ok := stream.Peek()
	if !ok || peek1.ID != "1" {
		t.Fatalf("expected peek 1, ok=%v id=%q", ok, peek1.ID)
	}

	if !stream.Next() || stream.Current().ID != "1" {
		t.Fatal("expected next=1")
	}

	peek2 := stream.PeekN(2)
	if len(peek2) != 2 {
		t.Fatalf("expected 2, got %d", len(peek2))
	}
	if peek2[0].ID != "2" || peek2[1].ID != "3" {
		t.Fatalf("expected 2 and 3")
	}

	peek3, ok := stream.Peek()
	if !ok || peek3.ID != "2" {
		t.Fatalf("expected peek 2, ok=%v id=%q", ok, peek3.ID)
	}

	var ids []string
	for stream.Next() {
		ids = append(ids, stream.Current().ID)
	}

	expected := []string{"2", "3", "4"}
	if len(ids) != len(expected) {
		t.Fatalf("expected %d ids, got %d", len(expected), len(ids))
	}
	for i := range expected {
		if ids[i] != expected[i] {
			t.Fatalf("ids[%d]: expected %q, got %q", i, expected[i], ids[i])
		}
	}
}

func TestPeekBufferCleanup(t *testing.T) {
	events := []ssestream.Event{
		{Type: "message_start", Data: []byte(`{"id":"1","data":"first"}`)},
		{Type: "message_delta", Data: []byte(`{"id":"2","data":"second"}`)},
		{Type: "content_block_delta", Data: []byte(`{"id":"3","data":"third"}`)},
	}
	decoder := &mockDecoder{events: events}
	stream := ssestream.NewStream[testStruct](decoder, nil)

	stream.PeekN(3)
	stream.Next()
	stream.Next()
	stream.Next()

	peeks := stream.PeekN(10)
	if len(peeks) != 0 {
		t.Fatalf("expected empty, got %d", len(peeks))
	}
}
