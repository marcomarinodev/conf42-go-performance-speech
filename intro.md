# Introduction

### Elevator Pitch
Have you ever asked yourself why Go is fast and how to squeeze more speed from you Go apps? This talks dive deep into Go memory model and runtime scheduler and which best practices to use to make your Go code lightning fast.
## Presentation

Hello everyone, I'm Marco Marino and I work as a Software Engineer at ION. As the name suggests, every day I deal with improving a Platform as a Service where even the slightest delay can mean millions lost. So in this talk, I will discuss how to leverage the power of Go to build and optimize applications where speed is a requirement.

## Go Strengths for Performance

Go is a powerful programming language when it comes to developing high-performance applications, and here are the reasons why:

- Fast concurrency mechanism with Goroutines
- Efficient Scheduling
- Cooperative Multitasking
- Sharing Resources

### Why Go is known as a performant language?

We already know that Go is a performant language and that it is used in many applications, but why?

Go is based on the CSP Model (Communicating Sequential Processes) that describes how different processes communicate with each other. But let's use some Go language glossary. In our case, a process syntactically means a Goroutine, and two or more Goroutines use channels to communicate with each other. This approach is the opposite of what other languages like Python do, where there is a global shared data structure and different mechanisms to guarantee exclusive access like semaphores, locks, queues, and more. But all this complexity in Go is somehow hidden because Go knows how to deal with multiple tasks at once, and knows how to pass data between them, meaning low latency between two Goroutines communicating. In Go, in the context of multithreading, you don't write data to common storage. You create Goroutines to share data via channels. And because there is no need for exclusive access to global data structures, you gain speed.

Now, it is important to note that traditional mechanisms like mutexes or locks are not prohibited in Go, but simply not the default approach for a Go program.

But the real goodness starts when we start talking about the Go Runtime Scheduler.

#### Go Runtime Scheduler

In a standard OS, threads are scheduled after a certain amount of time (milliseconds) when a hardware timer interrupts the processor, the OS kernel suspends the current executing thread, saves its state in registries so that when it has to resume it, it doesn't lose anything, and then it finds among all the threads available for being executed, the next one to run. This process is known as **context switching** and is such a slow task because of cache misses and the number of memory accesses.

Go doesn't totally rely on the OS scheduler, but it has its own runtime scheduler and it uses a threading model called **M:N threading model** where `m` Goroutines are being scheduled among `n` OS threads. Unlike the OS scheduler that every time a hardware timer hits, it suspends the current thread, the Go runtime scheduler relies on its own language constructs. For example, when a sleep occurs or blocks in a channel, the scheduler simply suspends the Goroutine to execute another one, and this can happen on the same thread, so a context switching in this case is not required. That is why context switching in Golang is way cheaper than rescheduling a thread.

#### Go Memory Model

Now a question could be, but how does it save the current Goroutine state? Again let's do a comparison with what the OS does with threads.

OS threads have a fixed-size stack where it saves the local variables of an in-progress function call or a suspended one. The fact that the stack has a fixed size is kind of a problem because it is a huge waste for a single Goroutine, for example, but at the same time, it could be too strict for hundreds of thousands of Goroutines being created, which is not a rare event in a Go-based application.

In contrast to this approach, Go creates very small stacks for each Goroutine, around 2KB, and the surprising thing is that each of these stacks is growing and shrinking as needed, and this is possible because of the capability of these stacks to borrow memory from the heap.

The fact that the stacks are dynamic allows us to also have better memory management as memory overhead is reduced when context switching happens, and that's because of the small states we have to save for each Goroutine.
