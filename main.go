package main

import (
	"crypto/aes"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/logrusorgru/aurora/v3"
)

var (
	myfile   *os.FileInfo
	e        error
	dropTime int64
	delay    float64
	name     string
	searches int64
	config   map[string]interface{}
)

func main() {
	offsetRec := AutoOffset()
	fmt.Println(aurora.Sprintf(aurora.Bold(aurora.White(("\n%v valid out of %v accounts - %v can Prename.\n"))), aurora.Bold(aurora.Red(accountsVerify)), aurora.Bold(aurora.Red(len(txtSlice))), aurora.Bold(aurora.Red(accountsCanNamechange))))

	fmt.Println(aurora.Sprintf(aurora.Bold(aurora.White(("\nPossible Offset: %v | if your using GCS and you have more then 6 try %v | If your name has more then 400 searches try %v...\n"))), aurora.Red(math.Round(offsetRec-10)), aurora.Red(math.Round(offsetRec-16)), aurora.Red(math.Round(offsetRec+13))))

	if accountsCanNamechange == 0 {
		fmt.Println(aurora.Bold(aurora.White("No useable accounts..")))
		os.Exit(0)
	}

	if len(os.Args) > 1 {
		name = os.Args[1]
		delayConv := os.Args[2]
		delay, _ = strconv.ParseFloat(delayConv, 64)

		fmt.Println(aurora.Sprintf(aurora.Bold(aurora.White((" Name: %v\nDelay: %v"))), aurora.Bold(aurora.Red(name)), aurora.Bold(aurora.Red(delay))))
	} else {

		fmt.Println(aurora.Sprintf(aurora.White(aurora.Bold("Name: "))))
		fmt.Print(aurora.SlowBlink(aurora.BrightRed(">> ")))
		fmt.Scanln(&name)

		fmt.Println(aurora.Sprintf(aurora.White(aurora.Bold("Offset: "))))
		fmt.Print(aurora.SlowBlink(aurora.BrightRed(">> ")))
		fmt.Scanln(&delay)
	}

	switch config[`ManualBearer`] {
	case true:
		var amount int
		fmt.Println(aurora.Sprintf(aurora.White(aurora.Bold("Amount of accounts (more then 1 is recommended for gc): "))))
		fmt.Print((aurora.BrightRed(">> ")))
		fmt.Scan(&amount)

		for i := 0; i < amount; i++ {
			fmt.Println(aurora.Sprintf(aurora.White(aurora.Bold("Bearer: "))))
			fmt.Print((aurora.BrightRed(">> ")))
			fmt.Scan(&bearer)

			approvedBearer = append(approvedBearer, bearer)
		}
	}

	switch config[`UseApi`] {
	case false:
		if len(os.Args) > 3 {
			dropTimeConv := os.Args[3]
			dropTime, _ = strconv.ParseInt(dropTimeConv, 10, 64)

			fmt.Print("\n")
		} else {
			fmt.Println(aurora.Sprintf(aurora.White(aurora.Bold("\n! Droptime %v: [https://www.epochconverter.com]")), aurora.Bold(aurora.BrightBlack("[UNIX]"))))
			fmt.Print(aurora.SlowBlink(aurora.BrightRed(">> ")))
			fmt.Scan(&dropTime)
			fmt.Print("\n")
		}
		checkVer()
	case true:
		dropTime, _ = getDroptime(name)
		if dropTime < int64(10000) {
			if len(os.Args) > 3 {
				dropTimeConv := os.Args[3]
				dropTime, _ = strconv.ParseInt(dropTimeConv, 10, 64)
			} else {
				fmt.Println(aurora.Sprintf(aurora.White(aurora.Bold("\n! Droptime %v: [https://www.epochconverter.com]")), aurora.Bold(aurora.BrightBlack("[UNIX]"))))
				fmt.Print(aurora.SlowBlink(aurora.BrightRed(">> ")))
				fmt.Scan(&dropTime)
				fmt.Print("\n")
			}
		}

		//testHttp()

		checkVer()
	}

	time.Sleep(5 * time.Second)

	fmt.Println(aurora.Sprintf(aurora.White(aurora.Bold("\nPress enter to quit..."))))
	fmt.Print(aurora.Sprintf(aurora.Red(aurora.Bold(">> "))))
	fmt.Scanf("h")
	os.Exit(0)
}

func checkVer() {

	sendTimes, recvTimes, statusCodes := gcReq()

	for _, sends := range sendTimes {
		fmt.Println(aurora.Sprintf(aurora.White(aurora.Bold("[%v] Sent Req @ %v")), aurora.Bold(aurora.Underline(aurora.Red(name))), aurora.White(aurora.Bold(formatTime(sends)))))
	}
	fmt.Println("")
	for i, recvs := range recvTimes {
		if statusCodes[i] != `200` {
			fmt.Println(aurora.Sprintf(aurora.White(aurora.Bold("[%v] Recv Req @ %v")), aurora.Bold(aurora.Underline(aurora.Red(string(statusCodes[i])))), aurora.White(aurora.Bold(formatTime(recvs)))))
		} else if statusCodes[i] == `200` {
			fmt.Println(aurora.Sprintf(aurora.White(aurora.Bold("[%v] Succesfully sniped %v onto account %v @ %v")), aurora.Bold(aurora.Underline(aurora.Green(statusCode[i]))), aurora.Bold(aurora.Underline(aurora.Red(name))), aurora.White(aurora.Bold(emailGot)), aurora.White(aurora.Bold(formatTime(recvs)))))
		}
	}

	for _, status := range statusCodes {
		if status == `200` {
			sendInfo()
		}
	}
}

func EncryptAES(key []byte, plaintext string) string {

	values := empty(plaintext)

	switch values {
	case true:
		fmt.Print(aurora.Bold(aurora.Red(("✗ "))))
		log.Println(aurora.Bold(aurora.White("Please check your config.json and enter a Key")))
	}

	c, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println(err)
	}

	// allocate space for ciphered data
	out := make([]byte, len(plaintext))

	// encrypt
	c.Encrypt(out, []byte(plaintext))
	// return hex string
	return hex.EncodeToString(out)
}
