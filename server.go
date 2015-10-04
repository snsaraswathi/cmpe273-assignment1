package main

import (
    "net/http"
    "errors"
    "fmt"
    "github.com/gorilla/rpc"
    "github.com/gorilla/rpc/json"
    "io/ioutil"
    "os"
    "strconv"
    "strings"
    "regexp"
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
        TradeId int	`json:"tradeId"`
}

type Response2 struct {
        Stocks string
        CurrentMarketValue float32
        UnvestedAmount float32
}

func createHashTable() {
	UserProfile = make(map[int]string,10000)
}

var UserProfile map[int]string
type StockMarket struct {}
var countForTId int = 0

func checkTotalPercentage(array []string) bool{
	var percent float32
	var inputFmt string
	var percent0 float64
	var sum float32=0.00
	var returnVal bool
	returnVal=true
     
	for in := 1; in<len(array); in=in+2 {
        if(len(array[in]) > 1) {
                inputFmt = array[in][:len(array[in])-1]
                percent0,_ = strconv.ParseFloat(inputFmt,32)
                percent = float32(percent0)
                sum += percent
                } else {
                fmt.Println("Enter a % value greater than 0")
		returnVal =  false
                }
        }
        if(sum!=100) {
                fmt.Println("% does not add up to 100")
		returnVal = false
        }
	return returnVal
}

func stringToFloat(inputString string) float32 {
	var TempFloat64 float64
	var TempFloat32 float32
        TempFloat64,_ = strconv.ParseFloat(inputString,32)
        TempFloat32 = float32(TempFloat64)
	return TempFloat32
}


func main() {
	createHashTable()
 	s := rpc.NewServer()
        s.RegisterCodec(json.NewCodec(), "application/json")
        s.RegisterService(new(StockMarket), "")
        http.Handle("/rpc", s)
        http.ListenAndServe("localhost:10000", nil)
}

func (t *StockMarket) BuyStocks(r *http.Request, args *Request1, reply *Reply) error {
	var s,s1,s2 []string
	var purchase []int
	var stockValue1 []string
	var trimSpaceVar string
	var symbols []string
	var stockVal float32
	var stock []float32
	var percentageCheck bool
        var percentInFloat float32
        var inputFmt1 string 
	var tempPurchase int
	var tempTrackBalance float32
        var stringResp string
	var uidForMap int	
	var temp string
	var stocks []string
	var strUnvest string
	var strForHash string
	var unvest64 float64

//Split the input string based on ","	
	s= strings.Split(args.StockSymbolAndPercentage, ",")
	
	for v := range(s) {
	//Check if the input string format was valid
		var validId = regexp.MustCompile(`^[A-Z]+:+[0-9]+%$`)
		var inputCheck bool
		inputCheck = (validId.MatchString(s[v]))
	//If input string has invalid format throw an error
		if(!inputCheck){
			return errors.New("Input error: Invalid syntax. Please try again")
		}
	//Split the string based on ":"
        	s1= strings.Split(s[v], ":")
        	if(len(s1) != 2) {
                	fmt.Println("Input error")
        	} else {
                	s2 = append(s2,s1[0])
                	s2 = append(s2,s1[1])
       	 	}
    	}
	
//Check if the percentage is = 100. If not throw an error
	percentageCheck = checkTotalPercentage(s2)
	if(!percentageCheck) {
		return errors.New("Input error: Perentage does not add up to 100%. Please try again")
	}

//Call the Yahoo Finanace API and parse the return val to extract price
	for i :=0; i<len(s2); i=i+2 {
		symbols = append(symbols,s2[i])
                response, err := http.Get("http://finance.yahoo.com/webservice/v1/symbols/"+s2[i]+"/quote?format=json")
	if err != nil {
        fmt.Printf("%s", err)
        os.Exit(1)
    } else {
        defer response.Body.Close()
        contents, err := ioutil.ReadAll(response.Body)
        if err != nil {
            fmt.Printf("%s", err)
            os.Exit(1)
        }   
        stringResp = string(contents)
			
	var validId = regexp.MustCompile("\"price\" : \"(.*?)\"")
        var valReturn = validId.FindString(stringResp)  
//If an invalid symbol was entered throw an error
	if(valReturn == "") {
		return errors.New("Invaild symbol entered")
	}	

        var stringVal1  = regexp.MustCompile("[0-9]*(\\.[0-9]+)")
	var stringVal = stringVal1.FindString(valReturn)
	trimSpaceVar = strings.TrimSpace(stringVal)
        stockValue1 = append(stockValue1,trimSpaceVar)
        }
        }

//Calculate the number of stocks to buy	
	for in := 1; in<len(s2); in=in+2 {
        	if(len(s2[in]) > 1) {
			inputFmt1 = s2[in][:len(s2[in])-1]
			percentInFloat = stringToFloat(inputFmt1)
                	stock = append(stock,(percentInFloat/100) * args.Budget)
		}
	}
        for stck := range(stockValue1) {
		stockVal = stringToFloat(stockValue1[stck])
		tempPurchase = int(stock[stck]/stockVal)
                purchase=append(purchase,tempPurchase)
		tempTrackBalance = stockVal * float32(tempPurchase)
		reply.UnvestedAmount += (stock[stck] - tempTrackBalance)
	}

//Set the value of trade ID
	countForTId = countForTId+1	
	reply.TradeId = countForTId
	uidForMap = reply.TradeId

//Concatenate the strings to form the return string
	for value := range(symbols) {
		temp = (symbols[value]+":"+(strconv.Itoa(purchase[value]))+":$"+stockValue1[value])
		stocks = append(stocks, temp)
	}
	reply.StrOut = stocks[0]
	for val := 1; val<len(stocks); val++ {
		reply.StrOut +=  "," +stocks[val]
	}	
	unvest64 = float64(reply.UnvestedAmount)
	strUnvest = strconv.FormatFloat(unvest64,'f',6,32)
	strForHash = reply.StrOut+","+strUnvest
//Insert into the hash table
	UserProfile[uidForMap] = strForHash
        return nil
}

