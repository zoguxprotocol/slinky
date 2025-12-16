package zogux

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"

	"github.com/zoguxprotocol/slinky/cmd/constants/marketmaps"
	"github.com/zoguxprotocol/slinky/oracle/config"
	"github.com/zoguxprotocol/slinky/providers/apis/coinmarketcap"
	apihandlers "github.com/zoguxprotocol/slinky/providers/base/api/handlers"
	"github.com/zoguxprotocol/slinky/providers/base/api/metrics"
	providertypes "github.com/zoguxprotocol/slinky/providers/types"
	mmclient "github.com/zoguxprotocol/slinky/service/clients/marketmap/types"
	mmtypes "github.com/zoguxprotocol/slinky/x/marketmap/types"
)

var (
	_         mmclient.MarketMapFetcher = &MultiMarketMapRestAPIFetcher{}
	ZOGUXChain                           = mmclient.Chain{
		ChainID: ChainID,
	}
)

// NewZOGUXResearchMarketMapFetcher returns a MultiMarketMapFetcher composed of zogux mainnet + research
// apiDataHandlers.
func DefaultZOGUXResearchMarketMapFetcher(
	rh apihandlers.RequestHandler,
	metrics metrics.APIMetrics,
	api config.APIConfig,
	logger *zap.Logger,
) (*MultiMarketMapRestAPIFetcher, error) {
	if rh == nil {
		return nil, fmt.Errorf("request handler is nil")
	}

	if metrics == nil {
		return nil, fmt.Errorf("metrics is nil")
	}

	if !api.Enabled {
		return nil, fmt.Errorf("api is not enabled")
	}

	if err := api.ValidateBasic(); err != nil {
		return nil, err
	}

	if len(api.Endpoints) != 2 {
		return nil, fmt.Errorf("expected two endpoint, got %d", len(api.Endpoints))
	}

	if logger == nil {
		return nil, fmt.Errorf("logger is nil")
	}

	// make a zogux research api-handler
	researchAPIDataHandler, err := NewResearchAPIHandler(logger, api)
	if err != nil {
		return nil, err
	}

	mainnetAPIDataHandler := &APIHandler{
		logger: logger,
		api:    api,
	}

	mainnetFetcher, err := apihandlers.NewRestAPIFetcher(
		rh,
		mainnetAPIDataHandler,
		metrics,
		api,
		logger,
	)
	if err != nil {
		return nil, err
	}

	researchFetcher, err := apihandlers.NewRestAPIFetcher(
		rh,
		researchAPIDataHandler,
		metrics,
		api,
		logger,
	)
	if err != nil {
		return nil, err
	}

	return NewZOGUXResearchMarketMapFetcher(
		mainnetFetcher,
		researchFetcher,
		logger,
		api.Name == ResearchCMCAPIHandlerName,
	), nil
}

// MultiMarketMapRestAPIFetcher is an implementation of a RestAPIFetcher that wraps
// two underlying Fetchers for fetching the market-map according to zogux mainnet and
// the additional markets that can be added according to the zogux research json.
type MultiMarketMapRestAPIFetcher struct {
	// zogux mainnet fetcher is the api-fetcher for the zogux mainnet market-map
	zoguxMainnetFetcher mmclient.MarketMapFetcher

	// zogux research fetcher is the api-fetcher for the zogux research market-map
	zoguxResearchFetcher mmclient.MarketMapFetcher

	// logger is the logger for the fetcher
	logger *zap.Logger

	// isCMCOnly is a flag that indicates whether the fetcher should only return CoinMarketCap markets.
	isCMCOnly bool
}

// NewZOGUXResearchMarketMapFetcher returns an aggregated market-map among the zogux mainnet and the zogux research json.
func NewZOGUXResearchMarketMapFetcher(
	mainnetFetcher, researchFetcher mmclient.MarketMapFetcher,
	logger *zap.Logger,
	isCMCOnly bool,
) *MultiMarketMapRestAPIFetcher {
	return &MultiMarketMapRestAPIFetcher{
		zoguxMainnetFetcher:  mainnetFetcher,
		zoguxResearchFetcher: researchFetcher,
		logger:              logger.With(zap.String("module", "zogux-research-market-map-fetcher")),
		isCMCOnly:           isCMCOnly,
	}
}

