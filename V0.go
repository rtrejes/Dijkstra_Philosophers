// Example code filo_seq.go - Sequential Dinning Philosophers
// Since is sequential, elapsed timne could be simply calculated by (roubnds*philosophers*(time_to_think+time_to_eat))
// By Cesar De Rose
package main

import (
	"fmt"
	"time"
)

func Philosopher(id, iteration int, timeToThink, timeToEat time.Duration) {
	// think
	for s := 1; s <= id*6; s++ {
		fmt.Printf(" ")
	}
	fmt.Printf(" T%d\n", iteration) // Thinking
	time.Sleep(timeToThink)

	// eat
	for s := 1; s <= id*6; s++ {
		fmt.Printf(" ")
	}
	fmt.Printf(" E%d\n", iteration) // Eating
	time.Sleep(timeToEat)
}

func main() {
	philosophers := 5
	rounds := 20

	timeToThink := 100 * time.Millisecond
	timeToEat := 200 * time.Millisecond

	fmt.Println("\n[P1]  [P2]  [P3]  [P4]  [P5]\n")

	// run philosophers sequentially
	start := time.Now()

	for i := 1; i <= rounds; i++ {
		for j := 0; j < philosophers; j++ {
			Philosopher(j, i, timeToThink, timeToEat)
		}
	}
	elapsed := time.Since(start)

	fmt.Printf("\nSequential Dinner took %s\n\n", elapsed)
}