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

		waitStart := time.Now()

		// Ask waiter for permission
		p.waiter.RequestPermission(p.id)

		// Now safe to pick up forks
		p.pickUpForks()
		metrics.addWait(time.Since(waitStart))

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
	metrics.startEating()
	time.Sleep(p.eatTime)
	metrics.stopEating()
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

	waiter := NewWaiter(seats)

	// Create 5 philosophers
	philosophers := make([]*Philosopher, 5)
	for i := 0; i < 5; i++ {
		philosophers[i] = &Philosopher{
			id:        i,
			leftFork:  forks[i],
			rightFork: forks[(i+1)%5],
			waiter:    waiter,
			thinkTime: 100 * time.Millisecond,
			eatTime:   200 * time.Millisecond,
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
		metrics.report(fmt.Sprintf("V2-seats%d", seats), elapsed, philosophers)
		fmt.Println("Meal counts:")
		totalMeals := 0
		for i, p := range philosophers {
			fmt.Printf("  Philosopher %d: %d meals\n", i, p.eatenCount)
			totalMeals += p.eatenCount
		}
		fmt.Printf("  Total: %d meals\n", totalMeals)
		fmt.Println("\n✓ No deadlock! The waiter solution works!")

	case <-time.After(10 * time.Minute): // ALTERADO: 30s podia estourar com tempos maiores
		fmt.Println("\n✗ Timeout! Something went wrong.")
	}
}

const seats = 2

type Metrics struct {
	mu        sync.Mutex
	totalWait time.Duration // soma de todas as esperas
	maxWait   time.Duration // maior espera individual
	waits     int           // quantas esperas foram medidas (= refeições)
	eatingNow int           // quantos comem agora
	maxEating int           // máximo de filósofos comendo ao mesmo tempo
}

var metrics Metrics

// addWait registra quanto tempo um filósofo esperou para conseguir comer.
func (m *Metrics) addWait(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.totalWait += d
	m.waits++
	if d > m.maxWait {
		m.maxWait = d
	}
}

func (m *Metrics) startEating() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.eatingNow++
	if m.eatingNow > m.maxEating {
		m.maxEating = m.eatingNow
	}
}

func (m *Metrics) stopEating() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.eatingNow--
}

func (m *Metrics) report(name string, total time.Duration, philosophers []*Philosopher) {
	m.mu.Lock()
	defer m.mu.Unlock()

	avg := time.Duration(0)
	if m.waits > 0 {
		avg = m.totalWait / time.Duration(m.waits)
	}

	fmt.Println("\n===== MÉTRICAS", name, "=====")
	fmt.Println("Tempo total de execução:         ", total)
	fmt.Println("Tempo médio de espera (garfos):  ", avg)
	fmt.Println("Maior tempo de espera:           ", m.maxWait)
	fmt.Println("Máx. filósofos comendo juntos:   ", m.maxEating)
	fmt.Print("Ciclos concluídos por filósofo:   ")
	for _, p := range philosophers {
		fmt.Printf("%d ", p.eatenCount)
	}
	fmt.Println()
	// linha pronta para copiar na planilha: versão;total(s);média(s);máx(s);máx comendo
	fmt.Printf("CSV;%s;%.3f;%.3f;%.3f;%d\n", name, total.Seconds(), avg.Seconds(), m.maxWait.Seconds(), m.maxEating)
}
