```
BenchmarkSlow-8   	      61	  21066890 ns/op	20218002 B/op	  189842 allocs/op
BenchmarkFast-8   	     202	   6466109 ns/op	 1880555 B/op	   52402 allocs/op
```
# Изменения  
Использование регулярных выражений было не обязательно, они были заменены на функции из пакета strings.  
# Mem  
## top  
```
Showing nodes accounting for 515.22MB, 91.41% of 563.63MB total
Dropped 23 nodes (cum <= 2.82MB)
Showing top 10 nodes out of 43
      flat  flat%   sum%        cum   cum%
  111.01MB 19.70% 19.70%   111.01MB 19.70%  encoding/json.unquote (inline)
   80.02MB 14.20% 33.89%    80.02MB 14.20%  reflect.mapassign_faststr0
   79.18MB 14.05% 47.94%   531.22MB 94.25%  hw3.FastSearch
   56.50MB 10.02% 57.97%    56.50MB 10.02%  encoding/json.(*decodeState).literalStore
      47MB  8.34% 66.30%   408.54MB 72.48%  encoding/json.(*decodeState).object
   41.50MB  7.36% 73.67%    41.50MB  7.36%  reflect.New
      36MB  6.39% 80.06%   451.54MB 80.11%  encoding/json.Unmarshal
   29.50MB  5.23% 85.29%   159.01MB 28.21%  encoding/json.(*decodeState).arrayInterface
   18.50MB  3.28% 88.57%   129.51MB 22.98%  encoding/json.(*decodeState).literalInterface
      16MB  2.84% 91.41%    53.50MB  9.49%  reflect.cvtBytesString
```
## list  
```
Total: 563.63MB
ROUTINE ======================== hw3.FastSearch in /home/salta/projects/go-hw/03/fast.go
   79.18MB   531.22MB (flat, cum) 94.25% of Total
         .          .     13:func FastSearch(out io.Writer) {
         .          .     14:	/*
         .          .     15:		!!! !!! !!!
         .          .     16:		обратите внимание - в задании обязательно нужен отчет
         .          .     17:		делать его лучше в самом начале, когда вы видите уже узкие места, но еще не оптимизировалм их
         .          .     18:		так же обратите внимание на команду в параметром -http
         .          .     19:		перечитайте еще раз задание
         .          .     20:		!!! !!! !!!
         .          .     21:	*/
         .          .     22:	// SlowSearch(out)
         .          .     23:	file, err := os.OpenFile("./data/users.txt", os.O_RDONLY, 0)
         .          .     24:	if err != nil {
         .          .     25:		panic(err)
         .          .     26:	}
         .          .     27:
         .          .     28:	seenBrowsers := []string{}
         .          .     29:	uniqueBrowsers := 0
         .          .     30:	foundUsers := ""
         .          .     31:
         .          .     32:	scanner := bufio.NewScanner(file)
         .          .     33:	count := -1
         .          .     34:	for scanner.Scan() {
         .          .     35:		count++
         .          .     36:		line := scanner.Bytes()
   15.50MB    15.50MB     37:		user := make(map[string]interface{})
         .          .     38:
         .   448.04MB     39:		err = json.Unmarshal(line, &user)
         .          .     40:		if err != nil {
         .          .     41:			panic(err)
         .          .     42:		}
         .          .     43:
         .          .     44:		isAndroid := false
         .          .     45:		isMSIE := false
         .          .     46:
         .          .     47:		browsers, ok := user["browsers"].([]interface{})
         .          .     48:		if !ok {
         .          .     49:			continue
         .          .     50:		}
         .          .     51:
         .          .     52:		for _, browserRaw := range browsers {
         .          .     53:			browser, ok := browserRaw.(string)
         .          .     54:			if !ok {
         .          .     55:				continue
         .          .     56:			}
         .          .     57:			if strings.Contains(browser, "Android") {
         .          .     58:				isAndroid = true
         .          .     59:				notSeenBefore := true
         .          .     60:				for _, item := range seenBrowsers {
         .          .     61:					if item == browser {
         .          .     62:						notSeenBefore = false
         .          .     63:					}
         .          .     64:				}
         .          .     65:				if notSeenBefore {
         .          .     66:					seenBrowsers = append(seenBrowsers, browser)
         .          .     67:					uniqueBrowsers++
         .          .     68:				}
         .          .     69:			}
         .          .     70:		}
         .          .     71:
         .          .     72:		for _, browserRaw := range browsers {
         .          .     73:			browser, ok := browserRaw.(string)
         .          .     74:			if !ok {
         .          .     75:				continue
         .          .     76:			}
         .          .     77:			if strings.Contains(browser, "MSIE") {
         .          .     78:				isMSIE = true
         .          .     79:				notSeenBefore := true
         .          .     80:				for _, item := range seenBrowsers {
         .          .     81:					if item == browser {
         .          .     82:						notSeenBefore = false
         .          .     83:					}
         .          .     84:				}
         .          .     85:				if notSeenBefore {
  512.50kB   512.50kB     86:					seenBrowsers = append(seenBrowsers, browser)
         .          .     87:					uniqueBrowsers++
         .          .     88:				}
         .          .     89:			}
         .          .     90:		}
         .          .     91:
         .          .     92:		if !(isAndroid && isMSIE) {
         .          .     93:			continue
         .          .     94:		}
         .        1MB     95:		email := strings.Replace(user["email"].(string), "@", " [at] ", 1)
   62.67MB    64.67MB     96:		foundUsers += fmt.Sprintf("[%d] %s <%s>\n", count, user["name"], email)
         .          .     97:	}
         .          .     98:
         .          .     99:	if err := scanner.Err(); err != nil {
         .          .    100:		fmt.Printf("error reading file: %s\n", err)
         .          .    101:	}
         .          .    102:
  514.38kB     1.51MB    103:	fmt.Fprintln(out, "found users:\n"+foundUsers)
         .          .    104:	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
         .          .    105:}
```
[mem](./mem.svg)  
# CPU  
## top  
```
Showing nodes accounting for 1140ms, 60.64% of 1880ms total
Showing top 10 nodes out of 124
      flat  flat%   sum%        cum   cum%
     320ms 17.02% 17.02%      460ms 24.47%  encoding/json.checkValid
     220ms 11.70% 28.72%      220ms 11.70%  encoding/json.unquoteBytes
     120ms  6.38% 35.11%      370ms 19.68%  runtime.mallocgc
     110ms  5.85% 40.96%      110ms  5.85%  encoding/json.(*decodeState).rescanLiteral
     100ms  5.32% 46.28%      100ms  5.32%  runtime.nextFreeFast (inline)
      70ms  3.72% 50.00%       70ms  3.72%  encoding/json.stateInString
      50ms  2.66% 52.66%     1820ms 96.81%  hw3.FastSearch
      50ms  2.66% 55.32%       50ms  2.66%  indexbytebody
      50ms  2.66% 57.98%       50ms  2.66%  runtime.memclrNoHeapPointers
      50ms  2.66% 60.64%      190ms 10.11%  runtime.slicebytetostring
```
## list  
```
Total: 1.88s
ROUTINE ======================== hw3.FastSearch in /home/salta/projects/go-hw/03/fast.go
      50ms      1.82s (flat, cum) 96.81% of Total
         .          .     13:func FastSearch(out io.Writer) {
         .          .     14:	/*
         .          .     15:		!!! !!! !!!
         .          .     16:		обратите внимание - в задании обязательно нужен отчет
         .          .     17:		делать его лучше в самом начале, когда вы видите уже узкие места, но еще не оптимизировалм их
         .          .     18:		так же обратите внимание на команду в параметром -http
         .          .     19:		перечитайте еще раз задание
         .          .     20:		!!! !!! !!!
         .          .     21:	*/
         .          .     22:	// SlowSearch(out)
         .          .     23:	file, err := os.OpenFile("./data/users.txt", os.O_RDONLY, 0)
         .          .     24:	if err != nil {
         .          .     25:		panic(err)
         .          .     26:	}
         .          .     27:
         .          .     28:	seenBrowsers := []string{}
         .          .     29:	uniqueBrowsers := 0
         .          .     30:	foundUsers := ""
         .          .     31:
         .          .     32:	scanner := bufio.NewScanner(file)
         .          .     33:	count := -1
         .       40ms     34:	for scanner.Scan() {
         .          .     35:		count++
         .          .     36:		line := scanner.Bytes()
         .       10ms     37:		user := make(map[string]interface{})
         .          .     38:
         .      1.57s     39:		err = json.Unmarshal(line, &user)
         .          .     40:		if err != nil {
         .          .     41:			panic(err)
         .          .     42:		}
         .          .     43:
         .          .     44:		isAndroid := false
         .          .     45:		isMSIE := false
         .          .     46:
      10ms       10ms     47:		browsers, ok := user["browsers"].([]interface{})
         .          .     48:		if !ok {
         .          .     49:			continue
         .          .     50:		}
         .          .     51:
      10ms       10ms     52:		for _, browserRaw := range browsers {
         .          .     53:			browser, ok := browserRaw.(string)
         .          .     54:			if !ok {
         .          .     55:				continue
         .          .     56:			}
         .       50ms     57:			if strings.Contains(browser, "Android") {
         .          .     58:				isAndroid = true
         .          .     59:				notSeenBefore := true
         .          .     60:				for _, item := range seenBrowsers {
         .          .     61:					if item == browser {
         .          .     62:						notSeenBefore = false
         .          .     63:					}
         .          .     64:				}
         .          .     65:				if notSeenBefore {
         .          .     66:					seenBrowsers = append(seenBrowsers, browser)
         .          .     67:					uniqueBrowsers++
         .          .     68:				}
         .          .     69:			}
         .          .     70:		}
         .          .     71:
         .          .     72:		for _, browserRaw := range browsers {
         .          .     73:			browser, ok := browserRaw.(string)
         .          .     74:			if !ok {
         .          .     75:				continue
         .          .     76:			}
         .       60ms     77:			if strings.Contains(browser, "MSIE") {
         .          .     78:				isMSIE = true
         .          .     79:				notSeenBefore := true
         .          .     80:				for _, item := range seenBrowsers {
         .       10ms     81:					if item == browser {
         .          .     82:						notSeenBefore = false
         .          .     83:					}
         .          .     84:				}
         .          .     85:				if notSeenBefore {
         .          .     86:					seenBrowsers = append(seenBrowsers, browser)
         .          .     87:					uniqueBrowsers++
         .          .     88:				}
         .          .     89:			}
         .          .     90:		}
         .          .     91:
      20ms       20ms     92:		if !(isAndroid && isMSIE) {
         .          .     93:			continue
         .          .     94:		}
         .          .     95:		email := strings.Replace(user["email"].(string), "@", " [at] ", 1)
         .       20ms     96:		foundUsers += fmt.Sprintf("[%d] %s <%s>\n", count, user["name"], email)
         .          .     97:	}
         .          .     98:
         .          .     99:	if err := scanner.Err(); err != nil {
         .          .    100:		fmt.Printf("error reading file: %s\n", err)
         .          .    101:	}
         .          .    102:
         .       10ms    103:	fmt.Fprintln(out, "found users:\n"+foundUsers)
      10ms       10ms    104:	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
         .          .    105:}
```
[cpu](./cpu.svg)  
# Выводы
Сейчас больше всего на производительность влияет парсинг JSON.  