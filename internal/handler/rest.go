package handler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/gogo/protobuf/proto"

	"github.com/overload-ak/cosmos-firewall/internal/middleware"
	"github.com/overload-ak/cosmos-firewall/logger"
)

func RestHandler(validator middleware.Validator, forwarder middleware.Forwarder) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			restResponse(writer, http.StatusInternalServerError, "read all body error: ", nil)
			return
		}
		logger.Infof("REST Method: [%s], RequestURI: [%s]", request.Method, request.URL.RequestURI())
		logger.Info("REST request body base64: ", base64.StdEncoding.EncodeToString(body))
		url := request.URL.RequestURI()

		if !validator.IsRESTRouterAllowed(url) {
			restResponse(writer, http.StatusMethodNotAllowed, "method not allowed", nil)
			return
		}
		switch url {
		case "/cosmos/tx/v1beta1/simulate":
			simulateReq := tx.SimulateRequest{}
			if err = proto.Unmarshal(body, &simulateReq); err != nil {
				if err = json.Unmarshal(body, &simulateReq); err != nil {
					type BroadcastTxRequest struct {
						TxBytes []byte `json:"tx_bytes,omitempty"`
						Mode    string `json:"mode,omitempty"`
					}
					var req1 BroadcastTxRequest
					if err = json.Unmarshal(body, &req1); err != nil {
						restResponse(writer, http.StatusInternalServerError, fmt.Sprintf("broadcastTxRequest json unmarshal %v", err.Error()), nil)
						return
					}
					simulateReq.TxBytes = req1.TxBytes
				}
			}
			if err != nil {
				restResponse(writer, http.StatusInternalServerError, fmt.Sprintf("simulateRequest json unmarshal %v", err.Error()), nil)
				return
			}
			if simulateReq.Tx != nil {
				for _, signature := range simulateReq.Tx.Signatures {
					if len(signature) != 64 && len(signature) != 65 {
						restResponse(writer, http.StatusNonAuthoritativeInfo, "signature format error", nil)
						return
					}
				}
				if err = validator.CheckTxAuthInfo(*simulateReq.Tx.AuthInfo); err != nil {
					restResponse(writer, http.StatusUnprocessableEntity, err.Error(), nil)
					return
				}

				if err = validator.CheckTxBody(*simulateReq.Tx.Body); err != nil {
					restResponse(writer, http.StatusUnprocessableEntity, err.Error(), nil)
					return
				}
			}
			if simulateReq.TxBytes != nil {
				if err = validator.CheckTxBytes(simulateReq.TxBytes); err != nil {
					restResponse(writer, http.StatusUnprocessableEntity, err.Error(), nil)
					return
				}
			}
		case "/cosmos/tx/v1beta1/txs":
			var req tx.BroadcastTxRequest
			if err = json.Unmarshal(body, &req); err != nil {
				type BroadcastTxRequest struct {
					TxBytes []byte `json:"tx_bytes,omitempty"`
					Mode    string `json:"mode,omitempty"`
				}
				var req1 BroadcastTxRequest
				if err = json.Unmarshal(body, &req1); err != nil {
					restResponse(writer, http.StatusInternalServerError, fmt.Sprintf("json unmarshal BroadcastTxRequest error: %s", err.Error()), nil)
					return
				}
				req.TxBytes = req1.TxBytes
				req.Mode = tx.BroadcastMode(tx.BroadcastMode_value[req1.Mode])
			}
			if req.TxBytes == nil {
				restResponse(writer, http.StatusInternalServerError, "invalid empty tx bytes", nil)
				return
			}
			switch req.Mode {
			case tx.BroadcastMode_BROADCAST_MODE_UNSPECIFIED:
			case tx.BroadcastMode_BROADCAST_MODE_BLOCK:
			case tx.BroadcastMode_BROADCAST_MODE_SYNC:
			case tx.BroadcastMode_BROADCAST_MODE_ASYNC:
			}
			if err = validator.CheckTxBytes(req.TxBytes); err != nil {
				restResponse(writer, http.StatusUnprocessableEntity, err.Error(), nil)
				return
			}
		}
		if forwarder.Enable() {
			if err = forwarder.Request(middleware.RESTREQUEST, writer, request, bytes.NewReader(body)); err != nil {
				restResponse(writer, http.StatusInternalServerError, fmt.Sprintf("forwarder request error: %s", err.Error()), nil)
				return
			}
			return
		}
		restResponse(writer, http.StatusOK, "SUCCESS", nil)
	}
}

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func restResponse(writer http.ResponseWriter, code int, msg string, data interface{}) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(code)
	response := &Response{
		Code: code,
		Msg:  msg,
		Data: data,
	}
	d, err := json.Marshal(response)
	logger.Infof("REST response statusCode: %d, body: %s", code, base64.StdEncoding.EncodeToString(d))
	if err != nil {
		logger.Errorf("output json marshal error: %s", err.Error())
	} else {
		if _, err := writer.Write(d); err != nil {
			logger.Errorf("output write error: %s", err.Error())
		}
	}
}
