package zogux

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/zoguxprotocol/slinky/oracle/config"
	"github.com/zoguxprotocol/slinky/pkg/arrays"
	"github.com/zoguxprotocol/slinky/providers/apis/coinmarketcap"
	zoguxtypes "github.com/zoguxprotocol/slinky/providers/apis/zogux/types"
	providertypes "github.com/zoguxprotocol/slinky/providers/types"
	"github.com/zoguxprotocol/slinky/service/clients/marketmap/types"
)

var _ types.MarketMapAPIDataHandler = (*ResearchAPIHandler)(nil)

// NewResearchAPIHandler returns a new MarketMap MarketMapAPIDataHandler.
func NewResearchAPIHandler(
	logger *zap.Logger,
	api config.APIConfig,
) (*ResearchAPIHandler, error) {
	if api.Name != ResearchAPIHandlerName && api.Name != ResearchCMCAPIHandlerName {
		return nil, fmt.Errorf("expected api config name %s or %s, got %s", ResearchAPIHandlerName, ResearchCMCAPIHandlerName, api.Name)
	}

	if !api.Enabled {
		return nil, fmt.Errorf("api config for %s is not enabled", ResearchAPIHandlerName)
	}

	if err := api.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("invalid api config for %s: %w", ResearchAPIHandlerName, err)
	}

	return &ResearchAPIHandler{
		APIHandler: APIHandler{
			api:    api,
			logger: logger,
		},
		url:     api.Endpoints[1].URL, // We assume the first URL is the endpoint of the zogux mainnet
		onlyCMC: api.Name == ResearchCMCAPIHandlerName,
	}, nil
}

// ResearchAPIHandler is a subclass of the zogux_chain.ResearchAPIHandler. It interprets the zogux ResearchJSON
// as a market-map.
type ResearchAPIHandler struct {
	APIHandler

	// url is the URL to query for the market map.
	url string
	// onlyCMC is a flag that indicates whether the handler should only return CoinMarketCap markets.
	onlyCMC bool
}

// CreateURL returns a static url (the url of the first configured endpoint). If the zogux chain is not
// configured in the request, an error is returned.
func (h *ResearchAPIHandler) CreateURL(chains []types.Chain) (string, error) {
	// expect at least one chain to be a zogux chain
	if _, ok := arrays.CheckEntryInArray(types.Chain{
		ChainID: ChainID,
	}, chains); !ok {
		return "", fmt.Errorf("zogux chain is not configured in request for the zogux research json")
	}

	return h.url, nil
}

// ParseResponse parses the response from the zogux ResearchJSON API into a MarketMap, and
// unmarshals the market-map in accordance with the underlying zogux ResearchAPIHandler.
func (h *ResearchAPIHandler) ParseResponse(
	chains []types.Chain,
	resp *http.Response,
) types.MarketMapResponse {
	// expect at least one chain to be a zogux chain
	chain, ok := arrays.CheckEntryInArray(types.Chain{
		ChainID: ChainID,
	}, chains)
	if !ok {
		h.logger.Error("zogux chain is not configured in request for the zogux research json")
		return types.NewMarketMapResponseWithErr(
			chains,
			providertypes.NewErrorWithCode(
				fmt.Errorf("expected one chain, got %d", len(chains)),
				providertypes.ErrorInvalidAPIChains,
			),
		)
	}

	// parse the response
	// unmarshal the response body into a zogux research json
	var research zoguxtypes.ResearchJSON
	if err := json.NewDecoder(resp.Body).Decode(&research); err != nil {
		h.logger.Error("failed to parse zogux research json response", zap.Error(err))
		return types.NewMarketMapResponseWithErr(
			chains,
			providertypes.NewErrorWithCode(
				fmt.Errorf("failed to parse zogux research json response: %w", err),
				providertypes.ErrorFailedToDecode,
			),
		)
	}

	// convert the zogux research json into a QueryAllMarketsParamsResponse
	respMarketParams, err := h.researchJSONToQueryAllMarketsParamsResponse(research)
	if err != nil {
		h.logger.Error("failed to convert zogux research json into QueryAllMarketsParamsResponse", zap.Error(err))
		return types.NewMarketMapResponseWithErr(
			chains,
			providertypes.NewErrorWithCode(
				fmt.Errorf("failed to convert zogux research json into QueryAllMarketsParamsResponse: %w", err),
				providertypes.ErrorFailedToDecode,
			),
		)
	}

	// convert the response to a market-map
	marketMap, err := ConvertMarketParamsToMarketMap(respMarketParams)
	if err != nil {
		h.logger.Error("failed to convert QueryAllMarketsParamsResponse into MarketMap", zap.Error(err))
		return types.NewMarketMapResponseWithErr(
			chains,
			providertypes.NewErrorWithCode(
				fmt.Errorf("failed to convert QueryAllMarketsParamsResponse into MarketMap: %w", err),
				providertypes.ErrorFailedToDecode,
			),
		)
	}

	// resolve the response under the zogux chain
	resolved := make(types.ResolvedMarketMap)
	resolved[chain] = types.NewMarketMapResult(&marketMap, time.Now())

	h.logger.Debug("successfully resolved zogux research json into a market map", zap.Int("num_markets", len(marketMap.MarketMap.Markets)))
	return types.NewMarketMapResponse(resolved, nil)
}

