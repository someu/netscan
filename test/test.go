package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	go handle(ctx, 2000*time.Millisecond)
	select {
	case <-ctx.Done():
		fmt.Println("main", ctx.Err())
	}
	time.Sleep(time.Second * 5)
}

func handle(ctx context.Context, duration time.Duration) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("handle", ctx.Err())
			return
		case <-time.After(duration):
			fmt.Println("process request with", duration)
			return
		default:
			break
		}
		fmt.Println("handle")
	}

}
