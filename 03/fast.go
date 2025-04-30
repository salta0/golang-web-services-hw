package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	easyjson "github.com/mailru/easyjson"
)

type User struct {
	Name     string
	Browsers []string
	Email    string
}

// вам надо написать более быструю оптимальную этой функции
func FastSearch(out io.Writer) {
	/*
		!!! !!! !!!
		обратите внимание - в задании обязательно нужен отчет
		делать его лучше в самом начале, когда вы видите уже узкие места, но еще не оптимизировалм их
		так же обратите внимание на команду в параметром -http
		перечитайте еще раз задание
		!!! !!! !!!
	*/
	// SlowSearch(out)
	file, err := os.OpenFile("./data/users.txt", os.O_RDONLY, 0)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	seenBrowsers := make(map[string]int)
	scanner := bufio.NewScanner(file)
	fmt.Fprintf(out, "found users:\n")
	count := 0
	for scanner.Scan() {
		line := scanner.Bytes()
		user := &User{
			Name:     "",
			Browsers: make([]string, 0, 5),
			Email:    "",
		}
		err := easyjson.Unmarshal(line, user)
		if err != nil {
			panic(err)
		}

		isAndroid := false
		isMSIE := false

		browsers := user.Browsers
		for _, browser := range browsers {
			if strings.Contains(browser, "Android") {
				isAndroid = true
				seenBrowsers[browser] = 1
			} else if strings.Contains(browser, "MSIE") {
				isMSIE = true
				seenBrowsers[browser] = 1
			}
		}

		if isAndroid && isMSIE {
			email := strings.Replace(user.Email, "@", " [at] ", 1)
			fmt.Fprintf(out, "[%d] %s <%s>\n", count, user.Name, email)
		}
		count++
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("error reading file: %s\n", err)
	}

	fmt.Fprintln(out, "\nTotal unique browsers", len(seenBrowsers))
}
