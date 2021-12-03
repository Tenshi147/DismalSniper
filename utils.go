package main

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/logrusorgru/aurora/v3"
)

type jsonVars struct {
	UNIX        int64   `json:"UNIX,omitempty"`
	Error       string  `json:"error,omitempty"`
	AccessToken *string `json:"accessToken,omitempty"`
}

var (
	txtSlice              []string
	access                jsonVars
	bytesToSend           []byte
	bearer                string
	outputSlice           []string
	accountType           string
	auth                  map[string]interface{}
	emails                []string
	accountsVerify        int64
	accountsCanNamechange int64
	approvedBearer        []string
)

func init() {

	webhookvar, _ := ioutil.ReadFile("config.json")
	json.Unmarshal(webhookvar, &config)

	header := `
·▄▄▄▄  ▪  .▄▄ · • ▌ ▄ ·.  ▄▄▄· ▄▄▌  
██▪ ██ ██ ▐█ ▀. ·██ ▐███▪▐█ ▀█ ██•  
▐█· ▐█▌▐█·▄▀▀▀█▄▐█ ▌▐▌▐█·▄█▀▀█ ██▪  
██. ██ ▐█▌▐█▄▪▐███ ██▌▐█▌▐█ ▪▐▌▐█▌▐▌
▀▀▀▀▀• ▀▀▀ ▀▀▀▀ ▀▀  █▪▀▀▀ ▀  ▀ .▀▀▀ `

	for _, char := range []string{"•", "·", ".", "▪"} {
		header = strings.ReplaceAll(header, char, aurora.Sprintf(aurora.Faint(aurora.White("%v")), char))
	}
	for _, char := range []string{"█", "▄", "▌", "▀", "▌", "▀"} {
		header = strings.ReplaceAll(header, char, aurora.Sprintf(aurora.Bold(aurora.BrightWhite(("%v"))), char))
	}
	for _, char := range []string{"▐"} {
		header = strings.ReplaceAll(header, char, aurora.Sprintf(aurora.Faint(aurora.White(("%v"))), char))
	}

	// Credit to kqzzs method of seperately coloring text. ^^^^^

	fmt.Print(header)

	fmt.Printf(aurora.Sprintf(aurora.Bold(aurora.White(`
Ver: %v

`)), aurora.Bold(aurora.BrightBlack("2.0.0"))))

	switch config[`ManualBearer`] {
	case false:
		file, err := os.Open("accounts.txt")
		i := 0
		if err == nil {
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.Contains(line, "\n") {
					line = line[:len(line)-1]
				}
				if strings.Contains(line, "\r") {
					line = line[:len(line)-1]
				}
				txtSlice = append(txtSlice, scanner.Text())
				i++
			}

		}
	case true:
		accountType = "Manual Input"
	}

	if len(txtSlice) == 0 {
		fmt.Println(aurora.Sprintf(aurora.White(aurora.Bold("No accounts in accounts.txt"))))
	}
}

// Credit to trinity, some of this code is from goSnipe.

func init() {
	i := 0

	switch config[`ManualBearer`] {
	case false:
		for ew, input := range txtSlice {
			i++
			time.Sleep(time.Second)
			splitLogin := strings.Split(input, ":")

			oAuth2(splitLogin[0], splitLogin[1])

			if strings.Contains(bearer, "Invalid") {

			} else {

				check := isGc()

				if check == "Giftcard" {
					approvedBearer = append(approvedBearer, bearer)
					emails = append(emails, splitLogin[0])
					fmt.Println(aurora.Sprintf(aurora.Bold(aurora.White("Succesfully Authenticated %v | Giftcard Account")), aurora.Bold(aurora.Red(splitLogin[0]))))
					accountsVerify++
					accountType = "Giftcard"
					if i == 3 && ew+1 != len(txtSlice) {
						dropStamp := time.Now()
						delDroptime := dropStamp.Add(+time.Second * 60)
						for {
							fmt.Printf(aurora.Sprintf(aurora.White(aurora.Bold("Sleeping for | %v This is to not get Microsoft rate limited!   \r")), aurora.Bold(aurora.Red(time.Until(delDroptime).Round(time.Second).Seconds()))))
							if time.Until(delDroptime) < 0*time.Second {
								i = 0
								break
							}
						}
					}
					continue
				} else {
					fmt.Println(aurora.Sprintf(aurora.Bold(aurora.White("Unsuccesfully Authenticated %v | Giftcard Account")), aurora.Bold(aurora.Red(splitLogin[0]))))
				}
			}
			continue
		}
	case true:
	}
}

