package middleware

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"

	"github.com/overload-ak/cosmos-firewall/config"
	"github.com/overload-ak/cosmos-firewall/internal/types"
	"github.com/overload-ak/cosmos-firewall/logger"
)

type Validator struct {
	Routers *Routers
	Cfg     *config.Config
}

func NewValidator(cfg *config.Config) Validator {
	routers, err := NewRouters(cfg.Chain.ChainID)
	if err != nil {
		panic(err)
	}
	return Validator{Routers: routers, Cfg: cfg}
}

func (v Validator) IsJSONPRCRouterAllowed(router string) bool {
	for _, p := range v.Routers.GetRPCRouters() {
		if strings.EqualFold(p, router) {
			return true
		}
	}
	return false
}

func (v Validator) IsGRPCRouterAllowed(router string) bool {
	for _, p := range v.Routers.GetGRPCRouters() {
		if strings.EqualFold(p, router) {
			return true
		}
	}
	return false
}

func (v Validator) IsRESTRouterAllowed(router string) bool {
	patterns := make([]types.PathPattern, 0, len(v.Routers.GetRESTRouters()))
	for _, p := range v.Routers.GetRESTRouters() {
		patterns = append(patterns, types.NewPathPattern(p))
	}
	for _, pattern := range patterns {
		if pattern.Match(router) {
			return true
		}
	}
	return false
}

func (v Validator) CheckTxBytes(txBytes []byte) error {
	txRaw := tx.TxRaw{}
	if err := proto.Unmarshal(txBytes, &txRaw); err != nil {
		return errors.Wrapf(err, "proto unmarshal txBytes")
	}
	if len(txRaw.Signatures) < v.Cfg.Chain.MinimumSignatures {
		return errors.New("signatures is empty")
	}
	for _, signature := range txRaw.Signatures {
		if len(signature) != 64 && len(signature) != 65 {
			return errors.New("signature format error")
		}
	}
	authInfo := tx.AuthInfo{}
	if err := proto.Unmarshal(txRaw.AuthInfoBytes, &authInfo); err != nil {
		return errors.Wrapf(err, "proto unmarshal authInfo")
	}
	if authInfo.Fee.GasLimit < v.Cfg.Chain.MinimumGasLimit {
		return errors.New("GasLimit is too small")
	}
	if err := v.CheckTxAuthInfo(authInfo); err != nil {
		return errors.Wrapf(err, "check txAuthInfo")
	}
	txBody := tx.TxBody{}
	if err := proto.Unmarshal(txRaw.BodyBytes, &txBody); err != nil {
		return errors.Wrapf(err, "proto unmarshal txBody")
	}
	if !checkWhiteRouters(txBody, v.Cfg.Chain.WhiteRouters) {
		fee := v.Cfg.Chain.GetMinFee()
		if authInfo.Fee == nil || !authInfo.Fee.Amount.IsAnyGTE(fee) {
			logger.Warnf("==> fee is too low expect: %s, actual:%s", authInfo.Fee.Amount.String(), fee.String())
			return errors.New("fee is too low")
		}
	}
	if err := v.CheckTxBody(txBody); err != nil {
		return errors.Wrapf(err, "check txBody")
	}
	return nil
}

func (v Validator) CheckTxBody(txBody tx.TxBody) error {
	if len(txBody.Memo) > v.Cfg.Chain.MaxMemo {
		return errors.New("memo field length exceeds limit")
	}
	if v.Cfg.Chain.ExtensionOptions == 0 && len(txBody.ExtensionOptions) > 0 {
		return errors.New("fill in illegal field ExtensionOptions")
	}
	if v.Cfg.Chain.NonCriticalExtensionOptions == 0 && len(txBody.NonCriticalExtensionOptions) > 0 {
		return errors.New("fill in illegal field NonCriticalExtensionOptions")
	}
	if len(txBody.Messages) <= 0 {
		return errors.New("transaction message is empty")
	}
	for _, message := range txBody.Messages {
		if message.TypeUrl == "" {
			return errors.New("message type url is empty")
		}
		if !v.IsGRPCRouterAllowed(message.TypeUrl) {
			return errors.New("unsupported transaction message type")
		}
	}
	return nil
}

func (v Validator) CheckTxAuthInfo(authInfo tx.AuthInfo) error {
	if v.Cfg.Chain.Granter == 0 && authInfo.Fee.Granter != "" {
		return errors.New("fill in illegal field Granter")
	}
	if v.Cfg.Chain.Payer == 0 && authInfo.Fee.Payer != "" {
		return errors.New("set Payer, non-normal client request")
	}
	if len(authInfo.SignerInfos) < v.Cfg.Chain.SignerInfos {
		return errors.New("multiple SignerInfos, non-normal client request")
	}
	for _, info := range authInfo.SignerInfos {
		if info.PublicKey == nil {
			return errors.New("public key is empty")
		}
		if !checkPublicKeyTypeUrl(info.PublicKey.TypeUrl, v.Cfg.Chain.PublicKeyTypeURL) {
			return errors.New("illegal public key type")
		}
		if len(info.PublicKey.Value) != 35 {
			return errors.New("public key format error")
		}
		if single, ok := info.ModeInfo.Sum.(*tx.ModeInfo_Single_); !ok {
			return errors.New("invalid signature sum")
		} else {
			if single.Single.Mode == signing.SignMode_SIGN_MODE_UNSPECIFIED || single.Single.Mode == signing.SignMode_SIGN_MODE_TEXTUAL {
				return errors.New("signature mode error")
			}
		}
	}
	return nil
}

func checkPublicKeyTypeUrl(typeUrl string, publicKeyTypeURL []string) bool {
	for _, url := range publicKeyTypeURL {
		if strings.EqualFold(url, typeUrl) {
			return true
		}
	}
	return false
}

func checkWhiteRouters(txBody tx.TxBody, whiteRouters []string) bool {
	for _, message := range txBody.Messages {
		for _, router := range whiteRouters {
			if strings.EqualFold(message.TypeUrl, router) {
				return true
			}
		}
	}
	return false
}
