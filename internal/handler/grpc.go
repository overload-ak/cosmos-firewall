package handler

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/types"
	"io"
	"net/http"
	"strconv"

	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/gogo/protobuf/proto"

	"github.com/overload-ak/cosmos-firewall/internal/middleware"
	"github.com/overload-ak/cosmos-firewall/logger"
)

func GRPCHandler(validator middleware.Validator) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			grpcResponse(writer, http.StatusExpectationFailed, "read all body error: "+err.Error(), nil)
			return
		}
		logger.Infof("GRPC Method: [%s], RequestURI: [%s]", request.Method, request.URL.RequestURI())
		logger.Info("GRPC request body base64: ", base64.StdEncoding.EncodeToString(body))
		if len(body) < 5 {
			grpcResponse(writer, http.StatusExpectationFailed, "invalid body", nil)
			return
		}
		body = body[1:]     // remove payload byte
		sizeBz := body[0:4] // size byte
		dataLen := binary.BigEndian.Uint32(sizeBz)
		body = body[4:] // remove size byte
		if uint32(len(body)) < dataLen {
			grpcResponse(writer, http.StatusExpectationFailed, "invalid body", nil)
			return
		}
		body = body[0:dataLen] // data
		url := request.URL.Path

		if !validator.IsGRPCPathAllowed(url) {
			grpcResponse(writer, http.StatusMethodNotAllowed, "method not allowed", nil)
			return
		}

		switch url {
		case "/cosmos.tx.v1beta1.Service/Simulate":
			simulateReq := tx.SimulateRequest{}
			if err = proto.Unmarshal(body, &simulateReq); err != nil {
				if err = json.Unmarshal(body, &simulateReq); err != nil {
					type BroadcastTxRequest struct {
						TxBytes []byte `json:"tx_bytes,omitempty"`
						Mode    string `json:"mode,omitempty"`
					}
					var req1 BroadcastTxRequest
					if err = json.Unmarshal(body, &req1); err != nil {
						grpcResponse(writer, http.StatusUnprocessableEntity, "unmarshal error: "+err.Error(), nil)
						return
					}
					simulateReq.TxBytes = req1.TxBytes
				}
			}
			if err != nil {
				grpcResponse(writer, http.StatusUnprocessableEntity, "unmarshal error: "+err.Error(), nil)
				return
			}
			if simulateReq.Tx != nil {
				for _, signature := range simulateReq.Tx.Signatures {
					if len(signature) != 64 && len(signature) != 65 {
						grpcResponse(writer, http.StatusNonAuthoritativeInfo, "signature format error", nil)
						return
					}
				}
				if err = validator.CheckTxAuthInfo(*simulateReq.Tx.AuthInfo); err != nil {
					grpcResponse(writer, http.StatusUnprocessableEntity, err.Error(), nil)
					return
				}
				if err = validator.CheckTxBody(*simulateReq.Tx.Body); err != nil {
					grpcResponse(writer, http.StatusUnprocessableEntity, err.Error(), nil)
					return
				}
			}
			if simulateReq.TxBytes != nil {
				if err = validator.CheckTxBytes(simulateReq.TxBytes); err != nil {
					grpcResponse(writer, http.StatusUnprocessableEntity, err.Error(), nil)
					return
				}
			}
		case "/cosmos.tx.v1beta1.Service/BroadcastTx":
			txRequest := new(tx.BroadcastTxRequest)
			if err = proto.Unmarshal(body, txRequest); err != nil {
				grpcResponse(writer, http.StatusUnprocessableEntity, "unmarshal error: "+err.Error(), nil)
				return
			}
			switch txRequest.Mode {
			case tx.BroadcastMode_BROADCAST_MODE_UNSPECIFIED:
			case tx.BroadcastMode_BROADCAST_MODE_BLOCK:
			case tx.BroadcastMode_BROADCAST_MODE_SYNC:
			case tx.BroadcastMode_BROADCAST_MODE_ASYNC:
			}
			if err = validator.CheckTxBytes(txRequest.TxBytes); err != nil {
				grpcResponse(writer, http.StatusUnprocessableEntity, err.Error(), nil)
				return
			}
		}
		// todo success response
		grpcResponse(writer, 0, "ok", nil)
	}
}

func grpcResponse(writer http.ResponseWriter, code int, msg string, data interface{}) {
	writer.Header().Set("Content-Type", "application/grpc")
	writer.Header().Add("grpc-status", strconv.Itoa(code))
	writer.Header().Add("grpc-message", msg)
	writer.WriteHeader(http.StatusOK)
	// todo grpc return error
	d, err := json.Marshal(types.NewCoins(types.NewCoin("FX", types.NewInt(1000000000000000000))))
	logger.Infof("GRPC response statusCode: %d, body: %s", code, base64.StdEncoding.EncodeToString(d))
	if err != nil {
		logger.Errorf("output json marshal error: %s", err.Error())
	} else {
		if _, err := writer.Write(d); err != nil {
			logger.Errorf("output write error: %s", err.Error())
		}
	}
}
