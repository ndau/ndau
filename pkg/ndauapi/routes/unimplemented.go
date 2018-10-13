package routes

import (
	"net/http"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
)

type BalanceHistoryItem struct {
	Timestamp string
	Balance   math.Ndau
	TxHash    string
}

type ChaosHistoryResponse struct {
}

type OrderHistoryRecord struct {
	Timestamp math.Timestamp
	OrderInfo OrderChainInfo
}

type PreparedTx struct {
	TxData        string
	SignableBytes string
	Signature     string
}

type TxResult struct {
	TxHash     string
	ResultCode int
	ErrorMsg   string
}

type TransactionData struct {
}

type TxChangeValidationRequest struct {
}

type TxChangeSettlementRequest struct {
}

type TxClaimAccountRequest struct {
}

type TxClaimNodeRewardsRequest struct {
}

type TxCreditEAIRequest struct {
}

type TxDelegateRequest struct {
}

type TxLockRequest struct {
}

type TxNominateNodeRewardRequest struct {
}

type TxNotifyRequest struct {
}

type TxRegisterNodeRequest struct {
}

type TxReleaseFromEndowmentRequest struct {
}

type TxSetRewardsDestRequest struct {
}

type TxStakeRequest struct {
}

type TxTransferRequest struct {
}

type TxTransferAndLockRequest struct {
}

func HandleAccount(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqres.RespondJSON(w, reqres.NewAPIError("unimplemented endpoint", http.StatusInternalServerError))
	}
}

func HandleAccounts(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqres.RespondJSON(w, reqres.NewAPIError("unimplemented endpoint", http.StatusInternalServerError))
	}
}

func HandleBlockHash(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqres.RespondJSON(w, reqres.NewAPIError("unimplemented endpoint", http.StatusInternalServerError))
	}
}

func HandleBlockHeight(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqres.RespondJSON(w, reqres.NewAPIError("unimplemented endpoint", http.StatusInternalServerError))
	}
}

func HandleBlockRange(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqres.RespondJSON(w, reqres.NewAPIError("unimplemented endpoint", http.StatusInternalServerError))
	}
}

func HandleNumUnconfirmedTxs(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqres.RespondJSON(w, reqres.NewAPIError("unimplemented endpoint", http.StatusInternalServerError))
	}
}

func HandleOrderCurrent(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqres.RespondJSON(w, reqres.NewAPIError("unimplemented endpoint", http.StatusInternalServerError))
	}
}

func HandleOrderHash(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqres.RespondJSON(w, reqres.NewAPIError("unimplemented endpoint", http.StatusInternalServerError))
	}
}

func HandleOrderHeight(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqres.RespondJSON(w, reqres.NewAPIError("unimplemented endpoint", http.StatusInternalServerError))
	}
}

func HandleOrderHistory(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqres.RespondJSON(w, reqres.NewAPIError("unimplemented endpoint", http.StatusInternalServerError))
	}
}

func HandleTransactionFetch(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqres.RespondJSON(w, reqres.NewAPIError("unimplemented endpoint", http.StatusInternalServerError))
	}
}

func HandleChangeValidation(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqres.RespondJSON(w, reqres.NewAPIError("unimplemented endpoint", http.StatusInternalServerError))
	}
}

func HandleChangeSettlement(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqres.RespondJSON(w, reqres.NewAPIError("unimplemented endpoint", http.StatusInternalServerError))
	}
}

func HandleClaimAccount(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqres.RespondJSON(w, reqres.NewAPIError("unimplemented endpoint", http.StatusInternalServerError))
	}
}

func HandleClaimNodeRewards(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqres.RespondJSON(w, reqres.NewAPIError("unimplemented endpoint", http.StatusInternalServerError))
	}
}

func HandleCreditEAI(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqres.RespondJSON(w, reqres.NewAPIError("unimplemented endpoint", http.StatusInternalServerError))
	}
}

func HandleDelegate(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqres.RespondJSON(w, reqres.NewAPIError("unimplemented endpoint", http.StatusInternalServerError))
	}
}

func HandleLock(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqres.RespondJSON(w, reqres.NewAPIError("unimplemented endpoint", http.StatusInternalServerError))
	}
}

func HandleNominateNodeReward(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqres.RespondJSON(w, reqres.NewAPIError("unimplemented endpoint", http.StatusInternalServerError))
	}
}

func HandleNotify(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqres.RespondJSON(w, reqres.NewAPIError("unimplemented endpoint", http.StatusInternalServerError))
	}
}

func HandleRegisterNode(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqres.RespondJSON(w, reqres.NewAPIError("unimplemented endpoint", http.StatusInternalServerError))
	}
}

func HandleReleaseFromEndowment(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqres.RespondJSON(w, reqres.NewAPIError("unimplemented endpoint", http.StatusInternalServerError))
	}
}

func HandleSetRewardsDest(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqres.RespondJSON(w, reqres.NewAPIError("unimplemented endpoint", http.StatusInternalServerError))
	}
}

func HandleStake(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqres.RespondJSON(w, reqres.NewAPIError("unimplemented endpoint", http.StatusInternalServerError))
	}
}

func HandleTransfer(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqres.RespondJSON(w, reqres.NewAPIError("unimplemented endpoint", http.StatusInternalServerError))
	}
}

func HandleTransferAndLock(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqres.RespondJSON(w, reqres.NewAPIError("unimplemented endpoint", http.StatusInternalServerError))
	}
}

func HandleSubmitTx(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqres.RespondJSON(w, reqres.NewAPIError("unimplemented endpoint", http.StatusInternalServerError))
	}
}
