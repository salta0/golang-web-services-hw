# Mem
## top
```
Showing nodes accounting for 2058.96MB, 90.88% of 2265.51MB total
Dropped 27 nodes (cum <= 11.33MB)
Showing top 10 nodes out of 51
      flat  flat%   sum%        cum   cum%
  840.80MB 37.11% 37.11%   840.80MB 37.11%  regexp/syntax.(*compiler).inst (inline)
  350.72MB 15.48% 52.59%   350.72MB 15.48%  io.ReadAll
  295.53MB 13.04% 65.64%   295.53MB 13.04%  regexp/syntax.(*parser).newRegexp (inline)
  166.02MB  7.33% 72.97%  1545.87MB 68.23%  regexp.compile
  128.02MB  5.65% 78.62%   472.05MB 20.84%  regexp/syntax.parse
   74.91MB  3.31% 81.92%  1122.31MB 49.54%  hw3.FastSearch
   73.41MB  3.24% 85.16%  1142.71MB 50.44%  hw3.SlowSearch
   46.51MB  2.05% 87.22%    46.51MB  2.05%  encoding/json.unquote (inline)
      46MB  2.03% 89.25%       91MB  4.02%  regexp/syntax.(*compiler).init (inline)
   37.04MB  1.63% 90.88%    37.04MB  1.63%  regexp.(*bitState).reset
(pprof) list FastSearch
Total: 2.21GB
```
# list
```
Total: 2.21GB
ROUTINE ======================== hw3.FastSearch in /home/salta/projects/go-hw/03/fast.go
   74.91MB     1.10GB (flat, cum) 49.54% of Total
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
         .          .     23:	file, err := os.Open(filePath)
         .          .     24:	if err != nil {
         .          .     25:		panic(err)
         .          .     26:	}
         .          .     27:
         .   166.70MB     28:	fileContents, err := io.ReadAll(file)
         .          .     29:	if err != nil {
         .          .     30:		panic(err)
         .          .     31:	}
         .          .     32:
         .          .     33:	r := regexp.MustCompile("@")
         .          .     34:	seenBrowsers := []string{}
         .          .     35:	uniqueBrowsers := 0
         .          .     36:	foundUsers := ""
         .          .     37:
   35.36MB    36.38MB     38:	lines := strings.Split(string(fileContents), "\n")
         .          .     39:
         .          .     40:	users := make([]map[string]interface{}, 0)
         .          .     41:	for _, line := range lines {
    3.50MB     3.50MB     42:		user := make(map[string]interface{})
         .          .     43:		// fmt.Printf("%v %v\n", err, line)
   23.51MB   108.02MB     44:		err := json.Unmarshal([]byte(line), &user)
         .          .     45:		if err != nil {
         .          .     46:			panic(err)
         .          .     47:		}
  515.32kB   515.32kB     48:		users = append(users, user)
         .          .     49:	}
         .          .     50:
         .          .     51:	for i, user := range users {
         .          .     52:
         .          .     53:		isAndroid := false
         .          .     54:		isMSIE := false
         .          .     55:
         .          .     56:		browsers, ok := user["browsers"].([]interface{})
         .          .     57:		if !ok {
         .          .     58:			// log.Println("cant cast browsers")
         .          .     59:			continue
         .          .     60:		}
         .          .     61:
         .          .     62:		for _, browserRaw := range browsers {
         .          .     63:			browser, ok := browserRaw.(string)
         .          .     64:			if !ok {
         .          .     65:				// log.Println("cant cast browser to string")
         .          .     66:				continue
         .          .     67:			}
         .   485.44MB     68:			if ok, err := regexp.MatchString("Android", browser); ok && err == nil {
         .          .     69:				isAndroid = true
         .          .     70:				notSeenBefore := true
         .          .     71:				for _, item := range seenBrowsers {
         .          .     72:					if item == browser {
         .          .     73:						notSeenBefore = false
         .          .     74:					}
         .          .     75:				}
         .          .     76:				if notSeenBefore {
         .          .     77:					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
         .          .     78:					seenBrowsers = append(seenBrowsers, browser)
         .          .     79:					uniqueBrowsers++
         .          .     80:				}
         .          .     81:			}
         .          .     82:		}
         .          .     83:
         .          .     84:		for _, browserRaw := range browsers {
         .          .     85:			browser, ok := browserRaw.(string)
         .          .     86:			if !ok {
         .          .     87:				// log.Println("cant cast browser to string")
         .          .     88:				continue
         .          .     89:			}
         .   308.22MB     90:			if ok, err := regexp.MatchString("MSIE", browser); ok && err == nil {
         .          .     91:				isMSIE = true
         .          .     92:				notSeenBefore := true
         .          .     93:				for _, item := range seenBrowsers {
         .          .     94:					if item == browser {
         .          .     95:						notSeenBefore = false
         .          .     96:					}
         .          .     97:				}
         .          .     98:				if notSeenBefore {
         .          .     99:					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
  512.50kB   512.50kB    100:					seenBrowsers = append(seenBrowsers, browser)
         .          .    101:					uniqueBrowsers++
         .          .    102:				}
         .          .    103:			}
         .          .    104:		}
         .          .    105:
         .          .    106:		if !(isAndroid && isMSIE) {
         .          .    107:			continue
         .          .    108:		}
         .          .    109:
         .          .    110:		// log.Println("Android and MSIE user:", user["name"], user["email"])
         .          .    111:		email := r.ReplaceAllString(user["email"].(string), " [at] ")
   11.03MB    12.54MB    112:		foundUsers += fmt.Sprintf("[%d] %s <%s>\n", i, user["name"], email)
         .          .    113:	}
         .          .    114:
  514.38kB   514.38kB    115:	fmt.Fprintln(out, "found users:\n"+foundUsers)
         .          .    116:	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
         .          .    117:}

```

