package main

import (

	"fmt"
	"strconv"
	"sync"
	"sort"
	"strings"
)

type msruct struct {
	i int         // порядковый номер хэша
	str string   // хэш
}

func ExecutePipeline(hashSignJobs ...job) {
	wg := &sync.WaitGroup{}
	in := make(chan interface{})

	for _, jobItem := range hashSignJobs {
		wg.Add(1)
		out := make(chan interface{})
		go func(jobFunc job, in, out chan interface{}, wg *sync.WaitGroup) {
			defer wg.Done()
			defer close(out)
			jobFunc(in, out)
		}(jobItem, in, out, wg)
		in = out
	}
	wg.Wait()
}

func SingleHash(in,out chan interface{}) {
	wg := &sync.WaitGroup{}
	for data := range in {
		wg.Add(1)
		data := fmt.Sprintf("%v", data)
		hash_ := md_5(data)
		hash := <- hash_
		crc32_md5 := ""

		go func(data string, hash string) {
			defer wg.Done()
			crc32_ := getcrt(data)
			crc32_md5 =DataSignerCrc32(hash)
			crc32 := <- crc32_
			out <- crc32 + "~" + crc32_md5
		}(data, hash)
	}
	 wg.Wait()
}

func  MultiHash(in,out chan interface{}) {

	    ch2 := make(chan string)
		wg2 := &sync.WaitGroup{}
  		for data := range in {
			wg := &sync.WaitGroup{}
			ch1 := make(chan msruct)
			wg2.Add(1)
			wg.Add(6)
			data := fmt.Sprintf("%v", data)
			MultiHashResults := ""

		for i := 0; i < 6; i++ {
			go multi(data,MultiHashResults,i,ch1, wg)
		}
			go multisort(ch1,ch2,wg2)
		    go func(wg *sync.WaitGroup, ch chan msruct) {
				defer close(ch)
				wg.Wait()
			}(wg, ch1)
	}

	go func(wg *sync.WaitGroup, ch chan string) {
		defer close(ch)
		wg.Wait()
	}(wg2, ch2)

	for x := range ch2{
		out <- x
	}
}

func getcrt(data string) chan string {
	result := make(chan string, 1)
	go func(out chan<- string) {
		out <- DataSignerCrc32(data)
	}(result)

	return result
}


func md_5(data string) chan string {
	result := make(chan string, 1)
	go func(out chan<- string) {
		out <- DataSignerMd5(data)
	}(result)

	return result
}


	func multi(data, MultiHashResults string, i int, ch1 chan msruct,wg *sync.WaitGroup) {
		defer wg.Done()
		crc32 := DataSignerCrc32(strconv.Itoa(i) + data)
		MultiHashResults = MultiHashResults + crc32
		ch1 <- msruct{i,MultiHashResults}
}

func multisort(ch1 chan msruct,out chan string, wg *sync.WaitGroup)  {
	defer wg.Done()
    finish := map[int]string{}
	var data []int

	for x := range ch1{
		 finish[x.i] = x.str
		 data = append(data, x.i)
	 }
	sort.Ints(data)

	var results []string
	for i := range data {
		results = append(results, finish[i])
	}
	out <- strings.Join(results, "")
}



func CombineResults(in, out chan interface{}) {
	var res []string

	for result := range in {
		res = append(res, result.(string))
	}

	sort.Strings(res)
	out <-  strings.Join(res, "_")

}



