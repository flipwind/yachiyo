package main

import "time"

func main() {
	println("Hello Yachiyo!")

	// I won't leave you till the end.
	go func() {
		for {
			time.Sleep(100 * time.Second)
		}
	}()

	<-make(chan struct{})
}