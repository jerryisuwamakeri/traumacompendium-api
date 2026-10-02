package fakes

import (
	"context"
	"sync"

	"github.com/traumacompendium/backend/internal/notification/paystack"
)

// PaystackClient is a controllable fake paystack.Client.
type PaystackClient struct {
	mu sync.Mutex

	InitializeResult paystack.InitializeResult
	InitializeErr    error
	InitializeCalls  []paystack.InitializeRequest

	// VerifyStatus is returned for any reference VerifyTransaction is
	// called with, unless VerifyErr is set.
	VerifyStatus paystack.TransactionStatus
	VerifyErr    error
	VerifyCalls  []string
}

func NewPaystackClient() *PaystackClient {
	return &PaystackClient{VerifyStatus: paystack.TransactionSuccess}
}

func (f *PaystackClient) InitializeTransaction(_ context.Context, req paystack.InitializeRequest) (paystack.InitializeResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.InitializeCalls = append(f.InitializeCalls, req)
	if f.InitializeErr != nil {
		return paystack.InitializeResult{}, f.InitializeErr
	}
	if f.InitializeResult.Reference == "" {
		return paystack.InitializeResult{AuthorizationURL: "https://checkout.paystack.test/" + req.Reference, Reference: req.Reference}, nil
	}
	return f.InitializeResult, nil
}

func (f *PaystackClient) VerifyTransaction(_ context.Context, reference string) (paystack.VerifyResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.VerifyCalls = append(f.VerifyCalls, reference)
	if f.VerifyErr != nil {
		return paystack.VerifyResult{}, f.VerifyErr
	}
	return paystack.VerifyResult{Reference: reference, Status: f.VerifyStatus}, nil
}
