package main

import (
	"fmt"
	"sync"
	"time"
)

// Fork represents a dining fork as a mutex
type Fork struct {
	id int
	sync.Mutex
}

// Philosopher represents a dining philosopher
type Philosopher struct {
	id                  int
	leftFork, rightFork *Fork
	eatenCount          int
}

// dine simulates a philosopher's behavior
// This implementation WILL DEADLOCK!
func (p *Philosopher) dine(wg *sync.WaitGroup) {
	defer wg.Done()

	for i := 0; i < 5; i++ {
		// Think for a bit
		p.think()

		// Pick up left fork first
		p.leftFork.Lock()
		fmt.Printf("Philosopher %d picked up left fork %d\n", p.id, p.leftFork.id)

		// Small delay to increase chance of deadlock
		time.Sleep(25 * time.Millisecond)

		// Pick up right fork
		p.rightFork.Lock()
		fmt.Printf("Philosopher %d picked up right fork %d\n", p.id, p.rightFork.id)

		// Eat
		p.eat()

		// Put down forks
		p.rightFork.Unlock()
		p.leftFork.Unlock()
		fmt.Printf("Philosopher %d put down both forks\n", p.id)
	}
}

func (p *Philosopher) think() {
	fmt.Printf("Philosopher %d is thinking...\n", p.id)
	time.Sleep(time.Duration(100+p.id*10) * time.Millisecond)
}

func (p *Philosopher) eat() {
	fmt.Printf("Philosopher %d is eating (meal %d)\n", p.id, p.eatenCount+1)
	p.eatenCount++
	time.Sleep(200 * time.Millisecond)
}

func main() {
	fmt.Println("=== Dining Philosophers: Deadlock Example ===")
	fmt.Println()

	// Create 5 forks
	forks := make([]*Fork, 5)
	for i := 0; i < 5; i++ {
		forks[i] = &Fork{id: i}
	}

	// Create 5 philosophers
	philosophers := make([]*Philosopher, 5)
	for i := 0; i < 5; i++ {
		philosophers[i] = &Philosopher{
			id:        i,
			leftFork:  forks[i],
			rightFork: forks[(i+1)%5], // Circular arrangement
		}
	}

	// Start dining with a timeout to detect deadlock
	var wg sync.WaitGroup
	start := time.Now()
	wg.Add(5)

	// Start all philosophers
	for i := 0; i < 5; i++ {
		go philosophers[i].dine(&wg)
	}

	// Wait with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		fmt.Println("\n✓ All philosophers finished dining!")
		for i, p := range philosophers {
			fmt.Printf("  Philosopher %d ate %d times\n", i, p.eatenCount)
		}
		elapsed := time.Since(start)
		fmt.Printf("\nSequential Dinner took %s\n\n", elapsed)
	case <-time.After(15 * time.Second):
		fmt.Println("\n✗ DEADLOCK DETECTED!")
		fmt.Println("All philosophers are waiting for forks.")
		fmt.Println("This is the classic deadlock scenario.")
	}
}
