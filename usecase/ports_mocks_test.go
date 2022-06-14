// Code generated by MockGen. DO NOT EDIT.
// Source: ports.go

// Package usecase_test is a generated GoMock package.
package usecase_test

import (
	context "context"
	reflect "reflect"

	database "github.com/emerishq/demeris-api-server/api/database"
	cns "github.com/emerishq/demeris-backend-models/cns"
	tracelistener "github.com/emerishq/demeris-backend-models/tracelistener"
	sdkutilities "github.com/emerishq/sdk-service-meta/gen/sdk_utilities"
	gomock "github.com/golang/mock/gomock"
)

// MockDB is a mock of DB interface.
type MockDB struct {
	ctrl     *gomock.Controller
	recorder *MockDBMockRecorder
}

// MockDBMockRecorder is the mock recorder for MockDB.
type MockDBMockRecorder struct {
	mock *MockDB
}

// NewMockDB creates a new mock instance.
func NewMockDB(ctrl *gomock.Controller) *MockDB {
	mock := &MockDB{ctrl: ctrl}
	mock.recorder = &MockDBMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDB) EXPECT() *MockDBMockRecorder {
	return m.recorder
}

// Balances mocks base method.
func (m *MockDB) Balances(ctx context.Context, addresses []string) ([]tracelistener.BalanceRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Balances", ctx, addresses)
	ret0, _ := ret[0].([]tracelistener.BalanceRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Balances indicates an expected call of Balances.
func (mr *MockDBMockRecorder) Balances(ctx, addresses interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Balances", reflect.TypeOf((*MockDB)(nil).Balances), ctx, addresses)
}

// Chains mocks base method.
func (m *MockDB) Chains(ctx context.Context) ([]cns.Chain, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Chains", ctx)
	ret0, _ := ret[0].([]cns.Chain)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Chains indicates an expected call of Chains.
func (mr *MockDBMockRecorder) Chains(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Chains", reflect.TypeOf((*MockDB)(nil).Chains), ctx)
}

// Delegations mocks base method.
func (m *MockDB) Delegations(ctx context.Context, addresses []string) ([]database.DelegationResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delegations", ctx, addresses)
	ret0, _ := ret[0].([]database.DelegationResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Delegations indicates an expected call of Delegations.
func (mr *MockDBMockRecorder) Delegations(ctx, addresses interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delegations", reflect.TypeOf((*MockDB)(nil).Delegations), ctx, addresses)
}

// DenomTrace mocks base method.
func (m *MockDB) DenomTrace(ctx context.Context, chain, hash string) (tracelistener.IBCDenomTraceRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DenomTrace", ctx, chain, hash)
	ret0, _ := ret[0].(tracelistener.IBCDenomTraceRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DenomTrace indicates an expected call of DenomTrace.
func (mr *MockDBMockRecorder) DenomTrace(ctx, chain, hash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DenomTrace", reflect.TypeOf((*MockDB)(nil).DenomTrace), ctx, chain, hash)
}

// VerifiedDenoms mocks base method.
func (m *MockDB) VerifiedDenoms(arg0 context.Context) (map[string]cns.DenomList, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "VerifiedDenoms", arg0)
	ret0, _ := ret[0].(map[string]cns.DenomList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// VerifiedDenoms indicates an expected call of VerifiedDenoms.
func (mr *MockDBMockRecorder) VerifiedDenoms(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "VerifiedDenoms", reflect.TypeOf((*MockDB)(nil).VerifiedDenoms), arg0)
}

// MockSDKServiceClients is a mock of SDKServiceClients interface.
type MockSDKServiceClients struct {
	ctrl     *gomock.Controller
	recorder *MockSDKServiceClientsMockRecorder
}

// MockSDKServiceClientsMockRecorder is the mock recorder for MockSDKServiceClients.
type MockSDKServiceClientsMockRecorder struct {
	mock *MockSDKServiceClients
}

// NewMockSDKServiceClients creates a new mock instance.
func NewMockSDKServiceClients(ctrl *gomock.Controller) *MockSDKServiceClients {
	mock := &MockSDKServiceClients{ctrl: ctrl}
	mock.recorder = &MockSDKServiceClientsMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSDKServiceClients) EXPECT() *MockSDKServiceClientsMockRecorder {
	return m.recorder
}

// GetSDKServiceClient mocks base method.
func (m *MockSDKServiceClients) GetSDKServiceClient(version string) (sdkutilities.Service, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSDKServiceClient", version)
	ret0, _ := ret[0].(sdkutilities.Service)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSDKServiceClient indicates an expected call of GetSDKServiceClient.
func (mr *MockSDKServiceClientsMockRecorder) GetSDKServiceClient(version interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSDKServiceClient", reflect.TypeOf((*MockSDKServiceClients)(nil).GetSDKServiceClient), version)
}

// MockSDKServiceClient is a mock of SDKServiceClient interface.
type MockSDKServiceClient struct {
	ctrl     *gomock.Controller
	recorder *MockSDKServiceClientMockRecorder
}

// MockSDKServiceClientMockRecorder is the mock recorder for MockSDKServiceClient.
type MockSDKServiceClientMockRecorder struct {
	mock *MockSDKServiceClient
}

// NewMockSDKServiceClient creates a new mock instance.
func NewMockSDKServiceClient(ctrl *gomock.Controller) *MockSDKServiceClient {
	mock := &MockSDKServiceClient{ctrl: ctrl}
	mock.recorder = &MockSDKServiceClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSDKServiceClient) EXPECT() *MockSDKServiceClientMockRecorder {
	return m.recorder
}

// AccountNumbers mocks base method.
func (m *MockSDKServiceClient) AccountNumbers(arg0 context.Context, arg1 *sdkutilities.AccountNumbersPayload) (*sdkutilities.AccountNumbers2, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AccountNumbers", arg0, arg1)
	ret0, _ := ret[0].(*sdkutilities.AccountNumbers2)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AccountNumbers indicates an expected call of AccountNumbers.
func (mr *MockSDKServiceClientMockRecorder) AccountNumbers(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AccountNumbers", reflect.TypeOf((*MockSDKServiceClient)(nil).AccountNumbers), arg0, arg1)
}

// Block mocks base method.
func (m *MockSDKServiceClient) Block(arg0 context.Context, arg1 *sdkutilities.BlockPayload) (*sdkutilities.BlockData, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Block", arg0, arg1)
	ret0, _ := ret[0].(*sdkutilities.BlockData)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Block indicates an expected call of Block.
func (mr *MockSDKServiceClientMockRecorder) Block(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Block", reflect.TypeOf((*MockSDKServiceClient)(nil).Block), arg0, arg1)
}

// BroadcastTx mocks base method.
func (m *MockSDKServiceClient) BroadcastTx(arg0 context.Context, arg1 *sdkutilities.BroadcastTxPayload) (*sdkutilities.TransactionResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BroadcastTx", arg0, arg1)
	ret0, _ := ret[0].(*sdkutilities.TransactionResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BroadcastTx indicates an expected call of BroadcastTx.
func (mr *MockSDKServiceClientMockRecorder) BroadcastTx(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BroadcastTx", reflect.TypeOf((*MockSDKServiceClient)(nil).BroadcastTx), arg0, arg1)
}

// BudgetParams mocks base method.
func (m *MockSDKServiceClient) BudgetParams(arg0 context.Context, arg1 *sdkutilities.BudgetParamsPayload) (*sdkutilities.BudgetParams2, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BudgetParams", arg0, arg1)
	ret0, _ := ret[0].(*sdkutilities.BudgetParams2)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BudgetParams indicates an expected call of BudgetParams.
func (mr *MockSDKServiceClientMockRecorder) BudgetParams(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BudgetParams", reflect.TypeOf((*MockSDKServiceClient)(nil).BudgetParams), arg0, arg1)
}

// DelegatorRewards mocks base method.
func (m *MockSDKServiceClient) DelegatorRewards(arg0 context.Context, arg1 *sdkutilities.DelegatorRewardsPayload) (*sdkutilities.DelegatorRewards2, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DelegatorRewards", arg0, arg1)
	ret0, _ := ret[0].(*sdkutilities.DelegatorRewards2)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DelegatorRewards indicates an expected call of DelegatorRewards.
func (mr *MockSDKServiceClientMockRecorder) DelegatorRewards(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DelegatorRewards", reflect.TypeOf((*MockSDKServiceClient)(nil).DelegatorRewards), arg0, arg1)
}

// DistributionParams mocks base method.
func (m *MockSDKServiceClient) DistributionParams(arg0 context.Context, arg1 *sdkutilities.DistributionParamsPayload) (*sdkutilities.DistributionParams2, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DistributionParams", arg0, arg1)
	ret0, _ := ret[0].(*sdkutilities.DistributionParams2)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DistributionParams indicates an expected call of DistributionParams.
func (mr *MockSDKServiceClientMockRecorder) DistributionParams(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DistributionParams", reflect.TypeOf((*MockSDKServiceClient)(nil).DistributionParams), arg0, arg1)
}

// EmoneyInflation mocks base method.
func (m *MockSDKServiceClient) EmoneyInflation(arg0 context.Context, arg1 *sdkutilities.EmoneyInflationPayload) (*sdkutilities.EmoneyInflation2, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EmoneyInflation", arg0, arg1)
	ret0, _ := ret[0].(*sdkutilities.EmoneyInflation2)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EmoneyInflation indicates an expected call of EmoneyInflation.
func (mr *MockSDKServiceClientMockRecorder) EmoneyInflation(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EmoneyInflation", reflect.TypeOf((*MockSDKServiceClient)(nil).EmoneyInflation), arg0, arg1)
}

// EstimateFees mocks base method.
func (m *MockSDKServiceClient) EstimateFees(arg0 context.Context, arg1 *sdkutilities.EstimateFeesPayload) (*sdkutilities.Simulation, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EstimateFees", arg0, arg1)
	ret0, _ := ret[0].(*sdkutilities.Simulation)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EstimateFees indicates an expected call of EstimateFees.
func (mr *MockSDKServiceClientMockRecorder) EstimateFees(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EstimateFees", reflect.TypeOf((*MockSDKServiceClient)(nil).EstimateFees), arg0, arg1)
}

// LiquidityParams mocks base method.
func (m *MockSDKServiceClient) LiquidityParams(arg0 context.Context, arg1 *sdkutilities.LiquidityParamsPayload) (*sdkutilities.LiquidityParams2, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LiquidityParams", arg0, arg1)
	ret0, _ := ret[0].(*sdkutilities.LiquidityParams2)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LiquidityParams indicates an expected call of LiquidityParams.
func (mr *MockSDKServiceClientMockRecorder) LiquidityParams(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LiquidityParams", reflect.TypeOf((*MockSDKServiceClient)(nil).LiquidityParams), arg0, arg1)
}

// LiquidityPools mocks base method.
func (m *MockSDKServiceClient) LiquidityPools(arg0 context.Context, arg1 *sdkutilities.LiquidityPoolsPayload) (*sdkutilities.LiquidityPools2, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LiquidityPools", arg0, arg1)
	ret0, _ := ret[0].(*sdkutilities.LiquidityPools2)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LiquidityPools indicates an expected call of LiquidityPools.
func (mr *MockSDKServiceClientMockRecorder) LiquidityPools(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LiquidityPools", reflect.TypeOf((*MockSDKServiceClient)(nil).LiquidityPools), arg0, arg1)
}

// MintAnnualProvision mocks base method.
func (m *MockSDKServiceClient) MintAnnualProvision(arg0 context.Context, arg1 *sdkutilities.MintAnnualProvisionPayload) (*sdkutilities.MintAnnualProvision2, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MintAnnualProvision", arg0, arg1)
	ret0, _ := ret[0].(*sdkutilities.MintAnnualProvision2)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MintAnnualProvision indicates an expected call of MintAnnualProvision.
func (mr *MockSDKServiceClientMockRecorder) MintAnnualProvision(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MintAnnualProvision", reflect.TypeOf((*MockSDKServiceClient)(nil).MintAnnualProvision), arg0, arg1)
}

// MintEpochProvisions mocks base method.
func (m *MockSDKServiceClient) MintEpochProvisions(arg0 context.Context, arg1 *sdkutilities.MintEpochProvisionsPayload) (*sdkutilities.MintEpochProvisions2, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MintEpochProvisions", arg0, arg1)
	ret0, _ := ret[0].(*sdkutilities.MintEpochProvisions2)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MintEpochProvisions indicates an expected call of MintEpochProvisions.
func (mr *MockSDKServiceClientMockRecorder) MintEpochProvisions(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MintEpochProvisions", reflect.TypeOf((*MockSDKServiceClient)(nil).MintEpochProvisions), arg0, arg1)
}

// MintInflation mocks base method.
func (m *MockSDKServiceClient) MintInflation(arg0 context.Context, arg1 *sdkutilities.MintInflationPayload) (*sdkutilities.MintInflation2, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MintInflation", arg0, arg1)
	ret0, _ := ret[0].(*sdkutilities.MintInflation2)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MintInflation indicates an expected call of MintInflation.
func (mr *MockSDKServiceClientMockRecorder) MintInflation(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MintInflation", reflect.TypeOf((*MockSDKServiceClient)(nil).MintInflation), arg0, arg1)
}

// MintParams mocks base method.
func (m *MockSDKServiceClient) MintParams(arg0 context.Context, arg1 *sdkutilities.MintParamsPayload) (*sdkutilities.MintParams2, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MintParams", arg0, arg1)
	ret0, _ := ret[0].(*sdkutilities.MintParams2)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MintParams indicates an expected call of MintParams.
func (mr *MockSDKServiceClientMockRecorder) MintParams(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MintParams", reflect.TypeOf((*MockSDKServiceClient)(nil).MintParams), arg0, arg1)
}

// QueryTx mocks base method.
func (m *MockSDKServiceClient) QueryTx(arg0 context.Context, arg1 *sdkutilities.QueryTxPayload) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryTx", arg0, arg1)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryTx indicates an expected call of QueryTx.
func (mr *MockSDKServiceClientMockRecorder) QueryTx(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryTx", reflect.TypeOf((*MockSDKServiceClient)(nil).QueryTx), arg0, arg1)
}

// StakingParams mocks base method.
func (m *MockSDKServiceClient) StakingParams(arg0 context.Context, arg1 *sdkutilities.StakingParamsPayload) (*sdkutilities.StakingParams2, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StakingParams", arg0, arg1)
	ret0, _ := ret[0].(*sdkutilities.StakingParams2)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// StakingParams indicates an expected call of StakingParams.
func (mr *MockSDKServiceClientMockRecorder) StakingParams(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StakingParams", reflect.TypeOf((*MockSDKServiceClient)(nil).StakingParams), arg0, arg1)
}

// StakingPool mocks base method.
func (m *MockSDKServiceClient) StakingPool(arg0 context.Context, arg1 *sdkutilities.StakingPoolPayload) (*sdkutilities.StakingPool2, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StakingPool", arg0, arg1)
	ret0, _ := ret[0].(*sdkutilities.StakingPool2)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// StakingPool indicates an expected call of StakingPool.
func (mr *MockSDKServiceClientMockRecorder) StakingPool(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StakingPool", reflect.TypeOf((*MockSDKServiceClient)(nil).StakingPool), arg0, arg1)
}

// Supply mocks base method.
func (m *MockSDKServiceClient) Supply(arg0 context.Context, arg1 *sdkutilities.SupplyPayload) (*sdkutilities.Supply2, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Supply", arg0, arg1)
	ret0, _ := ret[0].(*sdkutilities.Supply2)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Supply indicates an expected call of Supply.
func (mr *MockSDKServiceClientMockRecorder) Supply(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Supply", reflect.TypeOf((*MockSDKServiceClient)(nil).Supply), arg0, arg1)
}

// SupplyDenom mocks base method.
func (m *MockSDKServiceClient) SupplyDenom(arg0 context.Context, arg1 *sdkutilities.SupplyDenomPayload) (*sdkutilities.Supply2, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SupplyDenom", arg0, arg1)
	ret0, _ := ret[0].(*sdkutilities.Supply2)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SupplyDenom indicates an expected call of SupplyDenom.
func (mr *MockSDKServiceClientMockRecorder) SupplyDenom(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SupplyDenom", reflect.TypeOf((*MockSDKServiceClient)(nil).SupplyDenom), arg0, arg1)
}

// TxMetadata mocks base method.
func (m *MockSDKServiceClient) TxMetadata(arg0 context.Context, arg1 *sdkutilities.TxMetadataPayload) (*sdkutilities.TxMessagesMetadata, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TxMetadata", arg0, arg1)
	ret0, _ := ret[0].(*sdkutilities.TxMessagesMetadata)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// TxMetadata indicates an expected call of TxMetadata.
func (mr *MockSDKServiceClientMockRecorder) TxMetadata(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TxMetadata", reflect.TypeOf((*MockSDKServiceClient)(nil).TxMetadata), arg0, arg1)
}
