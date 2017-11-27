# An Optimistic Software Transactional Memory - reworked

This is a reworked optimistic STM in Go.

## Features:

Composable concurrency. Pass around concurrent pieces of code.

```go
//# t1 definition
t1 := MySTM.NewT().
  Do(func(t *Transaction) bool {
    dinCell1 := t.ReadT(cell1).([]int) // read data from memory cell - reads are transactional operations
    fmt.Println("T1 :: Data in cell1 = dinCell1 = ", dinCell1)
    dinCell1[2] = dinCell1[2] - 2
    return t.WriteT(cell1, dinCell1)
  }).
  Do(func(t *Transaction) bool {
    dinCell1 := t.ReadT(cell1).([]int) // read data from memory cell - reads are transactional operations
    fmt.Println("T1' :: Data in cell1 = dinCell1 = ", dinCell1)
    return true
  }).
  Done()
//# t1 definition

//# t1 invocation somewhere else
// This is WIP, will refine some more
wg := new(sync.WaitGroup)
wg.Add(1)
t1.Go(wg) // executes the Transaction t1 concurrently
wg.Wait()
// This is WIP, will refine some more
//# t1 invocation somewhere else
```

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
