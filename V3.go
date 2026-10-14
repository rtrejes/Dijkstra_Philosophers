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

// Philosopher represents a dining philosopher
type Philosopher struct {
	id                  int
	leftFork, rightFork *Fork
	eatenCount          int
	asymmetric          bool // True if this philosopher uses asymmetric ordering
}

// NewPhilosopher creates a philosopher with the correct fork ordering
func NewPhilosopher(id int, forks []*Fork, numPhilosophers int) *Philosopher {
	leftForkID := id
	rightForkID := (id + 1) % numPhilosophers

	// The last philosopher uses asymmetric ordering
	// This breaks the circular wait
	asymmetric := (id == numPhilosophers-1)

	return &Philosopher{
		id:         id,
		leftFork:   forks[leftForkID],
		rightFork:  forks[rightForkID],
		asymmetric: asymmetric,
	}
}

// dine simulates a philosopher's behavior
func (p *Philosopher) dine(wg *sync.WaitGroup) {
	defer wg.Done()

	for i := 0; i < 20; i++ {
		p.think()
		waitStart := time.Now() 
		p.pickUpForks()
		metrics.addWait(time.Since(waitStart)) 
		p.eat()
		p.putDownForks()
	}
}

func (p *Philosopher) think() {
	fmt.Printf("Philosopher %d is thinking...\n", p.id)
	time.Sleep(100 * time.Millisecond)
}

func (p *Philosopher) pickUpForks() {
	if p.asymmetric {
		// Asymmetric: pick up RIGHT fork first, then LEFT
		p.rightFork.Lock()
		fmt.Printf("Philosopher %d (asymmetric) picked up RIGHT fork %d first\n",
			p.id, p.rightFork.id)

		p.leftFork.Lock()
		fmt.Printf("Philosopher %d (asymmetric) picked up LEFT fork %d second\n",
			p.id, p.leftFork.id)
	} else {
		// Normal: pick up LEFT fork first, then RIGHT
		p.leftFork.Lock()
		fmt.Printf("Philosopher %d picked up LEFT fork %d\n", p.id, p.leftFork.id)

		p.rightFork.Lock()
		fmt.Printf("Philosopher %d picked up RIGHT fork %d\n", p.id, p.rightFork.id)
	}
}

func (p *Philosopher) eat() {
	fmt.Printf("Philosopher %d is eating (meal %d)\n", p.id, p.eatenCount+1)
	p.eatenCount++
	metrics.startEating() 
	time.Sleep(200 * time.Millisecond)
	metrics.stopEating() 
}

func (p *Philosopher) putDownForks() {
	p.rightFork.Unlock()
	p.leftFork.Unlock()
	fmt.Printf("Philosopher %d put down both forks\n", p.id)
}

func main() {
	fmt.Println("=== Dining Philosophers: Asymmetric Solution ===")
	fmt.Println("Breaking circular wait with resource ordering")
	fmt.Println()

	numPhilosophers := 5

	// Create 5 forks
	forks := make([]*Fork, numPhilosophers)
	for i := 0; i < numPhilosophers; i++ {
		forks[i] = &Fork{id: i}
	}

	// Create philosophers with asymmetric ordering for the last one
	philosophers := make([]*Philosopher, numPhilosophers)
	for i := 0; i < numPhilosophers; i++ {
		philosophers[i] = NewPhilosopher(i, forks, numPhilosophers)
	}

	// Verify the ordering
	fmt.Println("Fork acquisition order:")
	for i, p := range philosophers {
		if p.asymmetric {
			fmt.Printf("  Philosopher %d: RIGHT fork %d → LEFT fork %d (ASYMMETRIC)\n",
				i, p.rightFork.id, p.leftFork.id)
		} else {
			fmt.Printf("  Philosopher %d: LEFT fork %d → RIGHT fork %d\n",
				i, p.leftFork.id, p.rightFork.id)
		}
	}
	fmt.Println()

	// Start dining
	var wg sync.WaitGroup
	startTime := time.Now()

	fmt.Println("Starting philosophers...")
	for i := 0; i < numPhilosophers; i++ {
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
		metrics.report("V3", elapsed, philosophers) 

		fmt.Println("Meal counts:")
		totalMeals := 0
		for i, p := range philosophers {
			marker := ""
			if p.asymmetric {
				marker = " (asymmetric)"
			}
			fmt.Printf("  Philosopher %d: %d meals%s\n", i, p.eatenCount, marker)
			totalMeals += p.eatenCount
		}
		fmt.Printf("  Total: %d meals\n", totalMeals)
		fmt.Println("\n✓ No deadlock! Asymmetric solution works!")

	case <-time.After(10 * time.Minute): // ALTERADO: 30s podia estourar com tempos maiores
		fmt.Println("\n✗ Timeout! Unexpected deadlock occurred.")
	}
}


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