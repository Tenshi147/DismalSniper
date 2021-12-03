package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/logrusorgru/aurora/v3"
)

var redirect string
var bearerReturn string

// i will not be explaining how msauth works, to tedious and will just take to long! feel free to
// use this website for information, https://mojang-api-docs.netlify.app/authentication/msa.html
// also expect this code to get a redo! i have many plans for this.

func oAuth2(email string, password string) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("Got error while creating cookie jar %s", err.Error())
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			redirect = req.URL.String()
			return nil
		},
		Jar: jar,
	}

	resp, err := http.NewRequest("GET", "https://login.live.com/oauth20_authorize.srf?client_id=000000004C12AE6F&redirect_uri=https://login.live.com/oauth20_desktop.srf&scope=service::user.auth.xboxlive.com::MBI_SSL&display=touch&response_type=token&locale=en", nil)
	if err != nil {
		fmt.Print(err)
		return
	}

	resp.Header.Set("User-Agent", "Mozilla/5.0 (XboxReplay; XboxLiveAuth/3.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36")

	response, err := client.Do(resp)
	if err != nil {
		panic(err)
	}

	jar.Cookies(resp.URL)

	bodyByte, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Print(err)
		return
	}

	myString := string(bodyByte[:])

	search1 := regexp.MustCompile(`value="(.*?)"`)
	search3 := regexp.MustCompile(`urlPost:'(.+?)'`)

	value := search1.FindAllStringSubmatch(myString, -1)[0][1]
	urlPost := search3.FindAllStringSubmatch(myString, -1)[0][1]

	emailEncode := url.QueryEscape(email)
	passwordEncode := url.QueryEscape(password)

	body := []byte(fmt.Sprintf("login=%v&loginfmt=%v&passwd=%v&PPFT=%v", emailEncode, emailEncode, passwordEncode, value))

	req, err := http.NewRequest("POST", urlPost, bytes.NewReader(body))

	if err != nil {
		fmt.Print(err)
		return
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (XboxReplay; XboxLiveAuth/3.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36")

	_, err = client.Do(req)
	if err != nil {
		panic(err)
	}

	respBytes, err := ioutil.ReadAll(response.Body)

	if err != nil {
		panic(err)
	}

	if strings.Contains(string(respBytes), "Sign in to") {
		fmt.Println(aurora.Green("invalid credentials"))
		fmt.Println(aurora.Sprintf(aurora.White(aurora.Bold("\nPress enter to quit..."))))
		fmt.Print(aurora.Sprintf(aurora.Red(aurora.Bold(">>"))))
		fmt.Scanf("h")
		bearer = "Invalid"
	}

	if strings.Contains(string(respBytes), "Help us protect your account") {
		fmt.Println(aurora.Green("2fa is enabled, which is not supported now"))
		fmt.Println(aurora.Sprintf(aurora.White(aurora.Bold("\nPress enter to quit..."))))
		fmt.Print(aurora.Sprintf(aurora.Red(aurora.Bold(">>"))))
		fmt.Scanf("h")
		bearer = "Invalid"
	}

	if !strings.Contains(redirect, "access_token") || redirect == urlPost {
		bearer = "Invalid"
	}

	if bearer != "Invalid" {
		gatherAuth()
	}

}

type bearerMs struct {
	Bearer string `json:"access_token"`
}

func gatherAuth() {

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Renegotiation: tls.RenegotiateFreelyAsClient,
			},
		},
	}

	splitBear := strings.Split(redirect, "#")[1]

	splitValues := strings.Split(splitBear, "&")

	//refresh_token := strings.Split(splitValues[4], "=")[1]
	access_token := strings.Split(splitValues[0], "=")[1]
	//expires_in := strings.Split(splitValues[2], "=")[1]

	body := []byte(`{"Properties": {"AuthMethod": "RPS", "SiteName": "user.auth.xboxlive.com", "RpsTicket": "` + access_token + `"}, "RelyingParty": "http://auth.xboxlive.com", "TokenType": "JWT"}`)
	post, err := http.NewRequest("POST", "https://user.auth.xboxlive.com/user/authenticate", bytes.NewBuffer(body))
	if err != nil {
		fmt.Print(err)
	}

	post.Header.Set("Content-Type", "application/json")
	post.Header.Set("Accept", "application/json")

	bodyRP, err := client.Do(post)
	if err != nil {
		fmt.Print(err)
	}

	rpBody, err := ioutil.ReadAll(bodyRP.Body)
	if err != nil {
		fmt.Print(err)
	}

	Token := extractValue(string(rpBody), "Token")
	uhs := extractValue(string(rpBody), "uhs")

	payload := []byte(`{"Properties": {"SandboxId": "RETAIL", "UserTokens": ["` + Token + `"]}, "RelyingParty": "rp://api.minecraftservices.com/", "TokenType": "JWT"}`)
	xstsPost, err := http.NewRequest("POST", "https://xsts.auth.xboxlive.com/xsts/authorize", bytes.NewBuffer(payload))
	if err != nil {
		fmt.Print(err)
	}

	xstsPost.Header.Set("Content-Type", "application/json")
	xstsPost.Header.Set("Accept", "application/json")

	bodyXS, err := client.Do(xstsPost)
	if err != nil {
		fmt.Print(err)
	}

	xsBody, err := ioutil.ReadAll(bodyXS.Body)
	if err != nil {
		fmt.Print(err)
	}

	switch bodyXS.StatusCode {
	case 401:
		switch !strings.Contains(string(xsBody), "XErr") {
		case !strings.Contains(string(xsBody), "2148916238"):
			fmt.Println("account belongs to someone under 18 and needs to be added to a family")

			fmt.Println(aurora.Sprintf(aurora.White(aurora.Bold("\nPress enter to quit..."))))
			fmt.Print(aurora.Sprintf(aurora.Red(aurora.Bold(">> "))))
			fmt.Scanf("h")
			os.Exit(0)
		case !strings.Contains(string(xsBody), "2148916233"):
			fmt.Println("account has no Xbox account, you must sign up for one first")

			fmt.Println(aurora.Sprintf(aurora.White(aurora.Bold("\nPress enter to quit..."))))
			fmt.Print(aurora.Sprintf(aurora.Red(aurora.Bold(">> "))))
			fmt.Scanf("h")
			os.Exit(0)
		}
	}

	xsToken := extractValue(string(xsBody), "Token")

	mcBearer := []byte(`{"identityToken" : "XBL3.0 x=` + uhs + `;` + xsToken + `", "ensureLegacyEnabled" : true}`)
	mcBPOST, err := http.NewRequest("POST", "https://api.minecraftservices.com/authentication/login_with_xbox", bytes.NewBuffer(mcBearer))
	if err != nil {
		fmt.Print(err)
	}

	mcBPOST.Header.Set("Content-Type", "application/json")

	bodyBearer, err := client.Do(mcBPOST)
	if err != nil {
		fmt.Print(err)
	}

	bearerValue, err := ioutil.ReadAll(bodyBearer.Body)
	if err != nil {
		fmt.Print(err)
	}
	var bearerMS bearerMs
	json.Unmarshal(bearerValue, &bearerMS)

	bearer = bearerMS.Bearer

}

func extractValue(body string, key string) string {
	keystr := "\"" + key + "\":[^,;\\]}]*"
	r, _ := regexp.Compile(keystr)
	match := r.FindString(body)
	keyValMatch := strings.Split(match, ":")
	return strings.ReplaceAll(keyValMatch[1], "\"", "")
}

func extractValueLogin(body string, key string) string {
	if strings.Contains(body, "accessToken") == true {
		keystr := "\"" + key + "\":[^,;\\]}]*"
		r, err := regexp.Compile(keystr)
		if err != nil {
			fmt.Println(err)
		} else {
			match := r.FindString(body)
			keyValMatch := strings.Split(match, ":")
			bearerReturn = strings.ReplaceAll(keyValMatch[1], "\"", "")
		}
	} else {
		bearerReturn = "Invalid"
	}

	return bearerReturn
}
