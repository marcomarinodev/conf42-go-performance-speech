package main

import (
	"context"
	"runtime"
	"sync"
	"time"
	"unicode"
)

// go:noinline
func RunSlowPipeline(ctx context.Context, source []string) <-chan string {

	outputChannel := producerSlow(ctx, source)

	stage1Channels := []<-chan string{}

	for i := 0; i < runtime.NumCPU(); i++ {
		lowerCaseChannel := transformToLowerSlow(ctx, outputChannel)

		stage1Channels = append(stage1Channels, lowerCaseChannel)
	}

	stage1Merged := mergeStringChansSlow(ctx, stage1Channels...)
	stage2Channels := []<-chan string{}

	for i := 0; i < runtime.NumCPU(); i++ {
		titleCaseChannel := transformToTitleSlow(ctx, stage1Merged)

		stage2Channels = append(stage2Channels, titleCaseChannel)
	}

	return mergeStringChansSlow(ctx, stage2Channels...)
}

func producerSlow(ctx context.Context, strings []string) <-chan string {
	outChannel := make(chan string)

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

func transformToLowerSlow(ctx context.Context, values <-chan string) <-chan string {
	outChannel := make(chan string)

	go func() {
		defer close(outChannel)

		select {
		case <-ctx.Done():
			return
		case s, ok := <-values:
			if ok {
				time.Sleep(time.Millisecond * 800)

				res := ""
				for _, char := range s {
					res += string(unicode.ToLower(char))

				}
				outChannel <- res
			} else {
				return
			}
		}
	}()

	return outChannel
}

func transformToTitleSlow(ctx context.Context, values <-chan string) <-chan string {
	outChannel := make(chan string)

	go func() {
		defer close(outChannel)

		select {
		case <-ctx.Done():
			return
		case s, ok := <-values:
			if ok {
				time.Sleep(time.Millisecond * 800)
				res := ""
				for i, char := range s {
					if i == 0 {
						res += string(unicode.ToTitle(char))
					} else {
						res += string(char)
					}
				}
				outChannel <- res
			} else {
				return
			}
		}
	}()

	return outChannel
}

func mergeStringChansSlow(ctx context.Context, cs ...<-chan string) <-chan string {
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
