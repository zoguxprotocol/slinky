package oracle

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/zoguxprotocol/slinky/oracle/config"
	"github.com/zoguxprotocol/slinky/providers/apis/zogux"
	"github.com/zoguxprotocol/slinky/providers/apis/marketmap"
	"github.com/zoguxprotocol/slinky/providers/base"
	apihandlers "github.com/zoguxprotocol/slinky/providers/base/api/handlers"
	apimetrics "github.com/zoguxprotocol/slinky/providers/base/api/metrics"
	providermetrics "github.com/zoguxprotocol/slinky/providers/base/metrics"
	"github.com/zoguxprotocol/slinky/service/clients/marketmap/types"
	mmtypes "github.com/zoguxprotocol/slinky/x/marketmap/types"
)

// MarketMapProviderFactory returns a sample implementation of the market map provider. This provider
// is responsible for fetching updates to the canonical market map on the given chain.
func MarketMapProviderFactory(
	logger *zap.Logger,
	providerMetrics providermetrics.ProviderMetrics,
	apiMetrics apimetrics.APIMetrics,
	cfg config.ProviderConfig,
) (*types.MarketMapProvider, error) {
	// Validate the provider config.
	err := cfg.ValidateBasic()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			MaxConnsPerHost: cfg.API.MaxQueries,
			Proxy:           http.ProxyFromEnvironment,
		},
		Timeout: cfg.API.Timeout,
	}

	var (
		apiDataHandler   types.MarketMapAPIDataHandler
		ids              []types.Chain
		marketMapFetcher types.MarketMapFetcher
	)

	requestHandler, err := apihandlers.NewRequestHandlerImpl(client)
	if err != nil {
		return nil, err
	}

	switch cfg.Name {
	case zogux.Name:
		apiDataHandler, err = zogux.NewAPIHandler(logger, cfg.API)
		ids = []types.Chain{{ChainID: zogux.ChainID}}
	case zogux.SwitchOverAPIHandlerName:
		marketMapFetcher, err = zogux.NewDefaultSwitchOverMarketMapFetcher(
			logger,
			cfg.API,
			requestHandler,
			apiMetrics,
		)
		ids = []types.Chain{{ChainID: zogux.ChainID}}
	case zogux.ResearchAPIHandlerName, zogux.ResearchCMCAPIHandlerName:
		marketMapFetcher, err = zogux.DefaultZOGUXResearchMarketMapFetcher(
			requestHandler,
			apiMetrics,
			cfg.API,
			logger,
		)
		ids = []types.Chain{{ChainID: zogux.ChainID}}
	default:
		marketMapFetcher, err = marketmap.NewMarketMapFetcher(
			logger,
			cfg.API,
			apiMetrics,
		)
		ids = []types.Chain{{ChainID: "local-node"}}
	}
	if err != nil {
		return nil, err
	}

	if marketMapFetcher == nil {
		marketMapFetcher, err = apihandlers.NewRestAPIFetcher(
			requestHandler,
			apiDataHandler,
			apiMetrics,
			cfg.API,
			logger,
		)
		if err != nil {
			return nil, err
		}
	}

	queryHandler, err := types.NewMarketMapAPIQueryHandlerWithMarketMapFetcher(
		logger,
		cfg.API,
		marketMapFetcher,
		apiMetrics,
	)
	if err != nil {
		return nil, err
	}

	return types.NewMarketMapProvider(
		base.WithName[types.Chain, *mmtypes.MarketMapResponse](cfg.Name),
		base.WithLogger[types.Chain, *mmtypes.MarketMapResponse](logger),
		base.WithAPIQueryHandler(queryHandler),
		base.WithAPIConfig[types.Chain, *mmtypes.MarketMapResponse](cfg.API),
		base.WithMetrics[types.Chain, *mmtypes.MarketMapResponse](providerMetrics),
		base.WithIDs[types.Chain, *mmtypes.MarketMapResponse](ids),
	)
}
