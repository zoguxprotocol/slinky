package constants

import (
	"github.com/zoguxprotocol/slinky/oracle/config"
	"github.com/zoguxprotocol/slinky/oracle/constants"
	"github.com/zoguxprotocol/slinky/oracle/types"
	binanceapi "github.com/zoguxprotocol/slinky/providers/apis/binance"
	bitstampapi "github.com/zoguxprotocol/slinky/providers/apis/bitstamp"
	coinbaseapi "github.com/zoguxprotocol/slinky/providers/apis/coinbase"
	"github.com/zoguxprotocol/slinky/providers/apis/coingecko"
	"github.com/zoguxprotocol/slinky/providers/apis/coinmarketcap"
	"github.com/zoguxprotocol/slinky/providers/apis/defi/osmosis"
	"github.com/zoguxprotocol/slinky/providers/apis/defi/raydium"
	"github.com/zoguxprotocol/slinky/providers/apis/defi/uniswapv3"
	"github.com/zoguxprotocol/slinky/providers/apis/zogux"
	krakenapi "github.com/zoguxprotocol/slinky/providers/apis/kraken"
	"github.com/zoguxprotocol/slinky/providers/apis/marketmap"
	"github.com/zoguxprotocol/slinky/providers/apis/polymarket"
	"github.com/zoguxprotocol/slinky/providers/volatile"
	binancews "github.com/zoguxprotocol/slinky/providers/websockets/binance"
	"github.com/zoguxprotocol/slinky/providers/websockets/bitfinex"
	"github.com/zoguxprotocol/slinky/providers/websockets/bitstamp"
	"github.com/zoguxprotocol/slinky/providers/websockets/bybit"
	"github.com/zoguxprotocol/slinky/providers/websockets/coinbase"
	"github.com/zoguxprotocol/slinky/providers/websockets/cryptodotcom"
	"github.com/zoguxprotocol/slinky/providers/websockets/gate"
	"github.com/zoguxprotocol/slinky/providers/websockets/huobi"
	"github.com/zoguxprotocol/slinky/providers/websockets/kraken"
	"github.com/zoguxprotocol/slinky/providers/websockets/kucoin"
	"github.com/zoguxprotocol/slinky/providers/websockets/mexc"
	"github.com/zoguxprotocol/slinky/providers/websockets/okx"
	mmtypes "github.com/zoguxprotocol/slinky/service/clients/marketmap/types"
)

var (
	Providers = []config.ProviderConfig{
		// DEFI providers
		{
			Name: raydium.Name,
			API:  raydium.DefaultAPIConfig,
			Type: types.ConfigType,
		},
		{
			Name: uniswapv3.ProviderNames[constants.ETHEREUM],
			API:  uniswapv3.DefaultETHAPIConfig,
			Type: types.ConfigType,
		},
		{
			Name: uniswapv3.ProviderNames[constants.BASE],
			API:  uniswapv3.DefaultBaseAPIConfig,
			Type: types.ConfigType,
		},
		{
			Name: osmosis.Name,
			API:  osmosis.DefaultAPIConfig,
			Type: types.ConfigType,
		},

		// Exchange API providers
		{
			Name: binanceapi.Name,
			API:  binanceapi.DefaultNonUSAPIConfig,
			Type: types.ConfigType,
		},
		{
			Name: bitstampapi.Name,
			API:  bitstampapi.DefaultAPIConfig,
			Type: types.ConfigType,
		},
		{
			Name: coinbaseapi.Name,
			API:  coinbaseapi.DefaultAPIConfig,
			Type: types.ConfigType,
		},
		{
			Name: coingecko.Name,
			API:  coingecko.DefaultAPIConfig,
			Type: types.ConfigType,
		},
		{
			Name: coinmarketcap.Name,
			API:  coinmarketcap.DefaultAPIConfig,
			Type: types.ConfigType,
		},
		{
			Name: krakenapi.Name,
			API:  krakenapi.DefaultAPIConfig,
			Type: types.ConfigType,
		},
		{
			Name: volatile.Name,
			API:  volatile.DefaultAPIConfig,
			Type: types.ConfigType,
		},
		// Exchange WebSocket providers
		{
			Name:      binancews.Name,
			WebSocket: binancews.DefaultWebSocketConfig,
			Type:      types.ConfigType,
		},
		{
			Name:      bitfinex.Name,
			WebSocket: bitfinex.DefaultWebSocketConfig,
			Type:      types.ConfigType,
		},
		{
			Name:      bitstamp.Name,
			WebSocket: bitstamp.DefaultWebSocketConfig,
			Type:      types.ConfigType,
		},
		{
			Name:      bybit.Name,
			WebSocket: bybit.DefaultWebSocketConfig,
			Type:      types.ConfigType,
		},
		{
			Name:      coinbase.Name,
			WebSocket: coinbase.DefaultWebSocketConfig,
			Type:      types.ConfigType,
		},
		{
			Name:      cryptodotcom.Name,
			WebSocket: cryptodotcom.DefaultWebSocketConfig,
			Type:      types.ConfigType,
		},
		{
			Name:      gate.Name,
			WebSocket: gate.DefaultWebSocketConfig,
			Type:      types.ConfigType,
		},
		{
			Name:      huobi.Name,
			WebSocket: huobi.DefaultWebSocketConfig,
			Type:      types.ConfigType,
		},
		{
			Name:      kraken.Name,
			WebSocket: kraken.DefaultWebSocketConfig,
			Type:      types.ConfigType,
		},
		{
			Name:      kucoin.Name,
			WebSocket: kucoin.DefaultWebSocketConfig,
			API:       kucoin.DefaultAPIConfig,
			Type:      types.ConfigType,
		},
		{
			Name:      mexc.Name,
			WebSocket: mexc.DefaultWebSocketConfig,
			Type:      types.ConfigType,
		},
		{
			Name:      okx.Name,
			WebSocket: okx.DefaultWebSocketConfig,
			Type:      types.ConfigType,
		},

		// Polymarket provider
		{
			Name: polymarket.Name,
			API:  polymarket.DefaultAPIConfig,
			Type: types.ConfigType,
		},

		// MarketMap provider
		{
			Name: marketmap.Name,
			API:  marketmap.DefaultAPIConfig,
			Type: mmtypes.ConfigType,
		},
	}

	AlternativeMarketMapProviders = []config.ProviderConfig{
		{
			Name: zogux.Name,
			API:  zogux.DefaultAPIConfig,
			Type: mmtypes.ConfigType,
		},
		{
			Name: zogux.SwitchOverAPIHandlerName,
			API:  zogux.DefaultSwitchOverAPIConfig,
			Type: mmtypes.ConfigType,
		},
		{
			Name: zogux.ResearchAPIHandlerName,
			API:  zogux.DefaultResearchAPIConfig,
			Type: mmtypes.ConfigType,
		},
		{
			Name: zogux.ResearchCMCAPIHandlerName,
			API:  zogux.DefaultResearchCMCAPIConfig,
			Type: mmtypes.ConfigType,
		},
	}

	MarketMapProviderNames = map[string]struct{}{
		zogux.Name:                      {},
		zogux.SwitchOverAPIHandlerName:  {},
		zogux.ResearchAPIHandlerName:    {},
		zogux.ResearchCMCAPIHandlerName: {},
		marketmap.Name:                 {},
	}
)
