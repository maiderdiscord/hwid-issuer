package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/denisbrodbeck/machineid"
	"github.com/keygen-sh/keygen-go"
)

func init() {
	keygen.Account = "54de58e6-e0a0-46dd-8e7a-c4b310c2ebfc"
	keygen.Product = "785c9307-19a0-49b7-8992-caf236e307d0"
}

func main() {
	fmt.Print("ライセンスキー >> ")

	scanner := bufio.NewScanner(os.Stdin)
	if os.Getenv("DEBUG") == "true" {
		scanner.Scan()
	}
	licenseKey := scanner.Text()

	token, err := getToken(licenseKey)
	if err != nil {
		log.Printf("アクティベーショントークンの取得に失敗しました: %+v", err)
		pressEnter()
		os.Exit(1)
	}

	keygen.Token = token

	hwid, err := machineid.ID()
	if err != nil {
		log.Fatal(err)
	}

	license, err := keygen.Validate(hwid)
	if err != nil {
		switch err {
		case keygen.ErrLicenseNotActivated:
			_, err = license.Activate(hwid)
			if err != nil {
				log.Printf("アクティベート中にエラーが発生しました: %+v", err)
				pressEnter()
				os.Exit(1)
			}

		case keygen.ErrLicenseExpired:
			fmt.Println("ライセンスが失効しています。")
			pressEnter()
			os.Exit(1)

		default:
			fmt.Println("無効なライセンスです。")
			pressEnter()
			os.Exit(1)
		}
	}

	fmt.Println("アクティベートに成功しました。")
	pressEnter()
}

func pressEnter() {
	fmt.Print("続けるにはエンターキーを押してください。")
	bufio.NewScanner(os.Stdin).Scan()
}

type graphQLResponse struct {
	Data getTokenResponse `json:"data"`
}

type getTokenResponse struct {
	ActivateToken string `json:"activateToken"`
}

func getToken(licenseKey string) (string, error) {
	query := fmt.Sprintf(`{activateToken(licenseKey: \"%s\")}`, licenseKey)
	body := fmt.Sprintf(`{"query": "%s"}`, query)

	resp, err := http.Post("https://api.maider.net/query", "application/json", bytes.NewReader([]byte(body)))
	if err != nil {
		return "", err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	data := new(graphQLResponse)
	if err := json.Unmarshal(respBody, data); err != nil {
		return "", err
	}

	return data.Data.ActivateToken, nil
}
