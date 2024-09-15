package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	
	"github.com/gorilla/websocket"
)

const BINANCE_ENDPOINT = "wss://stream.binance.com:9443/ws/"

type TickData struct {
	Symbol string
	Price float64
	Time int64
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan OHLC)
var mutex = &sync.Mutex{}

// Structs to store trade and OHLC data
type BinanceAggTrade struct {
	Symbol    string  `json:"s"`
	Price     string  `json:"p"`
	Time 	  int64   `json:"T"`
}

type OHLC struct {
	Symbol   string  `json:"symbol"`
	Open     float64 `json:"open"`
	High     float64 `json:"high"`
	Low      float64 `json:"low"` 
	Close    float64 `json:"close"`
	Timestamp int64  `json:"timestamp"`
}

var symbols = []string{"btcusdt", "ethusdt", "pepeusdt"}

var ohlcData = map[string]*OHLC{}


var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	
	for _, symbol := range symbols {
		ohlcData[symbol] = nil
	}

	http.HandleFunc("/ws", handleConnections)

	go handleBroadcasts()

	for _, symbol := range symbols {
		go connectBinance(symbol)
	}

	log.Println("Websocket server has started on :8081")
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal("Server error", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("Upgrade error", err)
		return
	}
	defer ws.Close()

	mutex.Lock()
	clients[ws] = true
	mutex.Unlock()

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			log.Printf("Client disconnected: %v", err)
			mutex.Lock()
			delete(clients, ws)
			mutex.Unlock()
			return
		}
	}
}

func connectBinance(symbol string) {
	url := fmt.Sprintf("%s%s@aggTrade", BINANCE_ENDPOINT, symbol)
	log.Printf("Connecting to Binance WebSocket: %s", url)

	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("WebSocket dial error:", err)
		return
	}
	defer c.Close()

	ohlcInterval := 1 * time.Minute
	resetTimer := time.NewTicker(ohlcInterval)
	defer resetTimer.Stop()

	for {
		select {
		case <-resetTimer.C:
			mutex.Lock()
			ohlc := ohlcData[symbol]
			if ohlc != nil {
				broadcast <- *ohlc
				ohlcData[symbol] = nil
			}
			mutex.Unlock()

		default: 
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("Error reading Binance websocket:", err)
				return
			}

			var trade BinanceAggTrade
			err = json.Unmarshal(message, &trade)
			if err != nil {
				log.Printf("error parsing binance message: %v", err)
				continue
			}

			price, err := strconv.ParseFloat(trade.Price, 64)
			if err == nil {
				aggregateTick(TickData{
					Symbol: strings.ToLower(trade.Symbol),
					Price: price,
					Time: trade.Time,
				})
			}
		}
	}
}

func aggregateTick(tick TickData) {
	mutex.Lock()
	defer mutex.Unlock()

	ohlc := ohlcData[tick.Symbol]
	if ohlc == nil {
		ohlcData[tick.Symbol] = &OHLC{
			Symbol: tick.Symbol,
			Open: tick.Price,
			High: tick.Price,
			Low: tick.Price,
			Close: tick.Price,
			Timestamp: time.Now().Unix(),
		}
		return
	}

	ohlc.Close = tick.Price
	if tick.Price > ohlc.High {
		ohlc.High = tick.Price
	}
	if tick.Price < ohlc.Low {
		ohlc.Low = tick.Price
	}
}

func handleBroadcasts() {
	for {
		ohlc := <-broadcast

		log.Printf("Broadcast into ohlc: %+v\n", ohlc);

		mutex.Lock()
		for client := range clients {
			err := client.WriteJSON(ohlc)
			if err != nil {
				log.Printf("Websocket error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
		mutex.Unlock()
	}
}
