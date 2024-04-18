package main

import (
	"context"
	"runtime"
	"strings"
	"sync"
	"time"
)

// after 1 second, the first goroutine produced the result
// and sent it over the current channel, but there's already a
// goroutine in the mergeStringChans ready to forward the result
// to the trasnformToTitle stage that takes 1 more second to capitalize
// the string.
func RunPipeline2(ctx context.Context, source []string) <-chan string {

	outputChannel := producer2(ctx, source)

	stage1Channels := []<-chan string{}

	for i := 0; i < runtime.NumCPU(); i++ {
		lowerCaseChannel := transformToLower2(ctx, outputChannel)

		stage1Channels = append(stage1Channels, lowerCaseChannel)
	}

	stage1Merged := mergeStringChans2(ctx, stage1Channels...)
	stage2Channels := []<-chan string{}

	for i := 0; i < runtime.NumCPU(); i++ {
		titleCaseChannel := transformToTitle2(ctx, stage1Merged)

		stage2Channels = append(stage2Channels, titleCaseChannel)
	}

	return mergeStringChans2(ctx, stage2Channels...)
}

func producer2(ctx context.Context, strings []string) <-chan string {
	outChannel := make(chan string, len(strings))

	go func() {
		defer close(outChannel)

		for _, s := range strings {
			select {
			case <-ctx.Done():
				return
			default:
				outChannel <- s
			}
		}
	}()

	return outChannel
}

func transformToLower2(ctx context.Context, values <-chan string) <-chan string {
	outChannel := make(chan string)

	go func() {
		defer close(outChannel)

		select {
		case <-ctx.Done():
			return
		case s, ok := <-values:
			if ok {
				time.Sleep(time.Millisecond * 800)
				outChannel <- strings.ToLower(s)
			} else {
				return
			}
		}
	}()

	return outChannel
}

func transformToTitle2(ctx context.Context, values <-chan string) <-chan string {
	outChannel := make(chan string)

	go func() {
		defer close(outChannel)

		select {
		case <-ctx.Done():
			return
		case s, ok := <-values:
			if ok {
				time.Sleep(time.Millisecond * 800)
				outChannel <- strings.ToUpper(s[:1]) + s[1:]
			} else {
				return
			}
		}
	}()

	return outChannel
}

func mergeStringChans2(ctx context.Context, cs ...<-chan string) <-chan string {
	var wg sync.WaitGroup
	out := make(chan string)

	output := func(c <-chan string) {
		defer wg.Done()
		for n := range c {
			select {
			case out <- n:
			case <-ctx.Done():
				return
			}
		}
	}

	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