// - checks if an acc is gc or not.

func isGc() string {
	conn, err := tls.Dial("tcp", "api.minecraftservices.com"+":443", nil)
	if err != nil {
		fmt.Print(err)
	}

	fmt.Fprintln(conn, "GET /minecraft/profile/namechange HTTP/1.1\r\nHost: api.minecraftservices.com\r\nUser-Agent: Dismal/1.0\r\nAuthorization: Bearer "+bearer+"\r\n\r\n")

	e := make([]byte, 12)
	_, err = conn.Read(e)
	if err != nil {
		fmt.Print(err)
	}

	// checks status codes..
	switch string(e[9:12]) {
	case `404`:
		accountType = "Giftcard"
		accountsCanNamechange++
	}

	return accountType
}

func getDroptime(name string) (int64, string) {

	// makes a new get request
	resp, _ := http.NewRequest("GET",
		"https://api.star.shopping/droptime/"+name,
		nil)

	resp.Header.Set("user-agent", "Sniper")

	data, err := http.DefaultClient.Do(resp)
	if err != nil {
		fmt.Println(err)
	}

	defer data.Body.Close()

	// reads the body..
	dropTimeBytes, err := ioutil.ReadAll(data.Body)
	if err != nil {
		fmt.Print(err)
	}

	var f jsonVars
	json.Unmarshal(dropTimeBytes, &f)

	return f.UNIX, f.Error
}

func empty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

// - skinChange is used to change the accs skin IF u get the name. -

func skinChange() {
	// makes a post body with the contents url, variant and content type these are used in the http request.
	postBody, err := json.Marshal(map[string]string{
		"Content-Type": "application/json",
		"url":          config[`ChangeSkinLink`].(string),
		"variant":      config[`SkinModel`].(string),
	})
	if err != nil {
		fmt.Print(err)
	}
	// i turn the post body into bytes buffer which can be used for the body of resp..
	responseBody := bytes.NewBuffer(postBody)

	resp, err := http.NewRequest("POST", "https://api.minecraftservices.com/minecraft/profile/skins", responseBody)
	if err != nil {
		fmt.Print(err)
	}
	resp.Header.Set("Authorization", "bearer "+bearer)

	// sends req..

	skin, err := http.DefaultClient.Do(resp)
	if err != nil {
		fmt.Print(err)
	}

	// status checks for a valid status code..
	switch skin.StatusCode {
	case 200:
		fmt.Println(aurora.Sprintf(aurora.White(aurora.Bold("[%v] Succesfully Changed your skin")), aurora.Bold(aurora.Underline(aurora.Green(skin.StatusCode)))))
	case 401:
		fmt.Println(aurora.Sprintf(aurora.White(aurora.Bold("[%v] Failed skin change!")), aurora.Bold(aurora.Underline(aurora.Red(skin.StatusCode)))))
	}

}

func formatTime(t time.Time) string {
	return t.Format("15:04:05.00000")
}

var search int64

func sendInfo() {
	time.Sleep(time.Second)
	switch {
	case config[`ChangeskinOnSnipe`] == true:
		skinChange()
	}
	switch {
	case config[`UseWebhook`] == true:
		time.Sleep(5 * time.Millisecond)
		dismalWebhook()
		switch config[`UsePersonal`] {
		case true:
			personalWebhook()
		}
	}
}

