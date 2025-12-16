package zogux_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/zoguxprotocol/slinky/providers/apis/coinmarketcap"
	zoguxtypes "github.com/zoguxprotocol/slinky/providers/apis/zogux/types"
	"github.com/zoguxprotocol/slinky/providers/base/testutils"
	"github.com/zoguxprotocol/slinky/providers/websockets/binance"
	"github.com/zoguxprotocol/slinky/providers/websockets/coinbase"
	"github.com/zoguxprotocol/slinky/providers/websockets/gate"
	"github.com/zoguxprotocol/slinky/providers/websockets/kucoin"
	"github.com/zoguxprotocol/slinky/providers/websockets/mexc"
	"github.com/zoguxprotocol/slinky/providers/websockets/okx"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/zoguxprotocol/slinky/oracle/config"
	slinkytypes "github.com/zoguxprotocol/slinky/pkg/types"
	"github.com/zoguxprotocol/slinky/providers/apis/zogux"
	"github.com/zoguxprotocol/slinky/service/clients/marketmap/types"
	mmtypes "github.com/zoguxprotocol/slinky/x/marketmap/types"
)

func TestNewResearchAPIHandler(t *testing.T) {
	t.Run("fail if the name is incorrect", func(t *testing.T) {
		_, err := zogux.NewResearchAPIHandler(zap.NewNop(), config.APIConfig{
			Name: "incorrect",
		})
		require.Error(t, err)
	})

	t.Run("fail if the api is not enabled", func(t *testing.T) {
		_, err := zogux.NewResearchAPIHandler(zap.NewNop(), config.APIConfig{
			Name:    zogux.ResearchAPIHandlerName,
			Enabled: false,
		})
		require.Error(t, err)
	})

	t.Run("test failure of api-config validation", func(t *testing.T) {
		cfg := zogux.DefaultResearchAPIConfig
		cfg.Endpoints = []config.Endpoint{
			{
				URL: "",
			},
		}

		_, err := zogux.NewResearchAPIHandler(zap.NewNop(), cfg)
		require.Error(t, err)
	})

	t.Run("test failure if no endpoint is given", func(t *testing.T) {
		cfg := zogux.DefaultResearchAPIConfig
		cfg.Endpoints = nil

		_, err := zogux.NewResearchAPIHandler(zap.NewNop(), cfg)
		require.Error(t, err)
	})

	t.Run("test success", func(t *testing.T) {
		_, err := zogux.NewResearchAPIHandler(zap.NewNop(), zogux.DefaultResearchAPIConfig)
		require.NoError(t, err)
	})
}

// TestCreateURL tests that:
//   - If no chain in the given chains are zogux - fail
//   - If one chain in the given chains is zogux - return the first endpoint configured
func TestCreateURLResearchHandler(t *testing.T) {
	ah, err := zogux.NewResearchAPIHandler(
		zap.NewNop(),
		zogux.DefaultResearchAPIConfig,
	)
	require.NoError(t, err)

	t.Run("non-zogux chains", func(t *testing.T) {
		chains := []types.Chain{
			{
				ChainID: "osmosis",
			},
		}

		url, err := ah.CreateURL(chains)
		require.Error(t, err)
		require.Empty(t, url)
	})
	t.Run("multiple chains w/ a zogux chain", func(t *testing.T) {
		chains := []types.Chain{
			{
				ChainID: "osmosis",
			},
			{
				ChainID: zogux.ChainID,
			},
		}

		url, err := ah.CreateURL(chains)
		require.NoError(t, err)
		require.Equal(t, zogux.DefaultResearchAPIConfig.Endpoints[1].URL, url)
	})
}