func (t *StockMarket) CheckPortfolio(r *http.Request, args *Request2, response *Response2) error {  
	var check bool = false
	var s []string
	var s3 []string
    	var s4 []string
	var stockValue []string
	var currStock32 float32
	var oldStock32 float32
	var trimSpaceVar1 string
	var temp string
	var symbols []string
	var stocksPurchased []string
	var localUnvestAmt string
	var localUnvest32 float32
	var index int
	var temp2 string
	var stocks2 []string
        var stringResp1 string
        var totalCurrentMarketValue float32
        var stockValue32 float32
        var currStock_32 float32

//Loop through the hash table to find the trade id 
	for key,views := range UserProfile {
	
	//Check if the user input and trade id in hash table match 
                if(key == args.TradeId) {
                        check = true
	//Do "splits" on the string to seperate the symbols, stocks purchased, stock value and  unvested amount
        		s= strings.Split(views, ",")
			index = len(s)-1	
			localUnvestAmt = s[index]
               		for v := 0; v<len(s)-1; v++ {
        			s3= strings.Split(s[v], ":")
                		s4 = append(s4,s3[0])
                		s4 = append(s4,s3[1])
				s4 = append(s4,s3[2])
    			}
	//Create an array containg the number of stocks purchased
	for st := 1; st<len(s4); st=st+3 {
		stocksPurchased = append(stocksPurchased, s4[st])
	}

	for i :=0; i<len(s4); i=i+3 {
	//Create an array containg the symbol 
	symbols = append(symbols,s4[i])
//Call the Yahoo Finanace API to find current market value and parse the return value to extract price
	apiResponse, err := http.Get("http://finance.yahoo.com/webservice/v1/symbols/"+s4[i]+"/quote?format=json")
        if err != nil {
        fmt.Printf("%s", err)
        os.Exit(1)
    	} else {
        defer apiResponse.Body.Close()
        contents, err := ioutil.ReadAll(apiResponse.Body)
        if err != nil {
            fmt.Printf("%s", err)
            os.Exit(1)
        }
        stringResp1 = string(contents)

        var validId1 = regexp.MustCompile("\"price\" : \"(.*?)\"")
        var valReturn1 = validId1.FindString(stringResp1)

        var stringVal2  = regexp.MustCompile("[0-9]*(\\.[0-9]+)")
        var stringVal3 = stringVal2.FindString(valReturn1)
        trimSpaceVar1 = strings.TrimSpace(stringVal3)

//Convert the stock value to float32
	currStock32 = stringToFloat(trimSpaceVar1)

//Trim  "$" symbol of the old stock value 
	temp = strings.TrimPrefix(s4[i+2], "$")

//Convert the old stock value to float32	
	oldStock32 = stringToFloat(temp)
	
//Check for profit or loss
	if(oldStock32 > currStock32 ){
		//Profit: add a "+" sign 
			stockValue= append(stockValue, ":+$"+trimSpaceVar1)
                } else if (oldStock32 < currStock32){
		//Loss: add a "-" sign 
			stockValue= append(stockValue, ":-$"+trimSpaceVar1)
                } else {
		//No profit or loss: add only the ":$" symbol 
			stockValue= append(stockValue, ":$"+trimSpaceVar1)
                }
        }
    }
	}//end if
     }//end for

//Trade id in the input does not match any "keys" in the hash table
     if(!check) {
		return errors.New("Invaild trade ID")
     }
//Loop over to create the return string
	for value := range(stockValue) {
                temp2 = (symbols[value]+":" + stocksPurchased[value] + stockValue[value])
		stocks2 = append(stocks2, temp2)
        }

	response.Stocks = stocks2[0]
	for val := 1; val<len(stocks2); val++ {
                 response.Stocks +=  "," +stocks2[val]
         }
//Use regexp to trim the + or - and the :$ symbols from the stock value	
	for index1 := range(stockValue) {
 	var exp  = regexp.MustCompile("[0-9]*(\\.[0-9]+)")
        var temp1 = exp.FindString(stockValue[index1])

	currStock_32 = stringToFloat(temp1)
	stockValue32 = stringToFloat(stocksPurchased[index1])

//Calculate the current market value	
	totalCurrentMarketValue += (currStock_32 * stockValue32)
	}
	response.CurrentMarketValue = totalCurrentMarketValue
	localUnvest32 = stringToFloat(localUnvestAmt)
	response.UnvestedAmount = localUnvest32
	return nil
}
