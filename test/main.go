package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	// 1. Gọi 3 nhà vận chuyển (chạy song song)
	ch1 := fetchPrice("GHTK", 3000*time.Millisecond, 30000)
	ch2 := fetchPrice("SHOPEE", 4000*time.Millisecond, 35000)
	ch3 := fetchPrice("Ahamove", 1000*time.Millisecond, 37000) // Cho Ahamove nhanh nhất (1s)

	// 2. Gom 3 channel thành 1 channel tổng 'out'
	out := fanIn(ch1, ch2, ch3)

	// 3. Đọc dữ liệu từ channel tổng (Bên nào trả về trước sẽ in ra trước!)
	for msg := range out {
		fmt.Println(msg)
	}
}

func fetchPrice(provider string, duration time.Duration, price int) <-chan string {
	ch := make(chan string)

	go func() {
		time.Sleep(duration)
		ch <- fmt.Sprintf("Gia van chuyen cua %s la %d", provider, price)
		close(ch)
	}()
	return ch
}

func fanIn(input1, input2, input3 <-chan string) <-chan string {
	out := make(chan string)
	var wg sync.WaitGroup

	// Tạo 1 helper function để đọc data từ 1 channel bất kỳ rồi đẩy sang 'out'
	multiplex := func(c <-chan string) {
		defer wg.Done()
		for msg := range c { // Vòng range tự kết thúc khi channel 'c' bị close
			out <- msg
		}
	}

	wg.Add(3)
	// Bật 3 Goroutines ngầm gom data từ 3 channel song song nhau
	go multiplex(input1)
	go multiplex(input2)
	go multiplex(input3)

	// Bật 1 Goroutine riêng chuyên đứng chờ 3 đứa trên xong để close(out)
	go func() {
		wg.Wait()
		close(out) // Close để vòng lặp range ở main biết đường dừng lại
	}()

	return out
}
