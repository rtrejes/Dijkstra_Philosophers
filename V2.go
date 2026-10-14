package main

import (
	"fmt"
	"sync"
	"time"
)

// Fork represents a dining fork
type Fork struct {
	id int
	sync.Mutex
}

// Waiter controls access to the dining table using a semaphore
type Waiter struct {
	semaphore chan struct{}
}

// NewWaiter creates a waiter that allows up to maxSeated philosophers
func NewWaiter(maxSeated int) *Waiter {
	return &Waiter{
		semaphore: make(chan struct{}, maxSeated),
	}
}

// RequestPermission asks the waiter for permission to pick up forks
func (w *Waiter) RequestPermission(philosopherID int) {
	fmt.Printf("Philosopher %d is requesting permission from waiter\n", philosopherID)
	w.semaphore <- struct{}{} // Acquire semaphore
	fmt.Printf("Philosopher %d got permission from waiter\n", philosopherID)
}

// ReleasePermission returns permission to the waiter
func (w *Waiter) ReleasePermission(philosopherID int) {
	<-w.semaphore // Release semaphore
	fmt.Printf("Philosopher %d returned permission to waiter\n", philosopherID)
}

// Philosopher represents a dining philosopher
type Philosopher struct {
	id                  int
	leftFork, rightFork *Fork
	waiter              *Waiter
	eatenCount          int
	thinkTime           time.Duration
	eatTime             time.Duration
}

// dine simulates a philosopher's behavior with waiter coordination
func (p *Philosopher) dine(wg *sync.WaitGroup) {
	defer wg.Done()

	for i := 0; i < 20; i++ {
		// Think
		p.think()

		// Ask waiter for permission
		p.waiter.RequestPermission(p.id)

		// Now safe to pick up forks
		p.pickUpForks()

		// Eat
		p.eat()

		// Put down forks
		p.putDownForks()

		// Return permission to waiter
		p.waiter.ReleasePermission(p.id)
	}
}

func (p *Philosopher) think() {
	fmt.Printf("Philosopher %d is thinking...\n", p.id)
	time.Sleep(p.thinkTime)
}

func (p *Philosopher) pickUpForks() {
	// Pick up left fork
	p.leftFork.Lock()
	fmt.Printf("Philosopher %d picked up left fork %d\n", p.id, p.leftFork.id)

	// Pick up right fork
	p.rightFork.Lock()
	fmt.Printf("Philosopher %d picked up right fork %d\n", p.id, p.rightFork.id)
}

func (p *Philosopher) eat() {
	fmt.Printf("Philosopher %d is eating (meal %d)\n", p.id, p.eatenCount+1)
	p.eatenCount++
	time.Sleep(p.eatTime)
}

func (p *Philosopher) putDownForks() {
	p.rightFork.Unlock()
	p.leftFork.Unlock()
	fmt.Printf("Philosopher %d put down both forks\n", p.id)
}

func main() {
	fmt.Println("=== Dining Philosophers: Waiter Solution ===")
	fmt.Println("Using a semaphore to prevent deadlock")
	fmt.Println()

	// Create 5 forks
	forks := make([]*Fork, 5)
	for i := 0; i < 5; i++ {
		forks[i] = &Fork{id: i}
	}

	// Create waiter that allows 4 philosophers at a time (N-1)
	waiter := NewWaiter(4)

	// Create 5 philosophers
	philosophers := make([]*Philosopher, 5)
	for i := 0; i < 5; i++ {
		philosophers[i] = &Philosopher{
			id:        i,
			leftFork:  forks[i],
			rightFork: forks[(i+1)%5],
			waiter:    waiter,
			thinkTime:  100 * time.Millisecond,
			eatTime:    200 * time.Millisecond,
		}
	}

	// Start dining
	var wg sync.WaitGroup
	startTime := time.Now()

	fmt.Println("Starting philosophers...")
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go philosophers[i].dine(&wg)
	}

	// Wait for completion with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		elapsed := time.Since(startTime)
		fmt.Println("\n✓ All philosophers finished dining!")
		fmt.Printf("Total time: %v\n\n", elapsed)
		fmt.Println("Meal counts:")
		totalMeals := 0
		for i, p := range philosophers {
			fmt.Printf("  Philosopher %d: %d meals\n", i, p.eatenCount)
			totalMeals += p.eatenCount
		}
		fmt.Printf("  Total: %d meals\n", totalMeals)
		fmt.Println("\n✓ No deadlock! The waiter solution works!")

	case <-time.After(30 * time.Second);
		fmt.Println("\n✗ Timeout! Something went wrong.")
	}
}