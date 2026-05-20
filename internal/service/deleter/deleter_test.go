package service

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestDeleter_Enqueue(t *testing.T) {
	var mu sync.Mutex
	deleted := make(map[string][]string)

	markFunc := func(userID string, shorts []string) error {
		mu.Lock()
		defer mu.Unlock()
		deleted[userID] = append(deleted[userID], shorts...)
		return nil
	}

	deleter := NewDeleter(markFunc, nil)
	defer deleter.Close()

	task := DeleteTask{
		UserID: "user1",
		IDs:    []string{"abc123", "def456"},
	}

	err := deleter.Enqueue(task)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	time.Sleep(300 * time.Millisecond)

	mu.Lock()
	if len(deleted["user1"]) != 2 {
		t.Errorf("expected 2 deleted IDs, got %d", len(deleted["user1"]))
	}
	mu.Unlock()
}

func TestDeleter_Enqueue_MultipleTasks(t *testing.T) {
	var mu sync.Mutex
	deleted := make(map[string][]string)

	markFunc := func(userID string, shorts []string) error {
		mu.Lock()
		defer mu.Unlock()
		deleted[userID] = append(deleted[userID], shorts...)
		return nil
	}

	deleter := NewDeleter(markFunc, nil)
	defer deleter.Close()

	for i := 0; i < 10; i++ {
		task := DeleteTask{
			UserID: "user1",
			IDs:    []string{"id" + string(rune('0'+i))},
		}
		err := deleter.Enqueue(task)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}

	time.Sleep(300 * time.Millisecond)

	mu.Lock()
	if len(deleted["user1"]) != 10 {
		t.Errorf("expected 10 deleted IDs, got %d", len(deleted["user1"]))
	}
	mu.Unlock()
}

func TestDeleter_Enqueue_QueueFull(t *testing.T) {
	blockChan := make(chan struct{})
	markFunc := func(userID string, shorts []string) error {
		<-blockChan
		return nil
	}

	deleter := NewDeleter(markFunc, nil)
	defer deleter.Close()

	for i := 0; i < 100; i++ {
		task := DeleteTask{
			UserID: "user1",
			IDs:    []string{"id"},
		}
		deleter.Enqueue(task)
	}

	time.Sleep(100 * time.Millisecond)

	for i := 0; i < 1024; i++ {
		task := DeleteTask{
			UserID: "user1",
			IDs:    []string{"id"},
		}
		deleter.Enqueue(task)
	}

	task := DeleteTask{
		UserID: "user1",
		IDs:    []string{"id"},
	}
	err := deleter.Enqueue(task)
	if err != ErrQueueFull {
		t.Errorf("expected ErrQueueFull, got %v", err)
	}

	close(blockChan)
}

func TestDeleter_MarkFunc_Error(t *testing.T) {
	markFunc := func(userID string, shorts []string) error {
		return errors.New("mark error")
	}

	deleter := NewDeleter(markFunc, nil)
	defer deleter.Close()

	task := DeleteTask{
		UserID: "user1",
		IDs:    []string{"abc123"},
	}

	err := deleter.Enqueue(task)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	time.Sleep(300 * time.Millisecond)
}

func TestDeleter_Close(t *testing.T) {
	var mu sync.Mutex
	deleted := make(map[string][]string)

	markFunc := func(userID string, shorts []string) error {
		mu.Lock()
		defer mu.Unlock()
		deleted[userID] = append(deleted[userID], shorts...)
		return nil
	}

	deleter := NewDeleter(markFunc, nil)

	task := DeleteTask{
		UserID: "user1",
		IDs:    []string{"abc123"},
	}

	deleter.Enqueue(task)

	time.Sleep(100 * time.Millisecond)

	deleter.Close()

	mu.Lock()
	if len(deleted["user1"]) != 1 {
		t.Errorf("expected 1 deleted ID, got %d", len(deleted["user1"]))
	}
	mu.Unlock()
}

func TestDeleter_MultipleUsers(t *testing.T) {
	var mu sync.Mutex
	deleted := make(map[string][]string)

	markFunc := func(userID string, shorts []string) error {
		mu.Lock()
		defer mu.Unlock()
		deleted[userID] = append(deleted[userID], shorts...)
		return nil
	}

	deleter := NewDeleter(markFunc, nil)
	defer deleter.Close()

	for i := 0; i < 5; i++ {
		task := DeleteTask{
			UserID: "user1",
			IDs:    []string{"id1_" + string(rune('0'+i))},
		}
		deleter.Enqueue(task)

		task2 := DeleteTask{
			UserID: "user2",
			IDs:    []string{"id2_" + string(rune('0'+i))},
		}
		deleter.Enqueue(task2)
	}

	time.Sleep(300 * time.Millisecond)

	mu.Lock()
	if len(deleted["user1"]) != 5 {
		t.Errorf("expected 5 deleted IDs for user1, got %d", len(deleted["user1"]))
	}
	if len(deleted["user2"]) != 5 {
		t.Errorf("expected 5 deleted IDs for user2, got %d", len(deleted["user2"]))
	}
	mu.Unlock()
}
