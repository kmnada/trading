import { useState, useEffect, useRef } from 'react';
import { ColorType, createChart, ISeriesApi, Time } from 'lightweight-charts';
import './App.css'
import ReactSelect from 'react-select';

const WEBSOCKET_URL = 'ws://localhost:8081/ws';

interface OHLC {
  symbol: string;
  open: number;
  high: number;
  low: number;
  close: number;
  timestamp: Time;
}

const symbolOptions = [
  { value: 'btcusdt', label: 'BTCUSDT' },
  { value: 'ethusdt', label: 'ETHUSDT' },
  { value: 'pepeusdt', label: 'PEPEUSDT' },
]

const App = () => {

  const [selectedSymbol, setSelectedSymbol] = useState(symbolOptions[0]);
  const [currentPrice, setCurrentPrice] = useState<number | null>(null);
  const [priceColor, setPriceColor] = useState('');

  const chartContainerRef = useRef<HTMLDivElement | null>(null);
  const chartRef = useRef<ISeriesApi<'Candlestick'> | null>(null);

  useEffect(() => {
    if (!chartContainerRef.current) return;

    const chart = createChart(chartContainerRef.current, {
      width: chartContainerRef.current.clientWidth,
      height: 400,
      layout: {
        background: { type: ColorType.Solid, color: "#fff" },
        textColor: "#000",
      },
      grid: {
        vertLines: { color: "#eee" },
        horzLines: { color: "#eee" },
      },
      leftPriceScale: { borderColor: '#d1d4dc' },
      rightPriceScale: { borderColor: '#d1d4dc' },
      timeScale: { borderColor: '#d1d4dc' }
    })

    const candlestickSeries = chart.addCandlestickSeries();
    chartRef.current = candlestickSeries;

    const handleResize = () => {
      if(chartContainerRef.current) {
        chart.applyOptions({ width: chartContainerRef.current.clientWidth });
      }
    }
    window.addEventListener('resize', handleResize);

    return () => {
      window.removeEventListener('resize', handleResize);
      chart.remove()
    }
  }, []);

  useEffect(() => {
    if(!selectedSymbol) return;

    console.log(`Switched to symbol ${selectedSymbol.value}`);

    const ws = new WebSocket(WEBSOCKET_URL);

    ws.onopen = () => {
      console.log(`Connected to websocket for symbol: ${selectedSymbol.value}`)
    }

    ws.onmessage = (event) => {

      try {
        const data: OHLC = JSON.parse(event.data);
  
        if(data.symbol === selectedSymbol.value) {
          const { open, high, low, close, timestamp } = data;
  
          chartRef.current?.update({
            time: timestamp,
            open,
            high,
            low,
            close
          })
  
          if (currentPrice !== null && close > currentPrice) {
            setPriceColor('green');
          } else if (currentPrice !== null && close < currentPrice) {
            setPriceColor('red');
          }
  
          setCurrentPrice(close);
        }

      } catch(error) {
        console.log('error message', error)
      }
    }

    ws.onerror = (error) => {
      console.log(`Websocket error: ${error}`, error)
    };

    return () => {
      ws.close()
    }

  }, [selectedSymbol, currentPrice])

  return (
    <div className='App'>
      <h2> {`OHLC Candle stick Chart for symbol ${selectedSymbol.label}`} </h2>
      <ReactSelect
        className='selector'
        options={symbolOptions}
        value={selectedSymbol}
        onChange={(selectedOption) => selectedOption && setSelectedSymbol(selectedOption)}
         />

         {
          currentPrice !== null && (
            <h3 style={{ color: priceColor }}>
              latest price: { currentPrice.toFixed(2)}
            </h3>
          )
         }
         <div ref={chartContainerRef} className='chart' />

    </div>
  )
}

export default App
