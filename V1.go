package main

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
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
	state               atomic.Int32
}

// dine simulates a philosopher's behavior
// This implementation WILL DEADLOCK!
func (p *Philosopher) dine(wg *sync.WaitGroup) {
	defer wg.Done()

	time.Sleep(time.Duration(rand.Int63n(int64(startJitter))))

	for i := 0; i < 20; i++ {
		// Think for a bit
		p.state.Store(stateThinking)
		p.think()

		waitStart := time.Now()

		// Pick up left fork first
		p.leftFork.Lock()
		p.state.Store(stateHoldingLeft)
		fmt.Printf("Philosopher %d picked up left fork %d\n", p.id, p.leftFork.id)

		// Small delay to increase chance of deadlock
		time.Sleep(10 * time.Millisecond)

		// Pick up right fork
		p.rightFork.Lock()
		fmt.Printf("Philosopher %d picked up right fork %d\n", p.id, p.rightFork.id)
		metrics.addWait(time.Since(waitStart))
		p.state.Store(stateEating)

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
	time.Sleep(100 * time.Millisecond)
}

func (p *Philosopher) eat() {
	fmt.Printf("Philosopher %d is eating (meal %d)\n", p.id, p.eatenCount+1)
	p.eatenCount++
	metrics.startEating()
	time.Sleep(200 * time.Millisecond)
	metrics.stopEating()
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
		fmt.Printf("\nDinner took %s\n\n", elapsed)
		metrics.report("V1", elapsed, philosophers)
	case <-time.After(30 * time.Second):
		fmt.Println("\n✗ DEADLOCK DETECTED!")
		fmt.Println("All philosophers are waiting for forks.")
		fmt.Println("This is the classic deadlock scenario.")

		fmt.Println("\nEstado de cada filósofo no momento do deadlock:")
		for i, p := range philosophers {
			right := (i + 1) % 5
			switch p.state.Load() {
			case stateHoldingLeft:
				dono := "ninguém"
				if philosophers[right].state.Load() == stateHoldingLeft {
					dono = fmt.Sprintf("o filósofo %d", right)
				}
				fmt.Printf("  Filósofo %d: SEGURA garfo %d | ESPERA garfo %d (que está com %s)\n",
					i, p.leftFork.id, p.rightFork.id, dono)
			case stateEating:
				fmt.Printf("  Filósofo %d: comendo\n", i)
			default:
				fmt.Printf("  Filósofo %d: pensando (não segura nada)\n", i)
			}
		}
	}
}

const (
	stateThinking int32 = iota
	stateHoldingLeft
	stateEating
)

type Metrics struct {
	mu        sync.Mutex
	totalWait time.Duration
	maxWait   time.Duration
	waits     int
	eatingNow int
	maxEating int
}

var metrics Metrics

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
	fmt.Printf("CSV;%s;%.3f;%.3f;%.3f;%d\n", name, total.Seconds(), avg.Seconds(), m.maxWait.Seconds(), m.maxEating)
}

const startJitter = 25 * time.Millisecond