func dismalWebhook() {
	id := config[`DiscordID`].(string)

	webhookLink := DecryptAES([]byte("XVlBzgbaiCMRXVlBzgbaiCMRXVlBzgba"), "ba5f43be0ecc8f8d1ebc38836f4f1cc9")
	webhookLink = webhookLink + "wu.herokuapp.com/webhook/" + name + "/" + accountType + "/" + id

	req, _ := http.NewRequest("GET", webhookLink, nil)

	res, _ := http.DefaultClient.Do(req)

	respData, _ := ioutil.ReadAll(res.Body)

	auth := make(map[string]interface{})

	json.Unmarshal(respData, &auth)

	fmt.Println(auth)

	if auth["value"].(bool) == true {
		fmt.Println(aurora.Sprintf(aurora.White(aurora.Bold("[%v] Sent Webhook!!")), aurora.Bold(aurora.Underline(aurora.Green(res.StatusCode)))))
	} else {
		fmt.Println(aurora.Sprintf(aurora.White(aurora.Bold("[%v] Couldnt send Webhook.. Check config?")), aurora.Bold(aurora.Underline(aurora.Red(res.StatusCode)))))
	}
}

func DecryptAES(key []byte, ct string) string {
	ciphertext, _ := hex.DecodeString(ct)

	c, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println(err)
	}

	pt := make([]byte, len(ciphertext))
	c.Decrypt(pt, ciphertext)

	s := string(pt[:])
	return s
}

func personalWebhook() {

	str := config[`PersonalWebhookBody`].(interface{})

	owo, _ := json.Marshal(str)

	twedad := string(owo)

	id := config[`DiscordID`].(string)
	webhookLink := config[`WebhookLink`].(string)

	twedad = strings.Replace(twedad, "{id}", fmt.Sprintf("<@%v>", id), 5)
	twedad = strings.Replace(twedad, "{name}", fmt.Sprintf("%v", name), 5)
	twedad = strings.Replace(twedad, "{accountType}", fmt.Sprintf("%v", accountType), 5)
	twedad = strings.Replace(twedad, "{searches}", fmt.Sprintf("%v", searches), 5)

	webhookReq, err := http.NewRequest("POST", webhookLink, bytes.NewReader([]byte(twedad)))
	if err != nil {
		fmt.Println(err)
	}
	webhookReq.Header.Set("Content-Type", "application/json")

	conn, err := http.DefaultClient.Do(webhookReq)
	if err != nil {
		fmt.Print(err)
	}

	if conn.StatusCode == 204 {
		fmt.Println(aurora.Sprintf(aurora.White(aurora.Bold("[%v] Sent Webhook!!")), aurora.Bold(aurora.Underline(aurora.Green(conn.StatusCode)))))
	} else if conn.StatusCode != 204 {
		fmt.Println(aurora.Sprintf(aurora.White(aurora.Bold("[%v] Couldnt send Webhook.. Check config?")), aurora.Bold(aurora.Underline(aurora.Red(conn.StatusCode)))))
	}
}

// - Used to calculate delay, some of it is accurate some isnt! never rely on recommended delay.. simply base ur delay off it. -

func AutoOffset() float64 {

	// makes the variables for the workflow.
	payload := []byte("GET /minecraft/profile/name/test HTTP/1.1\r\nHost: api.minecraftservices.com\r\nAuthorization: Bearer TestToken" + "\r\n")
	conn, _ := tls.Dial("tcp", "api.minecraftservices.com"+":443", nil)
	pingTimes := make([]float64, 10)
	var offset float64

	// loops 10 times, this writes a junk file to store junk data.. it writes the payload to conn and appends the time taken into pingTimes.
	for i := 0; i < 10; i++ {
		junk := make([]byte, 1000)
		conn.Write(payload)
		time1 := time.Now()
		conn.Write([]byte("\r\n"))
		conn.Read(junk)
		time2 := time.Since(time1)
		fmt.Println(aurora.Sprintf(aurora.Bold(aurora.White("Took - %s")), time2))
		pingTimes[i] = float64(time2.Milliseconds())

	}

	// calculates the sum and does the math.. / 10000 to get the decimal version of sum then i * 5100~ (u can also do 5000) but it
	// only times the decimal to get the non deciaml number Example: 57 (the delay recommendations are very similar to python delay scripts ive tested)
	sum := sum(pingTimes)

	offset = float64((sum / 10000)) * 5000

	return offset
}

func sum(array []float64) float64 {
	var sum1 float64 = 0
	for i := 0; i < 10; i++ {
		sum1 = sum1 + array[i]
	}
	return sum1
}
