package main

import (
	"log"
)

func pointer1() {
	arr := []int{1, 2, 3, 4, 5}

	arr2 := append(arr[1:2], 55)
	arr2[0] = 3
	log.Printf("POINTER1 %v", arr)
	log.Printf("POINTER1 %v", arr2)

}

func pointer() {
	type test1 struct {
		text *string
	}

	type test2 struct {
		text **string
	}

	var txt *string = nil
	t1 := test1{
		text: txt,
	}
	t2 := test2{
		text: &txt,
	}

	log.Printf("t1: %v, t2: %v", t1.text, *t2.text)

	t := "Hello World"
	txt = &t
	t1.text = txt

	// strings.ReplaceAll(*t1.text, "asd", "")
	log.Printf("t1: %v, t2: %v", t1.text, **t2.text)

}

func main() {

	log.Printf("[IRONY] START")

	// pointer1()
	addr2 := 192<<16 | 168<<8

	log.Printf("%d.%d", addr2>>16, addr2>>8&0xff)

	// tChan := make(chan int)
	// done := make(chan interface{})
	// var wait sync.WaitGroup

	// go func() {
	// 	for i := 0; i < 5; i++ {
	// 		wait.Add(1)
	// 		// func() { tChan <- i }()
	// 		log.Printf("BEFORE SEND %d", i)
	// 		tChan <- i
	// 		log.Printf("AFTER SEND %d", i)
	// 	}
	// 	// time.Sleep(1 * time.Second)
	// 	close(tChan)

	// 	log.Printf("INPUT 완료")
	// }()

	// go func() {
	// 	time.Sleep(10 * time.Millisecond)
	// 	// for {
	// 	// 	v, exist := <-tChan
	// 	// 	if !exist {
	// 	// 		break
	// 	// 	} else {
	// 	// 		wait.Done()
	// 	// 		log.Printf("GO1 %d", v)
	// 	// 	}
	// 	// 	// time.Sleep(1 * time.Second)
	// 	// 	time.Sleep(100 * time.Millisecond)
	// 	// }
	// 	wait.Wait()
	// 	close(done)
	// 	log.Printf("GO1 종료")
	// }()
	// // go func() {
	// // 	for {
	// // 		v, exist := <-tChan
	// // 		if !exist {
	// // 			break
	// // 		} else {
	// // 			log.Printf("GO2 %d", v)
	// // 		}
	// // 	}
	// // 	log.Printf("GO2 종료")
	// // }()

	// <-done

	log.Printf("[IRONY] E N D")

}
