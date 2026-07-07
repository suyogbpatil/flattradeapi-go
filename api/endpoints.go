package api

const (
	DefaultBaseURL            = "https://piconnect.flattrade.in/PiConnectAPI"
	DefaultLoginURL           = "https://auth.flattrade.in/"
	DefaultSessionURL         = "https://authapi.flattrade.in/trade/apitoken"
	DefaultWebSocketURL       = "wss://piconnect.flattrade.in/PiConnectWSAPI/"
	DefaultScripMasterBaseURL = "https://flattrade.s3.ap-south-1.amazonaws.com/scripmaster"

	loginURLAPIKeyName = "app_key"
)

const (
	EndpointPlaceOrder         = "/PlaceOrder"
	EndpointModifyOrder        = "/ModifyOrder"
	EndpointCancelOrder        = "/CancelOrder"
	EndpointExitSNOOrder       = "/ExitSNOOrder"
	EndpointOrderMargin        = "/GetOrderMargin"
	EndpointBasketMargin       = "/GetBasketMargin"
	EndpointOrderBook          = "/OrderBook"
	EndpointMultiLegOrderBook  = "/MultiLegOrderBook"
	EndpointSingleOrderHistory = "/SingleOrdHist"
	EndpointTradeBook          = "/TradeBook"
	EndpointPositionBook       = "/PositionBook"
	EndpointProductConversion  = "/ProductConversion"
)

const (
	EndpointPlaceGTTOrder    = "/PlaceGTTOrder"
	EndpointModifyGTTOrder   = "/ModifyGTTOrder"
	EndpointCancelGTTOrder   = "/CancelGTTOrder"
	EndpointPendingGTTOrders = "/GetPendingGTTOrder"
	EndpointEnabledGTTs      = "/GetEnabledGTTs"
	EndpointPlaceOCOOrder    = "/PlaceOCOOrder"
	EndpointModifyOCOOrder   = "/ModifyOCOOrder"
	EndpointCancelOCOOrder   = "/CancelOCOOrder"
)

const (
	EndpointHoldings = "/Holdings"
	EndpointLimits   = "/Limits"
)

const (
	EndpointIndexList       = "/GetIndexList"
	EndpointTopListNames    = "/TopListName"
	EndpointTopList         = "/TopList"
	EndpointTimePriceSeries = "/TPSeries"
	EndpointEODChartData    = "/EODChartData"
	EndpointOptionChain     = "/GetOptionChain"
	EndpointOptionGreek     = "/GetOptionGreek"
	EndpointExchangeMessage = "/ExchMsg"
	EndpointBrokerMessage   = "/GetBrokerMsg"
	EndpointSpanCalculator  = "/SpanCalc"
)

const (
	EndpointSetAlert          = "/SetAlert"
	EndpointCancelAlert       = "/CancelAlert"
	EndpointModifyAlert       = "/ModifyAlert"
	EndpointPendingAlert      = "/GetPendingAlert"
	EndpointEnabledAlertTypes = "/GetEnabledAlertTypes"
)

const (
	EndpointMaxPayoutAmount = "/GetMaxPayoutAmount"
	EndpointFundsPayout     = "/FundsPayOutReq"
	EndpointPayinReport     = "/GetPayinReport"
	EndpointPayoutReport    = "/GetPayoutReport"
	EndpointCancelPayout    = "/CancelPayout"
)

const (
	EndpointUserDetails = "/UserDetails"
	EndpointSearchScrip = "/SearchScrip"
	EndpointGetQuotes   = "/GetQuotes"
)

const (
	ScripMasterNSE       = DefaultScripMasterBaseURL + "/NSE_Equity.csv"
	ScripMasterBSE       = DefaultScripMasterBaseURL + "/BSE_Equity.csv"
	ScripMasterNFOEquity = DefaultScripMasterBaseURL + "/Nfo_Equity_Derivatives.csv"
	ScripMasterNFOIndex  = DefaultScripMasterBaseURL + "/Nfo_Index_Derivatives.csv"
	ScripMasterBFOEquity = DefaultScripMasterBaseURL + "/Bfo_Equity_Derivatives.csv"
	ScripMasterBFOIndex  = DefaultScripMasterBaseURL + "/Bfo_Index_Derivatives.csv"
	ScripMasterCurrency  = DefaultScripMasterBaseURL + "/Currency_Derivatives.csv"
	ScripMasterCommodity = DefaultScripMasterBaseURL + "/Commodity.csv"
)