// researchJSONToQueryAllMarketsParamsResponse converts a zogux research json into a
// QueryAllMarketsParamsResponse.
func (h *ResearchAPIHandler) researchJSONToQueryAllMarketsParamsResponse(research zoguxtypes.ResearchJSON) (zoguxtypes.QueryAllMarketParamsResponse, error) {
	// iterate over all entries in the research json + unmarshal it's market-params
	resp := zoguxtypes.QueryAllMarketParamsResponse{}
	for _, market := range research {
		if h.onlyCMC && market.CMCID < 0 {
			continue
		}

		// convert the zogux research json market-param into a MarketParam struct
		marketParam, err := h.marketParamFromResearchJSONMarketParam(market)
		if err != nil {
			return zoguxtypes.QueryAllMarketParamsResponse{}, err
		}

		// unmarshal the market-params into a MarketParam struct
		resp.MarketParams = append(resp.MarketParams, marketParam)
	}

	return resp, nil
}

// marketParamFromResearchJSONMarketParam converts a zogux research json market-param
// into a MarketParam struct.
func (h *ResearchAPIHandler) marketParamFromResearchJSONMarketParam(marketParam zoguxtypes.Params) (zoguxtypes.MarketParam, error) {
	var exchangeConfigJSON zoguxtypes.ExchangeConfigJson
	if !h.onlyCMC {
		exchangeConfigJSON = zoguxtypes.ExchangeConfigJson{
			Exchanges: marketParam.ExchangeConfigJSON,
		}
	} else {
		exchange := zoguxtypes.ExchangeMarketConfigJson{
			ExchangeName: coinmarketcap.Name,
			Ticker:       fmt.Sprintf("%d", marketParam.CMCID),
		}
		exchangeConfigJSON = zoguxtypes.ExchangeConfigJson{
			Exchanges: []zoguxtypes.ExchangeMarketConfigJson{exchange},
		}
	}
	// marshal to a json string
	exchangeConfigJSONBz, err := json.Marshal(exchangeConfigJSON)
	if err != nil {
		return zoguxtypes.MarketParam{}, err
	}

	var minExchanges uint32
	if h.onlyCMC {
		minExchanges = 1
	} else {
		minExchanges = marketParam.MinExchanges
	}

	return zoguxtypes.MarketParam{
		Id:                 marketParam.ID,
		Pair:               marketParam.Pair,
		Exponent:           int32(marketParam.Exponent),
		MinExchanges:       minExchanges,
		MinPriceChangePpm:  marketParam.MinPriceChangePpm,
		ExchangeConfigJson: string(exchangeConfigJSONBz),
	}, nil
}
