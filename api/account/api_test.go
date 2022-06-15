package account_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emerishq/demeris-api-server/api/account"
	"github.com/emerishq/demeris-api-server/lib/fflag"
	"github.com/emerishq/emeris-utils/logging"
	"github.com/gin-gonic/gin"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mocks struct {
	app *MockApp
}

func newAccountAPI(t *testing.T, setup func(mocks)) *account.AccountAPI {
	ctrl := gomock.NewController(t)
	m := mocks{
		app: NewMockApp(ctrl),
	}
	if setup != nil {
		setup(m)
	}
	return account.New(m.app)
}

func TestGetAccounts(t *testing.T) {
	var (
		ctx  = context.Background()
		resp = account.AccountsResponse{
			Balances: []account.Balance{
				{Address: "adr1", BaseDenom: "denom1", Amount: "42"},
				{Address: "adr2", BaseDenom: "denom2", Amount: "42"},
			},
			StakingBalances: []account.StakingBalance{
				{ValidatorAddress: "adr1", ChainName: "chain1", Amount: "42"},
				{ValidatorAddress: "adr2", ChainName: "chain1", Amount: "42"},
			},
		}
		respJSON, _ = json.Marshal(resp)
	)
	tests := []struct {
		name               string
		expectedStatusCode int
		expectedBody       string
		expectedError      string
		setup              func(mocks)
	}{
		{
			name:               "ok",
			expectedStatusCode: http.StatusOK,
			expectedBody:       string(respJSON),

			setup: func(m mocks) {
				adrs := []string{"adr1", "adr2"}
				m.app.EXPECT().DeriveRawAddress(ctx, "xxx").Return(adrs, nil)
				m.app.EXPECT().Balances(ctx, adrs).Return(resp.Balances, nil)
				m.app.EXPECT().StakingBalances(ctx, adrs).Return(resp.StakingBalances, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			assert := assert.New(t)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "rawaddress", Value: "xxx"}}
			c.Request, _ = http.NewRequestWithContext(ctx, http.MethodGet, "", nil)
			// add logger or else it fails
			logger := logging.New(logging.LoggingConfig{})
			c.Set(logging.LoggerKey, logger)
			ac := newAccountAPI(t, tt.setup)

			ac.GetAccounts(c)

			assert.Equal(tt.expectedStatusCode, w.Code)
			if tt.expectedError != "" {
				require.Len(c.Errors, 1, "expected one error but got %d", len(c.Errors))
				require.EqualError(c.Errors[0], tt.expectedError)
				return
			}
			require.Empty(c.Errors)
			assert.JSONEq(tt.expectedBody, w.Body.String())
		})
	}
}

func TestGetBalancesPerAddress(t *testing.T) {
	var (
		ctx  = context.Background()
		resp = account.BalancesResponse{
			Balances: []account.Balance{
				{Address: "adr1", BaseDenom: "denom1", Amount: "42"},
				{Address: "adr2", BaseDenom: "denom2", Amount: "42"},
			},
		}
		respJSON, _ = json.Marshal(resp)
	)
	tests := []struct {
		name               string
		expectedStatusCode int
		expectedBody       string
		expectedError      string
		setup              func(mocks)
	}{
		{
			name:               "ok",
			expectedStatusCode: http.StatusOK,
			expectedBody:       string(respJSON),

			setup: func(m mocks) {
				m.app.EXPECT().Balances(ctx, []string{"adr1"}).Return(resp.Balances, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			assert := assert.New(t)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "address", Value: "adr1"}}
			c.Request, _ = http.NewRequestWithContext(ctx, http.MethodGet, "", nil)
			// add logger or else it fails
			logger := logging.New(logging.LoggingConfig{})
			c.Set(logging.LoggerKey, logger)
			ac := newAccountAPI(t, tt.setup)

			ac.GetBalancesByAddress(c)

			assert.Equal(tt.expectedStatusCode, w.Code)
			if tt.expectedError != "" {
				require.Len(c.Errors, 1, "expected one error but got %d", len(c.Errors))
				require.EqualError(c.Errors[0], tt.expectedError)
				return
			}
			require.Empty(c.Errors)
			assert.JSONEq(tt.expectedBody, w.Body.String())
		})
	}

}

func TestGetDelegationsPerAddress(t *testing.T) {
	var (
		ctx  = context.Background()
		resp = account.StakingBalancesResponse{
			StakingBalances: []account.StakingBalance{
				{ValidatorAddress: "adr1", ChainName: "chain1", Amount: "42"},
				{ValidatorAddress: "adr2", ChainName: "chain1", Amount: "42"},
			},
		}
		respJSON, _ = json.Marshal(resp)
	)
	tests := []struct {
		name               string
		expectedStatusCode int
		expectedBody       string
		expectedError      string
		setup              func(mocks)
	}{
		{
			name:               "ok",
			expectedStatusCode: http.StatusOK,
			expectedBody:       string(respJSON),

			setup: func(m mocks) {
				m.app.EXPECT().StakingBalances(ctx, []string{"adr1"}).Return(resp.StakingBalances, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			assert := assert.New(t)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "address", Value: "adr1"}}
			c.Request, _ = http.NewRequestWithContext(ctx, http.MethodGet, "", nil)
			// FIXME remove when #801 is done
			fflag.EnableGlobal(account.FixSlashedDelegations)
			// add logger or else it fails
			logger := logging.New(logging.LoggingConfig{})
			c.Set(logging.LoggerKey, logger)
			ac := newAccountAPI(t, tt.setup)

			ac.GetDelegationsByAddress(c)

			assert.Equal(tt.expectedStatusCode, w.Code)
			if tt.expectedError != "" {
				require.Len(c.Errors, 1, "expected one error but got %d", len(c.Errors))
				require.EqualError(c.Errors[0], tt.expectedError)
				return
			}
			require.Empty(c.Errors)
			assert.JSONEq(tt.expectedBody, w.Body.String())
		})
	}

}