// Fetch fetches the market map from the underlying fetchers and combines the results. If any of the underlying
// fetchers fetch for a chain that is different from the chain that the fetcher is initialized with, those responses
// will be ignored.
func (f *MultiMarketMapRestAPIFetcher) Fetch(ctx context.Context, chains []mmclient.Chain) mmclient.MarketMapResponse {
	// call the underlying fetchers + await their responses
	// channel to aggregate responses
	zoguxMainnetResponseChan := make(chan mmclient.MarketMapResponse, 1) // buffer so that sends / receives are non-blocking
	zoguxResearchResponseChan := make(chan mmclient.MarketMapResponse, 1)

	var wg sync.WaitGroup
	wg.Add(2)

	// fetch zogux mainnet
	go func() {
		defer wg.Done()
		zoguxMainnetResponseChan <- f.zoguxMainnetFetcher.Fetch(ctx, chains)
		f.logger.Debug("fetched valid market-map from zogux mainnet")
	}()

	// fetch zogux research
	go func() {
		defer wg.Done()
		zoguxResearchResponseChan <- f.zoguxResearchFetcher.Fetch(ctx, chains)
		f.logger.Debug("fetched valid market-map from zogux research")
	}()

	// wait for both fetchers to finish
	wg.Wait()

	zoguxMainnetMarketMapResponse := <-zoguxMainnetResponseChan
	zoguxResearchMarketMapResponse := <-zoguxResearchResponseChan

	// if the zogux mainnet market-map response failed, return the zogux mainnet failed response
	if _, ok := zoguxMainnetMarketMapResponse.UnResolved[ZOGUXChain]; ok {
		f.logger.Error("zogux mainnet market-map fetch failed", zap.Any("response", zoguxMainnetMarketMapResponse))
		return zoguxMainnetMarketMapResponse
	}

	// if the zogux research market-map response failed, return the zogux research failed response
	if _, ok := zoguxResearchMarketMapResponse.UnResolved[ZOGUXChain]; ok {
		f.logger.Error("zogux research market-map fetch failed", zap.Any("response", zoguxResearchMarketMapResponse))
		return zoguxResearchMarketMapResponse
	}

	// otherwise, add all markets from zogux research
	zoguxMainnetMarketMap := zoguxMainnetMarketMapResponse.Resolved[ZOGUXChain].Value.MarketMap

	resolved, ok := zoguxResearchMarketMapResponse.Resolved[ZOGUXChain]
	if ok {
		for ticker, market := range resolved.Value.MarketMap.Markets {
			// if the market is not already in the zogux mainnet market-map, add it
			if _, ok := zoguxMainnetMarketMap.Markets[ticker]; !ok {
				f.logger.Debug("adding market from zogux research", zap.String("ticker", ticker))
				zoguxMainnetMarketMap.Markets[ticker] = market
			}
		}
	}

	// if the fetcher is only for CoinMarketCap markets, filter out all non-CMC markets
	if f.isCMCOnly {
		for ticker, market := range zoguxMainnetMarketMap.Markets {
			market.Ticker.MinProviderCount = 1
			zoguxMainnetMarketMap.Markets[ticker] = market

			var (
				seenCMC     = false
				cmcProvider mmtypes.ProviderConfig
			)

			for _, provider := range market.ProviderConfigs {
				if provider.Name == coinmarketcap.Name {
					seenCMC = true
					cmcProvider = provider
				}
			}

			// if we saw a CMC provider, add it to the market
			if seenCMC {
				market.ProviderConfigs = []mmtypes.ProviderConfig{cmcProvider}
				zoguxMainnetMarketMap.Markets[ticker] = market
				continue
			}

			// If we did not see a CMC provider, we can attempt to add it using the CMC marketmap
			cmcMarket, ok := marketmaps.CoinMarketCapMarketMap.Markets[ticker]
			if !ok {
				f.logger.Info("did not find CMC market for ticker", zap.String("ticker", ticker))
				delete(zoguxMainnetMarketMap.Markets, ticker)
				continue
			}

			// add the CMC provider to the market
			market.ProviderConfigs = cmcMarket.ProviderConfigs
			zoguxMainnetMarketMap.Markets[ticker] = market
		}
	}

	// validate the combined market-map
	if err := zoguxMainnetMarketMap.ValidateBasic(); err != nil {
		f.logger.Error("combined market-map failed validation", zap.Error(err))

		return mmclient.NewMarketMapResponseWithErr(
			chains,
			providertypes.NewErrorWithCode(
				fmt.Errorf("combined market-map failed validation: %w", err),
				providertypes.ErrorUnknown,
			),
		)
	}

	zoguxMainnetMarketMapResponse.Resolved[ZOGUXChain].Value.MarketMap = zoguxMainnetMarketMap

	return zoguxMainnetMarketMapResponse
}
