package handler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/rpc/jsonrpc/server"
	"github.com/tendermint/tendermint/rpc/jsonrpc/types"

	"github.com/overload-ak/cosmos-firewall/internal/middleware"
	"github.com/overload-ak/cosmos-firewall/logger"
)

func JSONRPCHandler(validator middleware.Validator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			jsonRpcResponse(w, http.StatusInternalServerError, types.RPCInvalidParamsError(nil, err))
			return
		}
		logger.Infof("JSONRPC Method: [%s], RequestURI: [%s]", r.Method, r.URL.RequestURI())
		logger.Info("JSONRPC request body base64: ", base64.StdEncoding.EncodeToString(body))
		path := r.URL.Path
		if !validator.IsJSONPRCRouterAllowed(path) {
			jsonRpcResponse(w, http.StatusMethodNotAllowed, types.RPCMethodNotFoundError(nil))
			return
		}
		if len(path) < 1 {
			var requests []types.RPCRequest
			if err = json.Unmarshal(body, &requests); err != nil {
				var request types.RPCRequest
				if err = json.Unmarshal(body, &request); err != nil {
					jsonRpcResponse(w, http.StatusInternalServerError, types.RPCParseError(err))
					return
				}
				requests = []types.RPCRequest{request}
			}
			for _, rpcRequest := range requests {
				request := rpcRequest
				if request.ID == nil {
					logger.Debug(
						"HTTPJSONRPC received a notification, skipping... (please send a non-empty ID if you want to call a method)",
						"req", request,
					)
					continue
				}
				if len(request.Params) > 0 {
					if request.Method == "broadcast_tx_commit" || request.Method == "check_tx" ||
						request.Method == "broadcast_tx_sync" || request.Method == "broadcast_tx_async" {
						txBytes, err := getTxBytesFromParams(request.Params)
						if err != nil {
							jsonRpcResponse(w, http.StatusInternalServerError, types.RPCInvalidParamsError(request.ID, err))
							return
						}
						if err = validator.CheckTxBytes(txBytes); err != nil {
							jsonRpcResponse(w, http.StatusInternalServerError, types.RPCInternalError(nil, err))
							return
						}
					}
				}
			}
		}
		// todo success response
		jsonRpcResponse(w, http.StatusOK, types.NewRPCSuccessResponse(nil, "success"))
	}
}

func getTxBytesFromParams(data json.RawMessage) ([]byte, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err == nil {
		rawTx, ok := raw["tx"]
		if !ok {
			return nil, fmt.Errorf("undefined transaction type. params: %v. Expected map ", rawTx)
		}
		var txBytes []byte
		if err := json.Unmarshal(rawTx, &txBytes); err != nil {
			return nil, fmt.Errorf("json unmarshal txBytes error: %s", err.Error())
		}
		return txBytes, nil
	} else {
		logger.Warnf("json unmarshal raw error: %s", err.Error())
	}
	var raws []json.RawMessage
	if err := json.Unmarshal(data, &raws); err == nil {
		var txBytes []byte
		if len(raws) > 0 {
			if err := json.Unmarshal(raws[0], &txBytes); err != nil {
				return nil, fmt.Errorf("json unmarshal raws txBytes error: %s", err.Error())
			}
		}
		return txBytes, nil
	} else {
		logger.Warnf("json unmarshal raws error: %s", err.Error())
	}
	return nil, errors.New("unknown type tx raw message")
}

func jsonRpcResponse(writer http.ResponseWriter, code int, res types.RPCResponse) {
	if code != http.StatusOK {
		if err := server.WriteRPCResponseHTTPError(writer, code, res); err != nil {
			logger.Error("failed to write response", "res", res, "err", err)
		}
		return
	}
	if err := server.WriteRPCResponseHTTP(writer, types.NewRPCSuccessResponse(types.JSONRPCIntID(0), res)); err != nil {
		logger.Error("failed to write response", "res", res, "err", err)
	}
}
