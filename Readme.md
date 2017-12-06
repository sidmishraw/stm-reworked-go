# An Optimistic Software Transactional Memory - reworked

Author: Sidharth Mishra

</br>

`This is the reworked optimistic STM in Go.`

</br>
</br>

## Features

Composable concurrency. Pass around concurrent pieces of code.

```go
//# t1 definition
t1 := MySTM.NewT().
        Do(func(t *Transaction) bool {
          dinCell1 := make([]int, 0)
          t.ReadT(cell1, &dinCell1) // read data from memory cell - reads are transactional operations
          dinCell1[2] = dinCell1[2] - 2
          return t.WriteT(cell1, dinCell1)
        }).
        Done("T1")
//# t1 definition

//# t1 invocation somewhere else. Manual style
// This is WIP, will refine some more
wg := new(sync.WaitGroup)
wg.Add(1)
t1.Go(wg) // executes the Transaction t1 concurrently
wg.Wait()
// This is WIP, will refine some more
//# t1 invocation somewhere else

//# t1 invocation somewhere else. Using builtin utility
// Execute t1 synchronously
// Makes the calling thread wait till t1 is done executing
MySTM.Exec(t1)

// Execute t1 asynchronously
MySTM.ForkAndExec(t1)
//# t1 invocation somewhere else
```

</br>
</br>

## Breaking changes from v0.0.1

* Full rework of the data storage mechanism in `MemoryCell`s. Used `gob` package's
  `Encoder` and `Decoder` to encode and decode data. The new MemoryCells hold data in form
  of bytes rather than a pointer to `Data`.

* Added a `Scan` phase to the transaction execution workflow. In this phase, I do a dry
  run for the actions. This is to make sure that I can have the `Do-Done` way of building
  transactions. When the transaction is scanning, it sets the `isScanning` flag to true.
  The `ReadT` and `WriteT` operations behave differently when the transaction is in `scan`
  phase. The actual execution happens in the `Execute actions phase`.

* Due to the addition of the `bytes.Buffer` to the MemoryCells to hold data, the `ReadT`
  API has now changed.

  The old way to read data from the transaction was:

  ```Go
  dataInMemCell := t.ReadT(memCell).([]int)
  ```

  The problem with the above method was that the data I returned was not a clone/copy.
  Changes to the data read would modify the data in the actual `MemoryCell` -- my mistake.
  To work around it, I decided to store data in form of bytes. By storing the data in form
  of bytes, I can just return the byte slice from the bytes.Buffer in the MemoryCell. I've
  tested it out and it works great -- performance evaluations have not been made yet.

  The new way of reading data from the STM inside a transaction is:

  ```Go
  dataInCell := make([]int, 0)
  t.ReadT(memCell, &dataInCell)
  ```

  This is because of the nature of reading data out of the bytes.Buffer using gob's
  Decoder.

* I also added the `Take Ownership` phase. I tried to combine the `Take Ownership` with
  the `WriteT` workflow in v0.0.1 but, I failed miserably. With the separation, it has
  made the code more robust while keeping the API changes minimal.

* With the new changes -- introduction of the `Scan` phase -- it has become more critical
  that no external variables(ones outside the STM) be used directly inside the
  transaction's actions or closures.

> Note: The `Scan` phase actually executes the actions/closures. Any code inside the
> actions will get executed.

* I was not able to find any good solutions for achieving Monadic behavior in Go --
  especially IO Monad. Because of this reason, I have provided the `Where` function. By
  using the `Where` function, the consumer can add `transactional variables` that are
  valid inside the transaction. These variables are owned by the transaction and can be
  used freely inside transaction's actions. But, make sure that you don't store actual
  references inside these transactional variables. Changes to them might cause bugs.

> Note: Please use copies of values when storing data in transactional variables. It might
> cause weird bugs :)

```Go
// Usage of t.GetTVar - use for reading transactional variables
v := t.GetTVar("foo")
if v != nil {
    // this portion of the code will not run in scan phase
    // so make sure that you don't put `WriteT` operations inside this
}

// Usage of t.PutTVar - use for writing transactional variables
t.PutTVar("foo", []int{1,2,3})
```

</br>
</br>

## Word from the author

The code has been verified using the following example:

The STM has 1 memory cell. This memory cell holds a slice `[]int{1,2,3,4,5}`. There are 2
transactions `T1` and `T2`.

* T1 will subtract 2 from the 3rd element of the slice.

* T2 will add 3 to the 3rd element of the slice.

The transactions T1 and T2 are executed concurrently. By virtue of `ACI\* compliance`, the
consistent state of the slice `[]int{1,2,3,4,5}` after T1 and T2 have operated on it is
`[]int{1,2,4,4,5}`.

Effectively, the 3rd element's value should increase by 1. The order the transactions
operated doesn't matter.

The simulation verifies the consistency by running the transactions T1 and T2 concurrently
for `1080` times. Furthermore, the entire simulation is run a `1000` times.

The logs of the operations can be found in the [root.log](./root.log) file.

</br>
</br>
</br>

## DISCLAIMER

<footer>
<p>
  <strong>
    Please do not CLONE or FORK or CONTRIBUTE. This is my homework xD!
  </strong>
</p>
<p>
Inspirations and references:

[[1]](https://doi.org/10.1007/s004460050028) N. Shavit and D. Touitou, "Software
transactional memory", Distrib. Comput., vol. 10, no. 2, pp. 99-116, Feb 1997 [Online].
Available: https://doi.org/10.1007/s004460050028

[[2]](https://michel.weimerskirch.net/wp-content/uploads/2008/02/software_transactional_memory.pdf)
M. Weimerskirch, “Software transactional memory,” [Online]. Available:
https://michel.weimerskirch.net/wp-content/uploads/2008/02/software_transactional_memory.pdf

[[3]](https://www.schoolofhaskell.com/school/advanced-haskell/beautiful-concurrency) S. P.
Jones, “Beautiful concurrency,” [Online]. Available:
https://www.schoolofhaskell.com/school/advanced-haskell/beautiful-concurrency

</p>
</footer>
