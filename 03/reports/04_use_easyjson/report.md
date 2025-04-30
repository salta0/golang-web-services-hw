```
BenchmarkSlow-8   	      58	  20705007 ns/op	20233325 B/op	  189839 allocs/op
BenchmarkFast-8   	    1119	   1292341 ns/op	  567922 B/op	    7315 allocs/op
```
# Изменения  
Вместо стандартного парсера был использован easyjson. Так же были убраны лишние циклы.  
# Mem
## top
```
Showing nodes accounting for 662.84MB, 98.88% of 670.38MB total
Dropped 20 nodes (cum <= 3.35MB)
Showing top 10 nodes out of 35
      flat  flat%   sum%        cum   cum%
  507.55MB 75.71% 75.71%   507.55MB 75.71%  github.com/mailru/easyjson/jlexer.(*Lexer).String
  106.58MB 15.90% 91.61%   622.14MB 92.80%  hw3.FastSearch
   15.51MB  2.31% 93.92%    15.51MB  2.31%  regexp/syntax.(*compiler).inst (inline)
   10.50MB  1.57% 95.49%    10.50MB  1.57%  regexp/syntax.(*parser).newRegexp (inline)
    6.55MB  0.98% 96.46%     6.55MB  0.98%  io.ReadAll
    4.50MB  0.67% 97.14%     4.50MB  0.67%  strings.(*Builder).grow
    3.51MB  0.52% 97.66%     3.51MB  0.52%  bufio.(*Scanner).Scan
    3.50MB  0.52% 98.18%    33.51MB  5.00%  regexp.compile
    2.65MB  0.39% 98.58%    48.24MB  7.20%  hw3.SlowSearch
       2MB   0.3% 98.88%       13MB  1.94%  regexp/syntax.parse
```