func TestParseResponseResearchAPI(t *testing.T) {
	ah, err := zogux.NewResearchAPIHandler(
		zap.NewNop(),
		zogux.DefaultResearchAPIConfig,
	)
	require.NoError(t, err)

	t.Run("fail if none of the chains given are zogux", func(t *testing.T) {
		chains := []types.Chain{
			{
				ChainID: "osmosis",
			},
		}

		resp := ah.ParseResponse(chains, &http.Response{})
		// expect a failure response for each chain
		require.Len(t, resp.UnResolved, 1)
		require.Len(t, resp.Resolved, 0)

		require.Error(t, resp.UnResolved[chains[0]])
	})

	t.Run("failing to parse ResearchJSON response", func(t *testing.T) {
		chains := []types.Chain{
			{
				ChainID: zogux.ChainID,
			},
		}

		resp := ah.ParseResponse(chains, &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString("")),
		})

		require.Len(t, resp.UnResolved, 1)
		require.Len(t, resp.Resolved, 0)

		require.Error(t, resp.UnResolved[chains[0]])
	})

	t.Run("failing to convert ResearchJSON response into QueryAllMarketsParams", func(t *testing.T) {
		chains := []types.Chain{
			{
				ChainID: zogux.ChainID,
			},
		}

		resp := ah.ParseResponse(chains, &http.Response{
			StatusCode: http.StatusOK,
			Body: io.NopCloser(bytes.NewBufferString(`{
				"1INCH": {
				}
			}`)),
		})

		require.Len(t, resp.UnResolved, 1)
		require.Len(t, resp.Resolved, 0)

		require.Error(t, resp.UnResolved[chains[0]])
	})

	t.Run("success", func(t *testing.T) {
		chains := []types.Chain{
			{
				ChainID: zogux.ChainID,
			},
		}

		researchJSON := zoguxtypes.ResearchJSON{
			"1INCH": {
				ResearchJSONMarketParam: zoguxtypes.ResearchJSONMarketParam{
					ID:                0,
					Pair:              "1INCH-USD",
					Exponent:          -10.0,
					MinPriceChangePpm: 4000,
					MinExchanges:      3,
					ExchangeConfigJSON: []zoguxtypes.ExchangeMarketConfigJson{
						{
							ExchangeName: "Binance",
							Ticker:       "1INCHUSDT",
						},
						{
							ExchangeName: "CoinbasePro",
							Ticker:       "1INCH-USD",
						},
						{
							ExchangeName: "Gate",
							Ticker:       "1INCH_USDT",
						},
						{
							ExchangeName: "Kucoin",
							Ticker:       "1INCH-USDT",
						},
						{
							ExchangeName: "Mexc",
							Ticker:       "1INCH_USDT",
						},
						{
							ExchangeName: "Okx",
							Ticker:       "1INCH-USDT",
						},
					},
				},
			},
		}
		bz, err := json.Marshal(researchJSON)
		require.NoError(t, err)

		resp := ah.ParseResponse(chains, testutils.CreateResponseFromJSON(string(bz)))

		require.Len(t, resp.UnResolved, 0)
		require.Len(t, resp.Resolved, 1)

		mm := resp.Resolved[chains[0]].Value.MarketMap
		require.Len(t, mm.Markets, 1)

		// index by the pair
		market, ok := mm.Markets["1INCH/USD"]
		require.True(t, ok)

		// check the ticker
		expectedTicker := mmtypes.Ticker{
			CurrencyPair:     slinkytypes.NewCurrencyPair("1INCH", "USD"),
			Decimals:         10,
			MinProviderCount: 3,
			Enabled:          true,
		}
		require.Equal(t, expectedTicker, market.Ticker)

		// check each provider
		expectedProviders := map[string]mmtypes.ProviderConfig{
			binance.Name: {
				Name:           binance.Name,
				OffChainTicker: "1INCHUSDT",
			},
			coinbase.Name: {
				Name:           coinbase.Name,
				OffChainTicker: "1INCH-USD",
			},
			gate.Name: {
				Name:           gate.Name,
				OffChainTicker: "1INCH_USDT",
			},
			kucoin.Name: {
				Name:           kucoin.Name,
				OffChainTicker: "1INCH-USDT",
			},
			mexc.Name: {
				Name:           mexc.Name,
				OffChainTicker: "1INCHUSDT",
			},
			okx.Name: {
				Name:           okx.Name,
				OffChainTicker: "1INCH-USDT",
			},
		}

		for _, provider := range market.ProviderConfigs {
			expectedProvider, ok := expectedProviders[provider.Name]
			require.True(t, ok)
			require.Equal(t, expectedProvider, provider)
		}
	})
}

func TestParseResponseResearchCMCAPI(t *testing.T) {
	ah, err := zogux.NewResearchAPIHandler(
		zap.NewNop(),
		zogux.DefaultResearchCMCAPIConfig,
	)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		chains := []types.Chain{
			{
				ChainID: zogux.ChainID,
			},
		}

		researchJSON := zoguxtypes.ResearchJSON{
			"1INCH": {
				ResearchJSONMarketParam: zoguxtypes.ResearchJSONMarketParam{
					ID:                0,
					Pair:              "1INCH-USD",
					Exponent:          -10.0,
					MinPriceChangePpm: 4000,
					MinExchanges:      3,
					ExchangeConfigJSON: []zoguxtypes.ExchangeMarketConfigJson{
						{
							ExchangeName: "Binance",
							Ticker:       "1INCHUSDT",
						},
						{
							ExchangeName: "CoinbasePro",
							Ticker:       "1INCH-USD",
						},
						{
							ExchangeName: "Gate",
							Ticker:       "1INCH_USDT",
						},
						{
							ExchangeName: "Kucoin",
							Ticker:       "1INCH-USDT",
						},
						{
							ExchangeName: "Mexc",
							Ticker:       "1INCH_USDT",
						},
						{
							ExchangeName: "Okx",
							Ticker:       "1INCH-USDT",
						},
					},
				},
				MetaData: zoguxtypes.MetaData{
					CMCID: 1,
				},
			},
		}

		bz, err := json.Marshal(researchJSON)
		require.NoError(t, err)

		resp := ah.ParseResponse(chains, testutils.CreateResponseFromJSON(string(bz)))

		require.Len(t, resp.UnResolved, 0)
		require.Len(t, resp.Resolved, 1)

		mm := resp.Resolved[chains[0]].Value.MarketMap
		require.Len(t, mm.Markets, 1)

		// index by the pair
		market, ok := mm.Markets["1INCH/USD"]
		require.True(t, ok)

		// check the ticker
		expectedTicker := mmtypes.Ticker{
			CurrencyPair:     slinkytypes.NewCurrencyPair("1INCH", "USD"),
			Decimals:         10,
			MinProviderCount: 1,
			Enabled:          true,
		}
		require.Equal(t, expectedTicker, market.Ticker)

		// check each provider
		expectedProviders := map[string]mmtypes.ProviderConfig{
			coinmarketcap.Name: {
				Name:           coinmarketcap.Name,
				OffChainTicker: "1",
			},
		}

		for _, provider := range market.ProviderConfigs {
			expectedProvider, ok := expectedProviders[provider.Name]
			require.True(t, ok)
			require.Equal(t, expectedProvider, provider)
		}
	})
}
