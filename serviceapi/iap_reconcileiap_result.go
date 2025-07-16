package codec

import (
	"fmt"
)

type ReconcileIAPResult struct {
	XPID    XPID
	IAPData IAPData
}

type IAPData struct {
	Balance       IAPBalance `json:"balance"`
	TransactionId int64      `json:"transactionid"`
}

type IAPBalance struct {
	Currency IAPCurrency `json:"currency"`
}

type IAPCurrency struct {
	EchoPoints IAPEchoPoints `json:"echopoints"`
}

type IAPEchoPoints struct {
	Value int64 `json:"val"`
}

// ReconcileIAPResult represents a response related to in-app purchases.

func NewReconcileIAPResult(userID XPID) *ReconcileIAPResult {
	return &ReconcileIAPResult{
		XPID: userID,
		IAPData: IAPData{
			Balance: IAPBalance{
				Currency: IAPCurrency{
					EchoPoints: IAPEchoPoints{
						Value: 0,
					},
				},
			},
			TransactionId: 1,
		},
	}
}

func (r *ReconcileIAPResult) String() string {
	return fmt.Sprintf("%T(user_id=%v, iap_data=%v)", r, r.XPID, r.IAPData)
}

func (r *ReconcileIAPResult) Stream(s *EasyStream) error {
	return RunErrorFunctions([]func() error{
		func() error { return s.StreamStruct(&r.XPID) },
		func() error { return s.StreamJson(&r.IAPData, true, NoCompression) },
	})
}
