package main

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

type threadHash struct {
	Pos string
	Res string
}

type byPos []threadHash

func (a byPos) Len() int           { return len(a) }
func (a byPos) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byPos) Less(i, j int) bool { return a[i].Pos < a[j].Pos }

func makeCrc32(out chan string, input string) {
	out <- DataSignerCrc32(input)
}

func makeSingleHash(wg *sync.WaitGroup, out chan interface{}, input, md5Hash string) {
	defer wg.Done()

	h32ResCh := make(chan string)
	h32Md5ResCh := make(chan string)

	go makeCrc32(h32ResCh, input)
	go makeCrc32(h32Md5ResCh, md5Hash)

	h32Res := <-h32ResCh
	h32Md5Res := <-h32Md5ResCh

	out <- (h32Res + "~" + h32Md5Res)
}

func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}

	for input := range in {
		wg.Add(1)
		strInput := fmt.Sprintf("%v", input)
		md5Res := DataSignerMd5(strInput)

		go makeSingleHash(wg, out, strInput, md5Res)
	}
	wg.Wait()
}

func makeSingleThreadHash(wg *sync.WaitGroup, out chan threadHash, pos, input string) {
	res := threadHash{
		Pos: pos,
		Res: DataSignerCrc32(pos + input),
	}
	out <- res
	wg.Done()
}

func combineThreadHash(in chan threadHash, out chan string) {
	threads := make([]threadHash, 6)
	for input := range in {
		threads = append(threads, input)
	}
	sort.Sort(byPos(threads))
	res := ""
	for _, el := range threads {
		res += el.Res
	}

	out <- res
}

func makeMultiHash(wg *sync.WaitGroup, out chan interface{}, input string) {
	defer wg.Done()

	singleThWg := &sync.WaitGroup{}
	singleThCh := make(chan threadHash)
	resultCh := make(chan string)

	for j := 0; j <= 5; j++ {
		singleThWg.Add(1)
		th := fmt.Sprintf("%v", j)
		go makeSingleThreadHash(singleThWg, singleThCh, th, input)
	}

	go combineThreadHash(singleThCh, resultCh)

	singleThWg.Wait()
	close(singleThCh)

	res := <-resultCh

	out <- res
}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for input := range in {
		strInput := fmt.Sprintf("%v", input)
		wg.Add(1)
		go makeMultiHash(wg, out, strInput)
	}
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	var res []string
	for input := range in {
		res = append(res, fmt.Sprintf("%v", input))
	}
	sort.Strings(res)

	out <- strings.Join(res, "_")
}

func executeJob(job job, wg *sync.WaitGroup, in, out chan interface{}) {
	defer wg.Done()

	job(in, out)
	close(out)
}

func ExecutePipeline(freeFlowJobs ...job) {
	wg := &sync.WaitGroup{}
	in := make(chan interface{})
	out := make(chan interface{})
	for _, job := range freeFlowJobs {
		wg.Add(1)
		go executeJob(job, wg, in, out)
		in = out
		out = make(chan interface{})
	}
	wg.Wait()
}
