package util

import "time"

func Chan2IntervalChan(in chan string, duration time.Duration) chan string {
	out := make(chan string)
	ticker := time.NewTicker(duration)
	message := ""
	go func() {
		defer close(out)
		for {
			select {
			case chunk, ok := <-in:
				if !ok {
					out <- message
					return
				}
				message += chunk
			case <-ticker.C:
				out <- message
				message = ""
			}
		}
	}()
	return out
}

func TakeNLastItems[T any](arr []T, n int) []T {
	if len(arr) <= n {
		return arr
	}
	return arr[len(arr)-n:]
}