## list  
```
Total: 670.38MB
ROUTINE ======================== hw3.FastSearch in /home/salta/projects/go-hw/03/fast.go
  106.58MB   622.14MB (flat, cum) 92.80% of Total
         .          .     20:func FastSearch(out io.Writer) {
         .          .     21:	/*
         .          .     22:		!!! !!! !!!
         .          .     23:		обратите внимание - в задании обязательно нужен отчет
         .          .     24:		делать его лучше в самом начале, когда вы видите уже узкие места, но еще не оптимизировалм их
         .          .     25:		так же обратите внимание на команду в параметром -http
         .          .     26:		перечитайте еще раз задание
         .          .     27:		!!! !!! !!!
         .          .     28:	*/
         .          .     29:	// SlowSearch(out)
         .          .     30:	file, err := os.OpenFile("./data/users.txt", os.O_RDONLY, 0)
         .          .     31:	if err != nil {
         .          .     32:		panic(err)
         .          .     33:	}
         .          .     34:	defer file.Close()
         .          .     35:
         .          .     36:	seenBrowsers := make(map[string]int)
         .          .     37:	scanner := bufio.NewScanner(file)
         .          .     38:	count := -1
         .          .     39:	fmt.Fprintf(out, "found users:\n")
         .          .     40:	for scanner.Scan() {
         .          .     41:		count++
         .     3.51MB     42:		line := scanner.Bytes()
         .          .     43:		user := &User{
         .          .     44:			Name:     "",
         .          .     45:			Browsers: make([]string, 0, 5),
         .          .     46:			Email:    "",
   87.51MB    87.51MB     47:		}
         .          .     48:		err := easyjson.Unmarshal(line, user)
         .          .     49:		if err != nil {
         .          .     50:			panic(err)
         .          .     51:		}
         .   507.55MB     52:
         .          .     53:		isAndroid := false
         .          .     54:		isMSIE := false
         .          .     55:
         .          .     56:		browsers := user.Browsers
         .          .     57:		for _, browser := range browsers {
         .          .     58:			if strings.Contains(browser, "Android") {
         .          .     59:				isAndroid = true
         .          .     60:				seenBrowsers[browser] = 1
         .          .     61:			} else if strings.Contains(browser, "MSIE") {
         .          .     62:				isMSIE = true
         .          .     63:				seenBrowsers[browser] = 1
         .          .     64:			}
    9.05MB     9.05MB     65:		}
         .          .     66:
         .          .     67:		if !(isAndroid && isMSIE) {
    6.52MB     6.52MB     68:			continue
         .          .     69:		}
         .          .     70:
         .          .     71:		email := strings.Replace(user.Email, "@", " [at] ", 1)
         .          .     72:		fmt.Fprintf(out, "[%d] %s <%s>\n", count, user.Name, email)
         .          .     73:	}
         .          .     74:
         .          .     75:	if err := scanner.Err(); err != nil {
         .     4.50MB     76:		fmt.Printf("error reading file: %s\n", err)
    3.50MB     3.50MB     77:	}
         .          .     78:
         .          .     79:	fmt.Fprintln(out, "\nTotal unique browsers", len(seenBrowsers))
         .          .     80:}
```
[mem](mem.svg)  
# CPU  
## top  
```
Showing nodes accounting for 950ms, 59.01% of 1610ms total
Showing top 10 nodes out of 146
      flat  flat%   sum%        cum   cum%
     360ms 22.36% 22.36%      360ms 22.36%  indexbytebody
     110ms  6.83% 29.19%      110ms  6.83%  runtime/internal/syscall.Syscall6
      90ms  5.59% 34.78%      330ms 20.50%  runtime.mallocgc
      90ms  5.59% 40.37%       90ms  5.59%  runtime.nextFreeFast (inline)
      80ms  4.97% 45.34%      300ms 18.63%  github.com/mailru/easyjson/jlexer.(*Lexer).FetchToken
      50ms  3.11% 48.45%       50ms  3.11%  indexbody
      50ms  3.11% 51.55%      260ms 16.15%  strings.Index
      40ms  2.48% 54.04%      120ms  7.45%  github.com/mailru/easyjson/jlexer.(*Lexer).unescapeStringToken
      40ms  2.48% 56.52%      190ms 11.80%  github.com/mailru/easyjson/jlexer.findStringLen
      40ms  2.48% 59.01%       90ms  5.59%  runtime.mapassign_faststr
```
## list  
```
Total: 1.61s
ROUTINE ======================== hw3.FastSearch in /home/salta/projects/go-hw/03/fast.go
      30ms      1.47s (flat, cum) 91.30% of Total
         .          .     20:func FastSearch(out io.Writer) {
         .          .     21:	/*
         .          .     22:		!!! !!! !!!
         .          .     23:		обратите внимание - в задании обязательно нужен отчет
         .          .     24:		делать его лучше в самом начале, когда вы видите уже узкие места, но еще не оптимизировалм их
         .          .     25:		так же обратите внимание на команду в параметром -http
         .          .     26:		перечитайте еще раз задание
         .          .     27:		!!! !!! !!!
         .          .     28:	*/
         .          .     29:	// SlowSearch(out)
         .          .     30:	file, err := os.OpenFile("./data/users.txt", os.O_RDONLY, 0)
         .          .     31:	if err != nil {
         .          .     32:		panic(err)
         .          .     33:	}
         .          .     34:	defer file.Close()
         .          .     35:
         .          .     36:	seenBrowsers := make(map[string]int)
         .          .     37:	scanner := bufio.NewScanner(file)
         .          .     38:	count := -1
         .          .     39:	fmt.Fprintf(out, "found users:\n")
         .          .     40:	for scanner.Scan() {
         .          .     41:		count++
         .          .     42:		line := scanner.Bytes()
         .          .     43:		user := &User{
         .      210ms     44:			Name:     "",
      10ms       10ms     45:			Browsers: make([]string, 0, 5),
         .          .     46:			Email:    "",
         .          .     47:		}
         .          .     48:		err := easyjson.Unmarshal(line, user)
         .      110ms     49:		if err != nil {
         .          .     50:			panic(err)
         .          .     51:		}
         .          .     52:
         .          .     53:		isAndroid := false
         .      740ms     54:		isMSIE := false
         .          .     55:
         .          .     56:		browsers := user.Browsers
         .          .     57:		for _, browser := range browsers {
         .          .     58:			if strings.Contains(browser, "Android") {
         .          .     59:				isAndroid = true
         .          .     60:				seenBrowsers[browser] = 1
         .          .     61:			} else if strings.Contains(browser, "MSIE") {
         .          .     62:				isMSIE = true
         .          .     63:				seenBrowsers[browser] = 1
      10ms       10ms     64:			}
         .      140ms     65:		}
         .          .     66:
         .       40ms     67:		if !(isAndroid && isMSIE) {
         .      120ms     68:			continue
         .          .     69:		}
         .       50ms     70:
         .          .     71:		email := strings.Replace(user.Email, "@", " [at] ", 1)
         .          .     72:		fmt.Fprintf(out, "[%d] %s <%s>\n", count, user.Name, email)
         .          .     73:	}
      10ms       10ms     74:
         .          .     75:	if err := scanner.Err(); err != nil {
         .          .     76:		fmt.Printf("error reading file: %s\n", err)
         .          .     77:	}
         .          .     78:
         .       20ms     79:	fmt.Fprintln(out, "\nTotal unique browsers", len(seenBrowsers))
         .          .     80:}
```  
[cpu](cpu.svg)  
# Вывод  
По сравнению с предыдущим шагом программа стала работать быстрее, но занимает больший объем памяти. 