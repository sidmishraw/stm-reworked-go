/**
* stm-reworked
* stm_reworked
*
* @author        sidmishraw
* @created-on    11/22/17 9:40 PM
*
 */
package main

import (
	"fmt"
	"sync"

	"github.com/sidmishraw/stm-reworked/stm"
)

/*
Data :: A typealias for stm.Data
*/
type Data = stm.Data

/*
STM :: A typealias for stm.STM
*/
type STM = stm.STM

/*Transaction :: A typealias for stm.Transaction*/
type Transaction = stm.Transaction

// MySTM :: The STM
var MySTM = new(STM)

func main() {

	// In go, the arrays are not reference types, so taking their address doesn't work as expected
	// In order to pass as references, we need to use slices.

	// for test1
	data1T1 := Data([]int{1, 2, 3, 4, 5})
	data2T1 := Data([]int{1, 2, 3, 4, 5})

	// for test2
	// casting Data, since I need a pointer to an interface and not the interface
	data1 := Data([]int{1, 2, 3, 4, 5})

	test1WithoutUsingSTMNL(&data1T1)

	fmt.Println()

	test1WithoutUsingSTMWL(&data2T1)
	test2WithSTM(&data1)
}

/*

test1WithoutUsingSTM

----------------------

For exhibiting an example of problems with multithreading -- non-determinism.

Run#1
Results of test 1 without locking
d2 =  [1 2 5001962 4 5]
data =  [1 2 5001962 4 5]

Run#2
Results of test 1 without locking
d2 =  [1 2 5252101 4 5]
data =  [1 2 5252101 4 5]

Run#3
Results of test 1 without locking
d2 =  [1 2 5566628 4 5]
data =  [1 2 5566628 4 5]

Run#4
Results of test 1 without locking
d2 =  [1 2 5555669 4 5]
data =  [1 2 5555669 4 5]

*/
func test1WithoutUsingSTMNL(data *Data) {

	d2 := (*data).([]int) // go's type assertion is tedious

	fmt.Println("d2 =", d2, " and d2 address = ", &d2)
	fmt.Println("data = ", *data, " and address:: ", data)

	for i := 0; i < 10001; i++ {
		wg := new(sync.WaitGroup) // the countdown latch in Go
		wg.Add(2)

		// starting `thread 1`.
		//--------------------
		// In `thread 1`, I'll try to update the 3rd position to 2 a 1000 times.
		go func(data *[]int) {
			// fmt.Println("t1")
			for i := 0; i < 1000; i++ { // for providing more uncertainity
				(*data)[2] -= 2
				// fmt.Println("t1 changed to 2")
			}
			wg.Done() // signal that one task is done
			// fmt.Println("t1 - d")
		}(&d2)

		// starting `thread 2`.
		//----------------------
		// In `thread 2`, I'll try to update the 3rd position to 3 a 1000 times.
		go func(data *[]int) {
			// fmt.Println("t2")
			for i := 0; i < 1000; i++ { // for providing more uncertainity
				(*data)[2] += 3
				// fmt.Println("t2 changed to 3")
			}
			wg.Done()
			// fmt.Println("t2 - d")
		}(&d2)

		wg.Wait() // wait till all tasks are done
		// if this wait group is not added, the main thread will resume
		// i.e. both the threads will run asynchronously.
		// That way, these types of operations will not behave like the bank account
		// transactions.
		// These waitgroups or countdownlatches can be used to make the main thread wait
		// till all the operations are complete and then the thread can proceed.

		// fmt.Println("Intermediate result of test 1 without locking")
		// fmt.Println("data = ", *data)
	}

	fmt.Println("Results of test 1 without locking")
	fmt.Println("d2 = ", d2)
	fmt.Println("data = ", *data)

	// As it can be seen from this example, it is not clear what the value of the 3rd
	// position in the array will be - it can be 2 or 3.
}

/*
test1WithoutUsingSTMWL

-----------------------

Testing with Locking(mutex) - non-determinism has been pruned by using locks and waitgroups

Run#1
WL:: Results of test 1 with locking
d2 =  [1 2 10001003 4 5]
data =  [1 2 10001003 4 5]

Run#2
WL:: Results of test 1 with locking
d2 =  [1 2 10001003 4 5]
data =  [1 2 10001003 4 5]

Run#3
WL:: Results of test 1 with locking
d2 =  [1 2 10001003 4 5]
data =  [1 2 10001003 4 5]

WL:: Results of test 1 with locking
d2 =  [1 2 10001003 4 5]
data =  [1 2 10001003 4 5]
*/
func test1WithoutUsingSTMWL(data *Data) {
	d2 := (*data).([]int) // tedious type assertion in Golang :(
	// cannot get the address of `(*data).([]int)`
	// because it is an expression and it is `homeless`
	// hence need to do it in 2 steps
	// d2 := (*data).([]int)
	// and get the address of d2 as &d2 <--- this is tedious -_-

	fmt.Println("d2 = ", d2, " and &d2 = ", &d2)

	mut := new(sync.Mutex)

	for i := 0; i < 10001; i++ {
		//# Transaction start ----
		wg := new(sync.WaitGroup) // the countdown latch in Go
		wg.Add(2)                 // added 2 to the countdownlatch or waitgroup

		// starting `thread 1`.
		//--------------------
		// In `thread 1`, I'll try to update the 3rd position to 2 a 1000 times.
		go func(data *[]int) {
			for i := 0; i < 1000; i++ { // for providing more uncertainity
				mut.Lock() // take lock for determinism
				(*data)[2] -= 2
				mut.Unlock() // unlock
			}
			wg.Done() // signal that one task is done
		}(&d2)

		// starting `thread 2`.
		//----------------------
		// In `thread 2`, I'll try to update the 3rd position to 3 a 1000 times.
		go func(data *[]int) {
			for i := 0; i < 1000; i++ { // for providing more uncertainity
				mut.Lock() // take lock for determinism
				(*data)[2] += 3
				mut.Unlock() // unlock
			}
			wg.Done() // signal that one task is done and decrement the countdownlatch
		}(&d2)

		wg.Wait() // wait till all tasks are done
		//# Transaction end ----

		// fmt.Println("Intermediate results of test 1 with locking ::")
		// fmt.Println("data = ", *data)
	}

	fmt.Println("WL:: Results of test 1 with locking")
	fmt.Println("d2 = ", d2)
	fmt.Println("data = ", *data)
}

/* test2WithSTM :: For exhibiting an example of synchronization with STM */
func test2WithSTM(data *Data) {
	cell1 := MySTM.MakeMemCell(data)
	for i := 0; i < 10001; i++ {
		// For a 10001 times, I want to do similiar operations on the slice in the data.
		// I'll make 2 transactions.
		// t1 will update the 3rd posn in the array by subtracting 2 to it
		// -- this will simulate Person 1 withdrawing 100$ into Account 3
		// t2 will update the 3rd posn in the array by adding 3 to it
		// -- this will simulate Person 2 depositing 300$ into Account 3
		// Finally, the value of the 3rd position should be consistent.
		t1 := MySTM.NewT().
			Do(func(t *Transaction) bool { return true }).
			Do(func(t *Transaction) bool { return true }).
			Do(func(t *Transaction) bool { return true }).
			Done()
		fmt.Println(t1)
	}
	fmt.Println(cell1)
}