package handler

import (
	"encoding/base64"
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/overload-ak/cosmos-firewall/internal/middleware"
	"github.com/overload-ak/cosmos-firewall/internal/types"
	"github.com/overload-ak/cosmos-firewall/logger"
)

func TransparentHandler(validator middleware.Validator, forwarder middleware.Forwarder) grpc.StreamHandler {
	streamer := &handler{validator: validator, forwarder: forwarder}
	return streamer.handler
}

type handler struct {
	validator middleware.Validator
	forwarder middleware.Forwarder
}

func (h *handler) handler(_ interface{}, serverStream grpc.ServerStream) error {
	fullMethodName, ok := grpc.MethodFromServerStream(serverStream)
	if !ok {
		return status.Errorf(codes.Internal, "lowLevelServerStream not exists in context")
	}
	f := &types.Frame{}
	if err := serverStream.RecvMsg(f); err != nil {
		return err
	}
	if err := h.processRequest(serverStream, fullMethodName, f); err != nil {
		return err
	}
	if h.forwarder.Enable() {
		return h.forwarder.GrpcForward(serverStream, fullMethodName, f)
	}
	if err := serverStream.SendMsg(nil); err != nil {
		return err
	}
	return nil
}

func (h *handler) processRequest(serverStream grpc.ServerStream, fullMethodName string, frame *types.Frame) error {
	body := frame.Payload
	logger.Infof("GRPC RequestURI: [%s]", fullMethodName)
	logger.Info("GRPC request body base64: ", base64.StdEncoding.EncodeToString(body))
	url := fullMethodName

	if !h.validator.IsGRPCRouterAllowed(url) {
		return errors.New("method not allowed")
	}
	var err error
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
					return errors.Wrapf(err, "unmarshal error: %s", err.Error())
				}
				simulateReq.TxBytes = req1.TxBytes
			}
		}
		if err != nil {
			return errors.Wrapf(err, "unmarshal error: %s", err.Error())
		}
		if simulateReq.Tx != nil {
			for _, signature := range simulateReq.Tx.Signatures {
				if len(signature) != 64 && len(signature) != 65 {
					return errors.Wrapf(err, "unmarshal error: %s", err.Error())
				}
			}
			if err = h.validator.CheckTxAuthInfo(*simulateReq.Tx.AuthInfo); err != nil {
				return errors.Wrapf(err, "unmarshal error: %s", err.Error())
			}
			if err = h.validator.CheckTxBody(*simulateReq.Tx.Body); err != nil {
				return errors.Wrapf(err, "unmarshal error: %s", err.Error())
			}
		}
		if simulateReq.TxBytes != nil {
			if err = h.validator.CheckTxBytes(simulateReq.TxBytes); err != nil {
				return errors.Wrapf(err, "unmarshal error: %s", err.Error())
			}
		}
	case "/cosmos.tx.v1beta1.Service/BroadcastTx":
		txRequest := new(tx.BroadcastTxRequest)
		if err = proto.Unmarshal(body, txRequest); err != nil {
			return errors.Wrapf(err, "unmarshal error: %s", err.Error())
		}
		switch txRequest.Mode {
		case tx.BroadcastMode_BROADCAST_MODE_UNSPECIFIED:
		case tx.BroadcastMode_BROADCAST_MODE_BLOCK:
		case tx.BroadcastMode_BROADCAST_MODE_SYNC:
		case tx.BroadcastMode_BROADCAST_MODE_ASYNC:
		}
		if err = h.validator.CheckTxBytes(txRequest.TxBytes); err != nil {
			return err
		}
	}
	return nil
}
