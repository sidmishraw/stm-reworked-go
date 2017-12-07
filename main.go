/**
* main.go
* @author Sidharth Mishra
* @description The main driver for showcasing the STM examples
* @created Wed Nov 22 2017 21:40:28 GMT-0800 (PST)
* @copyright 2017 Sidharth Mishra
* @last-modified Fri Nov 24 2017 02:13:07 GMT-0800 (PST)
 */
package main

import (
	"log"
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
var MySTM = stm.NewSTM()

// MySlice is the wrapper type for my data to be used in the STM
type MySlice struct {
	Slice []int
}

// Clone returns a copy of the MySlice object. This is used for implementing STM's Data inteface.
func (ms *MySlice) Clone() Data {
	scopy := make([]int, len(ms.Slice))
	copy(scopy, ms.Slice)
	myslice := new(MySlice)
	myslice.Slice = scopy
	return Data(myslice)
}

func main() {

	// In go, the arrays are not reference types, so taking their address doesn't work as expected
	// In order to pass as references, we need to use slices.

	// for test1
	// data1T1 := Data([]int{1, 2, 3, 4, 5})
	// data2T1 := Data([]int{1, 2, 3, 4, 5})

	// for test2
	// casting Data, since I need a pointer to an interface and not the interface
	data1 := new(MySlice)
	data1.Slice = []int{1, 2, 3, 4, 5}

	// test1WithoutUsingSTMNL(data1)

	// log.Println()

	// test1WithoutUsingSTMWL(&data2T1)

	// log.Println()
	log.Println("data1 = ", data1)

	for j := 0; j < 1000; j++ {
		test2WithSTM(data1)
	}
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
func test1WithoutUsingSTMNL(data Data) {

	d2 := data.(*MySlice) // go's type assertion is tedious

	log.Println("d2 =", d2, " and d2 address = ", &d2)
	log.Println("data = ", data, " and address:: ", &data)

	for i := 0; i < 10001; i++ {
		wg := new(sync.WaitGroup) // the countdown latch in Go
		wg.Add(2)

		// starting `thread 1`.
		//--------------------
		// In `thread 1`, I'll try to update the 3rd position to 2 a 1000 times.
		go func(data *MySlice) {
			for i := 0; i < 1000; i++ { // for providing more uncertainity
				data.Slice[2] -= 2
			}
			wg.Done() // signal that one task is done
		}(d2)

		// starting `thread 2`.
		//----------------------
		// In `thread 2`, I'll try to update the 3rd position to 3 a 1000 times.
		go func(data *MySlice) {
			for i := 0; i < 1000; i++ { // for providing more uncertainity
				data.Slice[2] += 3
			}
			wg.Done()
		}(d2)

		wg.Wait() // wait till all tasks are done
		// if this wait group is not added, the main thread will resume
		// i.e. both the threads will run asynchronously.
		// That way, these types of operations will not behave like the bank account
		// transactions.
		// These waitgroups or countdownlatches can be used to make the main thread wait
		// till all the operations are complete and then the thread can proceed.

		// log.Println("Intermediate result of test 1 without locking")
		// log.Println("data = ", *data)
	}

	log.Println("Results of test 1 without locking")
	log.Println("d2 = ", d2)
	log.Println("data = ", data)

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
func test1WithoutUsingSTMWL(data Data) {
	d2 := data.(*MySlice) // tedious type assertion in Golang :(

	// cannot get the address of `(*data).([]int)`
	// because it is an expression and it is `homeless`
	// hence need to do it in 2 steps
	// d2 := (*data).([]int)
	// and get the address of d2 as &d2 <--- this is tedious -_-
	log.Println("d2 = ", d2, " and &d2 = ", &d2)

	mut := new(sync.Mutex)

	for i := 0; i < 10001; i++ {
		//# Transaction start ----
		wg := new(sync.WaitGroup) // the countdown latch in Go
		wg.Add(2)                 // added 2 to the countdownlatch or waitgroup

		// starting `thread 1`.
		//--------------------
		// In `thread 1`, I'll try to update the 3rd position to 2 a 1000 times.
		go func(data *MySlice) {
			for i := 0; i < 1000; i++ { // for providing more uncertainity
				mut.Lock() // take lock for determinism
				data.Slice[2] -= 2
				mut.Unlock() // unlock
			}
			wg.Done() // signal that one task is done
		}(d2)

		// starting `thread 2`.
		//----------------------
		// In `thread 2`, I'll try to update the 3rd position to 3 a 1000 times.
		go func(data *MySlice) {
			for i := 0; i < 1000; i++ { // for providing more uncertainity
				mut.Lock() // take lock for determinism
				data.Slice[2] += 3
				mut.Unlock() // unlock
			}
			wg.Done() // signal that one task is done and decrement the countdownlatch
		}(d2)

		wg.Wait() // wait till all tasks are done
		//# Transaction end ----

		// log.Println("Intermediate results of test 1 with locking ::")
		// log.Println("data = ", *data)
	}

	log.Println("WL:: Results of test 1 with locking")
	log.Println("d2 = ", d2)
	log.Println("data = ", data)
}

/* test2WithSTM :: For exhibiting an example of synchronization with STM */
func test2WithSTM(data Data) {

	cell1 := MySTM.MakeMemCell(data)

	//# t1 definition
	t1 := MySTM.NewT().
		Do(func(t *Transaction) bool {
			dinCell1 := t.ReadT(cell1).(*MySlice) // read data from memory cell - reads are transactional operations
			log.Println("dinCell1 = ", dinCell1)
			dinCell1.Slice[2] = dinCell1.Slice[2] - 2
			return t.WriteT(cell1, dinCell1)
		}).
		Done("T1")
	//# t1 definition

	//# t2 definition
	t2 := MySTM.NewT().
		Do(func(t *Transaction) bool {
			dinCell1 := t.ReadT(cell1).(*MySlice) // read data from memory cell - reads are transactional operations
			log.Println("dinCell1 = ", dinCell1)
			dinCell1.Slice[2] = dinCell1.Slice[2] + 3
			return t.WriteT(cell1, dinCell1)
		}).
		Done("T2")
	//# t2 definition

	//# simulate a 1000 times
	for i := 0; i < 1080; i++ {
		// For a 10001 times, I want to do similiar operations on the slice in the data.
		// I'll make 2 transactions.
		// t1 will update the 3rd posn in the array by subtracting 2 to it
		// -- this will simulate Person 1 withdrawing 100$ into Account 3
		// t2 will update the 3rd posn in the array by adding 3 to it
		// -- this will simulate Person 2 depositing 300$ into Account 3
		// Finally, the value of the 3rd position should be consistent.
		// log.Println("Starting Iteration #", i+1)
		MySTM.Exec(t1, t2)
		// time.Sleep(1 * time.Millisecond)
	}
	//# simulate a 1000 times

	MySTM.Display()

	MySTM.Exec(MySTM.NewT().
		Do(func(t *Transaction) bool {
			data := t.ReadT(cell1)
			log.Println("New data = ", data)
			return true
		}).
		Done())
}
