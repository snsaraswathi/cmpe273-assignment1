package main

import (
    "bufio"
    "os"
    "fmt"
    "log"
    "net/http"
    "github.com/gorilla/rpc/json"
    "bytes"
)

type Request1 struct {
	StockSymbolAndPercentage string `json:"stockSymbolAndPercentage"`
	Budget float32 `json:"budget"`
}

type Reply struct {
	TradeId int
	StrOut string
	UnvestedAmount float32
}

type Request2 struct {
	TradeId int `json:"tradeId"`
}

type Response2 struct {
	Stocks string 
	CurrentMarketValue float32
	UnvestedAmount float32
}

func main() {
	var option int
	var money float32
	var input string
	var errChar string
	bio := bufio.NewReader(os.Stdin)

	fmt.Println("Choose one of the following options: ")
	fmt.Println("1. Buy stocks")
	fmt.Println("2. View profile")
	fmt.Println("Enter 1 or 2")
	line,_,_ := bio.ReadLine()
	fmt.Sscanf(string(line),"%d %s", &option, &errChar)
	if(errChar != "") {
		fmt.Println("Invalid input")
		os.Exit(1)	
	}

if(option == 1) {
//Get input
	var a Request1
        var reply Reply
	
	fmt.Println("Enter string input")
	line,_,_ := bio.ReadLine()
	fmt.Sscanf(string(line),"%s", &input)
    	fmt.Println("Enter budget")
	line2,_ ,_ := bio.ReadLine()
	fmt.Sscanf(string(line2),"%f", &money)
	a.Budget= money
        a.StockSymbolAndPercentage=input
        buf, _ := json.EncodeClientRequest("StockMarket.BuyStocks", &a)
        resp, err := http.Post("http://localhost:10000/rpc", "application/json", bytes.NewBuffer(buf))

	if err != nil {
		fmt.Println("Error occurred", err)
        } 
        defer resp.Body.Close()

        err = json.DecodeClientResponse(resp.Body, &reply)

	if err != nil {
                log.Fatal("Error occurred.. ",err)
        }

	fmt.Println("Your trade Id is: ",  reply.TradeId)
        fmt.Println("Summary of your trade: ", reply.StrOut)
        fmt.Println("Remaining balance: ", reply.UnvestedAmount) 

} else if(option == 2) {
	var tid int
	var b Request2
        var reply2 Response2
	
	fmt.Println("Enter your trade id")
	line,_ ,_ := bio.ReadLine()
        fmt.Sscanf(string(line),"%d %s", &tid, &errChar)
        if(errChar != "") {
                  fmt.Println("Invalid input")
                  os.Exit(1)
        }
        b.TradeId=tid
        buf, _ := json.EncodeClientRequest("StockMarket.CheckPortfolio", &b)

        resp2, err := http.Post("http://localhost:10000/rpc", "application/json", bytes.NewBuffer(buf))
        
        if err != nil {
                fmt.Println("Error:", err)
        }

	defer resp2.Body.Close()

        err = json.DecodeClientResponse(resp2.Body, &reply2)

        if err != nil {
                log.Fatal(err)
        }

        fmt.Println("Stocks bought: ",  reply2.Stocks)
        fmt.Println("CurrentMarketValue: ", reply2.CurrentMarketValue)
        fmt.Println("Remaining balance: ", reply2.UnvestedAmount)

} else {
	fmt.Println("Invalid input")
}
}
