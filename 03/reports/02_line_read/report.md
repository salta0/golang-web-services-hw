```
BenchmarkSlow-8   	      52	  20709741 ns/op	20205381 B/op	  189837 allocs/op
BenchmarkFast-8   	      62	  21206550 ns/op	15501808 B/op	  188737 allocs/op
```
# Изменениея
Файл загружается не целиком, а по линии.
# Mem
## top
```
Showing nodes accounting for 1016.27MB, 91.49% of 1110.85MB total
Dropped 20 nodes (cum <= 5.55MB)
Showing top 10 nodes out of 50
      flat  flat%   sum%        cum   cum%
  521.68MB 46.96% 46.96%   521.68MB 46.96%  regexp/syntax.(*compiler).inst (inline)
  203.02MB 18.28% 65.24%   203.02MB 18.28%  regexp/syntax.(*parser).newRegexp (inline)
   96.51MB  8.69% 73.93%   955.73MB 86.04%  regexp.compile
   64.01MB  5.76% 79.69%   296.03MB 26.65%  regexp/syntax.parse
      35MB  3.15% 82.84%       35MB  3.15%  encoding/json.unquote (inline)
   28.50MB  2.57% 85.41%       53MB  4.77%  regexp/syntax.(*compiler).init (inline)
   23.01MB  2.07% 87.48%    23.01MB  2.07%  reflect.mapassign_faststr0
      18MB  1.62% 89.10%       18MB  1.62%  regexp/syntax.(*parser).maybeConcat
   13.53MB  1.22% 90.32%  1071.02MB 96.41%  hw3.FastSearch
      13MB  1.17% 91.49%       13MB  1.17%  encoding/json.(*decodeState).literalStor
```
## list
```
Total: 1.08GB
ROUTINE ======================== hw3.FastSearch in /home/salta/projects/go-hw/03/fast.go
   13.53MB     1.05GB (flat, cum) 96.41% of Total
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
         .          .     28:	r := regexp.MustCompile("@")
         .          .     29:	seenBrowsers := []string{}
         .          .     30:	uniqueBrowsers := 0
         .          .     31:	foundUsers := ""
         .          .     32:
         .          .     33:	scanner := bufio.NewScanner(file)
         .          .     34:	i := -1
         .          .     35:	// Loop through the file and read each line
         .      514kB     36:	for scanner.Scan() {
         .          .     37:		i++
         .          .     38:		line := scanner.Bytes() // Get the line as a string
         .          .     39:		// fmt.Println(line)
       3MB        3MB     40:		user := make(map[string]interface{})
         .          .     41:		// fmt.Printf("%v %v\n", err, line)
         .   116.01MB     42:		err = json.Unmarshal(line, &user)
         .          .     43:		if err != nil {
         .          .     44:			panic(err)
         .          .     45:		}
         .          .     46:		// users = append(users, user)
         .          .     47:		isAndroid := false
         .          .     48:		isMSIE := false
         .          .     49:
         .          .     50:		browsers, ok := user["browsers"].([]interface{})
         .          .     51:		if !ok {
         .          .     52:			// log.Println("cant cast browsers")
         .          .     53:			continue
         .          .     54:		}
         .          .     55:
         .          .     56:		for _, browserRaw := range browsers {
         .          .     57:			browser, ok := browserRaw.(string)
         .          .     58:			if !ok {
         .          .     59:				// log.Println("cant cast browser to string")
         .          .     60:				continue
         .          .     61:			}
         .   569.80MB     62:			if ok, err := regexp.MatchString("Android", browser); ok && err == nil {
         .          .     63:				isAndroid = true
         .          .     64:				notSeenBefore := true
         .          .     65:				for _, item := range seenBrowsers {
         .          .     66:					if item == browser {
         .          .     67:						notSeenBefore = false
         .          .     68:					}
         .          .     69:				}
         .          .     70:				if notSeenBefore {
         .          .     71:					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
         .          .     72:					seenBrowsers = append(seenBrowsers, browser)
         .          .     73:					uniqueBrowsers++
         .          .     74:				}
         .          .     75:			}
         .          .     76:		}
         .          .     77:
         .          .     78:		for _, browserRaw := range browsers {
         .          .     79:			browser, ok := browserRaw.(string)
         .          .     80:			if !ok {
         .          .     81:				// log.Println("cant cast browser to string")
         .          .     82:				continue
         .          .     83:			}
         .   369.67MB     84:			if ok, err := regexp.MatchString("MSIE", browser); ok && err == nil {
         .          .     85:				isMSIE = true
         .          .     86:				notSeenBefore := true
         .          .     87:				for _, item := range seenBrowsers {
         .          .     88:					if item == browser {
         .          .     89:						notSeenBefore = false
         .          .     90:					}
         .          .     91:				}
         .          .     92:				if notSeenBefore {
         .          .     93:					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
         .          .     94:					seenBrowsers = append(seenBrowsers, browser)
         .          .     95:					uniqueBrowsers++
         .          .     96:				}
         .          .     97:			}
         .          .     98:		}
         .          .     99:
         .          .    100:		if !(isAndroid && isMSIE) {
         .          .    101:			continue
         .          .    102:		}
         .          .    103:
         .          .    104:		// log.Println("Android and MSIE user:", user["name"], user["email"])
         .   512.01kB    105:		email := r.ReplaceAllString(user["email"].(string), " [at] ")
   10.53MB    11.03MB    106:		foundUsers += fmt.Sprintf("[%d] %s <%s>\n", i, user["name"], email)
         .          .    107:
         .          .    108:	}
         .          .    109:
         .          .    110:	if err := scanner.Err(); err != nil {
         .          .    111:		fmt.Printf("error reading file: %s\n", err)
         .          .    112:	}
         .          .    113:
         .   514.38kB    114:	fmt.Fprintln(out, "found users:\n"+foundUsers)
         .          .    115:	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
         .          .    116:}
```
[mem](./mem.svg)
# CPU
## top
```
Showing nodes accounting for 560ms, 36.60% of 1530ms total
Showing top 10 nodes out of 210
      flat  flat%   sum%        cum   cum%
     100ms  6.54%  6.54%      450ms 29.41%  runtime.mallocgc
      80ms  5.23% 11.76%       80ms  5.23%  runtime.memclrNoHeapPointers
      70ms  4.58% 16.34%       70ms  4.58%  encoding/json.unquoteBytes
      60ms  3.92% 20.26%       60ms  3.92%  runtime.nextFreeFast (inline)
      60ms  3.92% 24.18%       80ms  5.23%  runtime.scanobject
      50ms  3.27% 27.45%       50ms  3.27%  runtime.scanblock
      40ms  2.61% 30.07%       40ms  2.61%  encoding/json.stateInString
      40ms  2.61% 32.68%      140ms  9.15%  regexp/syntax.(*parser).push
      30ms  1.96% 34.64%       60ms  3.92%  reflect.Value.SetMapIndex
      30ms  1.96% 36.60%      270ms 17.65%  regexp/syntax.(*compiler).inst
```
## list
```
Total: 1.53s
ROUTINE ======================== hw3.FastSearch in /home/salta/projects/go-hw/03/fast.go
      20ms      1.27s (flat, cum) 83.01% of Total
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
         .          .     28:	r := regexp.MustCompile("@")
         .          .     29:	seenBrowsers := []string{}
         .          .     30:	uniqueBrowsers := 0
         .          .     31:	foundUsers := ""
         .          .     32:
         .          .     33:	scanner := bufio.NewScanner(file)
         .          .     34:	i := -1
         .          .     35:	// Loop through the file and read each line
         .       10ms     36:	for scanner.Scan() {
         .          .     37:		i++
         .          .     38:		line := scanner.Bytes() // Get the line as a string
         .          .     39:		// fmt.Println(line)
         .       20ms     40:		user := make(map[string]interface{})
         .          .     41:		// fmt.Printf("%v %v\n", err, line)
         .      460ms     42:		err = json.Unmarshal(line, &user)
         .          .     43:		if err != nil {
         .          .     44:			panic(err)
         .          .     45:		}
         .          .     46:		// users = append(users, user)
         .          .     47:		isAndroid := false
         .          .     48:		isMSIE := false
         .          .     49:
         .          .     50:		browsers, ok := user["browsers"].([]interface{})
         .          .     51:		if !ok {
         .          .     52:			// log.Println("cant cast browsers")
         .          .     53:			continue
         .          .     54:		}
         .          .     55:
         .          .     56:		for _, browserRaw := range browsers {
         .          .     57:			browser, ok := browserRaw.(string)
         .          .     58:			if !ok {
         .          .     59:				// log.Println("cant cast browser to string")
         .          .     60:				continue
         .          .     61:			}
         .      410ms     62:			if ok, err := regexp.MatchString("Android", browser); ok && err == nil {
         .          .     63:				isAndroid = true
         .          .     64:				notSeenBefore := true
      10ms       10ms     65:				for _, item := range seenBrowsers {
         .          .     66:					if item == browser {
         .          .     67:						notSeenBefore = false
         .          .     68:					}
         .          .     69:				}
         .          .     70:				if notSeenBefore {
         .          .     71:					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
         .          .     72:					seenBrowsers = append(seenBrowsers, browser)
         .          .     73:					uniqueBrowsers++
         .          .     74:				}
         .          .     75:			}
         .          .     76:		}
         .          .     77:
         .          .     78:		for _, browserRaw := range browsers {
         .          .     79:			browser, ok := browserRaw.(string)
         .          .     80:			if !ok {
         .          .     81:				// log.Println("cant cast browser to string")
         .          .     82:				continue
         .          .     83:			}
         .      350ms     84:			if ok, err := regexp.MatchString("MSIE", browser); ok && err == nil {
         .          .     85:				isMSIE = true
         .          .     86:				notSeenBefore := true
         .          .     87:				for _, item := range seenBrowsers {
         .          .     88:					if item == browser {
         .          .     89:						notSeenBefore = false
         .          .     90:					}
         .          .     91:				}
         .          .     92:				if notSeenBefore {
         .          .     93:					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
         .          .     94:					seenBrowsers = append(seenBrowsers, browser)
         .          .     95:					uniqueBrowsers++
         .          .     96:				}
         .          .     97:			}
         .          .     98:		}
         .          .     99:
         .          .    100:		if !(isAndroid && isMSIE) {
         .          .    101:			continue
         .          .    102:		}
         .          .    103:
         .          .    104:		// log.Println("Android and MSIE user:", user["name"], user["email"])
         .          .    105:		email := r.ReplaceAllString(user["email"].(string), " [at] ")
         .          .    106:		foundUsers += fmt.Sprintf("[%d] %s <%s>\n", i, user["name"], email)
         .          .    107:
         .          .    108:	}
         .          .    109:
         .          .    110:	if err := scanner.Err(); err != nil {
         .          .    111:		fmt.Printf("error reading file: %s\n", err)
         .          .    112:	}
         .          .    113:
         .          .    114:	fmt.Fprintln(out, "found users:\n"+foundUsers)
      10ms       10ms    115:	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
         .          .    116:}
```
[cpu](./cpu.svg)
# Вывод  
Сейчас больше всего на производительность влияет работа с регулярными выражениями. Необходимо посмотреть, где их можно заменить на более быстрые варианты или оптимизировать.