[mem](./mem.svg)
# CPU
## top
```
Showing nodes accounting for 1800ms, 40.27% of 4470ms total
Dropped 119 nodes (cum <= 22.35ms)
Showing top 10 nodes out of 172
      flat  flat%   sum%        cum   cum%
     310ms  6.94%  6.94%     1230ms 27.52%  runtime.mallocgc
     250ms  5.59% 12.53%      250ms  5.59%  runtime.(*mspan).base (inline)
     200ms  4.47% 17.00%      200ms  4.47%  runtime.memclrNoHeapPointers
     180ms  4.03% 21.03%     1000ms 22.37%  runtime.scanobject
     170ms  3.80% 24.83%      230ms  5.15%  runtime.findObject
     170ms  3.80% 28.64%      170ms  3.80%  runtime.memmove
     170ms  3.80% 32.44%      170ms  3.80%  runtime.nextFreeFast (inline)
     140ms  3.13% 35.57%      220ms  4.92%  encoding/json.checkValid
     110ms  2.46% 38.03%      110ms  2.46%  runtime.pageIndexOf (inline)
     100ms  2.24% 40.27%      100ms  2.24%  runtime.(*gcBits).bitp (inline)
```
## list
```
Total: 4.47s
ROUTINE ======================== hw3.FastSearch in /home/salta/projects/go-hw/03/fast.go
      10ms      1.34s (flat, cum) 29.98% of Total
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
         .          .     23:	file, err := os.Open(filePath)
         .          .     24:	if err != nil {
         .          .     25:		panic(err)
         .          .     26:	}
         .          .     27:
         .       60ms     28:	fileContents, err := io.ReadAll(file)
         .          .     29:	if err != nil {
         .          .     30:		panic(err)
         .          .     31:	}
         .          .     32:
         .          .     33:	r := regexp.MustCompile("@")
         .          .     34:	seenBrowsers := []string{}
         .          .     35:	uniqueBrowsers := 0
         .          .     36:	foundUsers := ""
         .          .     37:
         .       30ms     38:	lines := strings.Split(string(fileContents), "\n")
         .          .     39:
         .          .     40:	users := make([]map[string]interface{}, 0)
         .          .     41:	for _, line := range lines {
         .          .     42:		user := make(map[string]interface{})
         .          .     43:		// fmt.Printf("%v %v\n", err, line)
         .      430ms     44:		err := json.Unmarshal([]byte(line), &user)
         .          .     45:		if err != nil {
         .          .     46:			panic(err)
         .          .     47:		}
         .          .     48:		users = append(users, user)
         .          .     49:	}
         .          .     50:
         .          .     51:	for i, user := range users {
         .          .     52:
         .          .     53:		isAndroid := false
         .          .     54:		isMSIE := false
         .          .     55:
         .          .     56:		browsers, ok := user["browsers"].([]interface{})
         .          .     57:		if !ok {
         .          .     58:			// log.Println("cant cast browsers")
         .          .     59:			continue
         .          .     60:		}
         .          .     61:
         .          .     62:		for _, browserRaw := range browsers {
         .          .     63:			browser, ok := browserRaw.(string)
         .          .     64:			if !ok {
         .          .     65:				// log.Println("cant cast browser to string")
         .          .     66:				continue
         .          .     67:			}
         .      530ms     68:			if ok, err := regexp.MatchString("Android", browser); ok && err == nil {
         .          .     69:				isAndroid = true
         .          .     70:				notSeenBefore := true
         .          .     71:				for _, item := range seenBrowsers {
         .          .     72:					if item == browser {
         .          .     73:						notSeenBefore = false
         .          .     74:					}
         .          .     75:				}
         .          .     76:				if notSeenBefore {
         .          .     77:					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
         .          .     78:					seenBrowsers = append(seenBrowsers, browser)
         .          .     79:					uniqueBrowsers++
         .          .     80:				}
         .          .     81:			}
         .          .     82:		}
         .          .     83:
         .          .     84:		for _, browserRaw := range browsers {
         .          .     85:			browser, ok := browserRaw.(string)
         .          .     86:			if !ok {
         .          .     87:				// log.Println("cant cast browser to string")
         .          .     88:				continue
         .          .     89:			}
         .      270ms     90:			if ok, err := regexp.MatchString("MSIE", browser); ok && err == nil {
         .          .     91:				isMSIE = true
         .          .     92:				notSeenBefore := true
      10ms       10ms     93:				for _, item := range seenBrowsers {
         .          .     94:					if item == browser {
         .          .     95:						notSeenBefore = false
         .          .     96:					}
         .          .     97:				}
         .          .     98:				if notSeenBefore {
         .          .     99:					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
         .          .    100:					seenBrowsers = append(seenBrowsers, browser)
         .          .    101:					uniqueBrowsers++
         .          .    102:				}
         .          .    103:			}
         .          .    104:		}
         .          .    105:
         .          .    106:		if !(isAndroid && isMSIE) {
         .          .    107:			continue
         .          .    108:		}
         .          .    109:
         .          .    110:		// log.Println("Android and MSIE user:", user["name"], user["email"])
         .          .    111:		email := r.ReplaceAllString(user["email"].(string), " [at] ")
         .       10ms    112:		foundUsers += fmt.Sprintf("[%d] %s <%s>\n", i, user["name"], email)
         .          .    113:	}
         .          .    114:
         .          .    115:	fmt.Fprintln(out, "found users:\n"+foundUsers)
         .          .    116:	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
         .          .    117:}

```
[cpu](./cpu.svg)

# Report
Можно заметить большие утечки памяти при работе с регулярными выражениями и чтение данных.  