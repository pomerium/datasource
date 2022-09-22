package auth0

//go:generate go run github.com/golang/mock/mockgen -destination=mock_auth0/mock.go . RoleManager,UserManager

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gopkg.in/auth0.v5/management"

	"github.com/pomerium/datasource/pkg/directory"
	"github.com/pomerium/datasource/pkg/directory/auth0/mock_auth0"
)

type mockNewRoleManagerFunc struct {
	CalledWithContext context.Context
	CalledWithConfig  *config

	ReturnRoleManager RoleManager
	ReturnUserManager UserManager
	ReturnError       error
}

func (m *mockNewRoleManagerFunc) f(
	ctx context.Context,
	cfg *config,
) (RoleManager, UserManager, error) {
	m.CalledWithContext = ctx
	m.CalledWithConfig = cfg

	return m.ReturnRoleManager, m.ReturnUserManager, m.ReturnError
}

type listOptionMatcher struct {
	expected management.RequestOption
}

func buildValues(opt management.RequestOption) map[string][]string {
	req, err := (&management.Management{}).NewRequest("GET", "example.com", nil, opt)
	if err != nil {
		panic(err)
	}
	return req.URL.Query()
}

func (lom listOptionMatcher) Matches(actual interface{}) bool {
	return gomock.Eq(buildValues(lom.expected)).Matches(buildValues(actual.(management.RequestOption)))
}

func (lom listOptionMatcher) String() string {
	return fmt.Sprintf("is equal to %v", buildValues(lom.expected))
}

func stringPtr(in string) *string {
	return &in
}

func TestProvider_GetDirectory(t *testing.T) {
	t.Parallel()

	expectedDomain := "login-example.auth0.com"
	expectedClientID := "c_id"
	expectedClientSecret := "secret"

	tests := []struct {
		name                         string
		setupRoleManagerExpectations func(*mock_auth0.MockRoleManager)
		newRoleManagerError          error
		expectedGroups               []directory.Group
		expectedUsers                []directory.User
		expectedError                error
	}{
		{
			name:                "errors if getting the role manager errors",
			newRoleManagerError: errors.New("new role manager error"),
			expectedError:       errors.New("auth0: could not get the role manager: new role manager error"),
		},
		{
			name: "errors if listing roles errors",
			setupRoleManagerExpectations: func(mrm *mock_auth0.MockRoleManager) {
				mrm.EXPECT().List(
					listOptionMatcher{expected: management.IncludeTotals(true)},
					listOptionMatcher{expected: management.Page(0)},
				).Return(nil, errors.New("list error"))
			},
			expectedError: errors.New("auth0: could not list roles: list error"),
		},
		{
			name: "errors if getting user ids errors",
			setupRoleManagerExpectations: func(mrm *mock_auth0.MockRoleManager) {
				mrm.EXPECT().List(
					listOptionMatcher{expected: management.IncludeTotals(true)},
					listOptionMatcher{expected: management.Page(0)},
				).Return(&management.RoleList{
					Roles: []*management.Role{
						{
							ID:   stringPtr("i-am-role-id"),
							Name: stringPtr("i-am-role-name"),
						},
					},
				}, nil)

				mrm.EXPECT().Users(
					"i-am-role-id",
					listOptionMatcher{expected: management.IncludeTotals(true)},
					listOptionMatcher{expected: management.Page(0)},
				).Return(nil, errors.New("users error"))
			},
			expectedError: errors.New("auth0: could not get users for role \"i-am-role-id\": users error"),
		},
		{
			name: "handles multiple pages of roles",
			setupRoleManagerExpectations: func(mrm *mock_auth0.MockRoleManager) {
				mrm.EXPECT().List(
					listOptionMatcher{expected: management.IncludeTotals(true)},
					listOptionMatcher{expected: management.Page(0)},
				).Return(&management.RoleList{
					List: management.List{
						Total: 3,
						Start: 0,
						Limit: 1,
					},
					Roles: []*management.Role{
						{
							ID:   stringPtr("i-am-role-id-1"),
							Name: stringPtr("i-am-role-name-1"),
						},
					},
				}, nil)

				mrm.EXPECT().List(
					listOptionMatcher{expected: management.IncludeTotals(true)},
					listOptionMatcher{expected: management.Page(1)},
				).Return(&management.RoleList{
					List: management.List{
						Total: 3,
						Start: 1,
						Limit: 1,
					},
					Roles: []*management.Role{
						{
							ID:   stringPtr("i-am-role-id-2"),
							Name: stringPtr("i-am-role-name-2"),
						},
					},
				}, nil)

				mrm.EXPECT().List(
					listOptionMatcher{expected: management.IncludeTotals(true)},
					listOptionMatcher{expected: management.Page(2)},
				).Return(&management.RoleList{
					List: management.List{
						Total: 3,
						Start: 2,
						Limit: 1,
					},
					Roles: []*management.Role{
						{
							ID:   stringPtr("i-am-role-id-3"),
							Name: stringPtr("i-am-role-name-3"),
						},
					},
				}, nil)

				mrm.EXPECT().Users(
					"i-am-role-id-1",
					listOptionMatcher{expected: management.IncludeTotals(true)},
					listOptionMatcher{expected: management.Page(0)},
				).Return(&management.UserList{}, nil)

				mrm.EXPECT().Users(
					"i-am-role-id-2",
					listOptionMatcher{expected: management.IncludeTotals(true)},
					listOptionMatcher{expected: management.Page(0)},
				).Return(&management.UserList{}, nil)

				mrm.EXPECT().Users(
					"i-am-role-id-3",
					listOptionMatcher{expected: management.IncludeTotals(true)},
					listOptionMatcher{expected: management.Page(0)},
				).Return(&management.UserList{}, nil)
			},
			expectedGroups: []directory.Group{
				{
					ID:   "i-am-role-id-1",
					Name: "i-am-role-name-1",
				},
				{
					ID:   "i-am-role-id-2",
					Name: "i-am-role-name-2",
				},
				{
					ID:   "i-am-role-id-3",
					Name: "i-am-role-name-3",
				},
			},
		},
		{
			name: "handles multiple pages of users",
			setupRoleManagerExpectations: func(mrm *mock_auth0.MockRoleManager) {
				mrm.EXPECT().List(
					listOptionMatcher{expected: management.IncludeTotals(true)},
					listOptionMatcher{expected: management.Page(0)},
				).Return(&management.RoleList{
					Roles: []*management.Role{
						{
							ID:   stringPtr("i-am-role-id-1"),
							Name: stringPtr("i-am-role-name-1"),
						},
					},
				}, nil)

				mrm.EXPECT().Users(
					"i-am-role-id-1",
					listOptionMatcher{expected: management.IncludeTotals(true)},
					listOptionMatcher{expected: management.Page(0)},
				).Return(&management.UserList{
					List: management.List{
						Total: 3,
						Start: 0,
						Limit: 1,
					},
					Users: []*management.User{
						{ID: stringPtr("i-am-user-id-1")},
					},
				}, nil)

				mrm.EXPECT().Users(
					"i-am-role-id-1",
					listOptionMatcher{expected: management.IncludeTotals(true)},
					listOptionMatcher{expected: management.Page(1)},
				).Return(&management.UserList{
					List: management.List{
						Total: 3,
						Start: 1,
						Limit: 1,
					},
					Users: []*management.User{
						{ID: stringPtr("i-am-user-id-2")},
					},
				}, nil)

				mrm.EXPECT().Users(
					"i-am-role-id-1",
					listOptionMatcher{expected: management.IncludeTotals(true)},
					listOptionMatcher{expected: management.Page(2)},
				).Return(&management.UserList{
					List: management.List{
						Total: 3,
						Start: 2,
						Limit: 1,
					},
					Users: []*management.User{
						{ID: stringPtr("i-am-user-id-3")},
					},
				}, nil)
			},
			expectedGroups: []directory.Group{
				{
					ID:   "i-am-role-id-1",
					Name: "i-am-role-name-1",
				},
			},
			expectedUsers: []directory.User{
				{
					ID:       "i-am-user-id-1",
					GroupIDs: []string{"i-am-role-id-1"},
				},
				{
					ID:       "i-am-user-id-2",
					GroupIDs: []string{"i-am-role-id-1"},
				},
				{
					ID:       "i-am-user-id-3",
					GroupIDs: []string{"i-am-role-id-1"},
				},
			},
		},
		{
			name: "correctly builds out groups and users",
			setupRoleManagerExpectations: func(mrm *mock_auth0.MockRoleManager) {
				mrm.EXPECT().List(
					listOptionMatcher{expected: management.IncludeTotals(true)},
					listOptionMatcher{expected: management.Page(0)},
				).Return(&management.RoleList{
					List: management.List{
						Total: 2,
						Start: 0,
						Limit: 1,
					},
					Roles: []*management.Role{
						{
							ID:   stringPtr("i-am-role-id-1"),
							Name: stringPtr("i-am-role-name-1"),
						},
					},
				}, nil)

				mrm.EXPECT().List(
					listOptionMatcher{expected: management.IncludeTotals(true)},
					listOptionMatcher{expected: management.Page(1)},
				).Return(&management.RoleList{
					List: management.List{
						Total: 2,
						Start: 1,
						Limit: 1,
					},
					Roles: []*management.Role{
						{
							ID:   stringPtr("i-am-role-id-2"),
							Name: stringPtr("i-am-role-name-2"),
						},
					},
				}, nil)

				mrm.EXPECT().Users(
					"i-am-role-id-1",
					listOptionMatcher{expected: management.IncludeTotals(true)},
					listOptionMatcher{expected: management.Page(0)},
				).Return(&management.UserList{
					Users: []*management.User{
						{ID: stringPtr("i-am-user-id-4")},
						{ID: stringPtr("i-am-user-id-3")},
						{ID: stringPtr("i-am-user-id-2")},
						{ID: stringPtr("i-am-user-id-1")},
					},
				}, nil)

				mrm.EXPECT().Users(
					"i-am-role-id-2",
					listOptionMatcher{expected: management.IncludeTotals(true)},
					listOptionMatcher{expected: management.Page(0)},
				).Return(&management.UserList{
					Users: []*management.User{
						{ID: stringPtr("i-am-user-id-1")},
						{ID: stringPtr("i-am-user-id-4")},
						{ID: stringPtr("i-am-user-id-5")},
					},
				}, nil)
			},
			expectedGroups: []directory.Group{
				{
					ID:   "i-am-role-id-1",
					Name: "i-am-role-name-1",
				},
				{
					ID:   "i-am-role-id-2",
					Name: "i-am-role-name-2",
				},
			},
			expectedUsers: []directory.User{
				{
					ID:       "i-am-user-id-1",
					GroupIDs: []string{"i-am-role-id-1", "i-am-role-id-2"},
				},
				{
					ID:       "i-am-user-id-2",
					GroupIDs: []string{"i-am-role-id-1"},
				},
				{
					ID:       "i-am-user-id-3",
					GroupIDs: []string{"i-am-role-id-1"},
				},
				{
					ID:       "i-am-user-id-4",
					GroupIDs: []string{"i-am-role-id-1", "i-am-role-id-2"},
				},
				{
					ID:       "i-am-user-id-5",
					GroupIDs: []string{"i-am-role-id-2"},
				},
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mRoleManager := mock_auth0.NewMockRoleManager(ctrl)

			mNewManagersFunc := mockNewRoleManagerFunc{
				ReturnRoleManager: mRoleManager,
				ReturnError:       tc.newRoleManagerError,
			}

			if tc.setupRoleManagerExpectations != nil {
				tc.setupRoleManagerExpectations(mRoleManager)
			}

			p := New(
				WithDomain(expectedDomain),
				WithClientID(expectedClientID),
				WithClientSecret(expectedClientSecret),
			)
			p.cfg.newManagers = mNewManagersFunc.f

			actualGroups, actualUsers, err := p.GetDirectory(context.Background())
			if tc.expectedError != nil {
				assert.EqualError(t, err, tc.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.expectedGroups, actualGroups)
			assert.Equal(t, tc.expectedUsers, actualUsers)

			assert.Equal(t, expectedDomain, mNewManagersFunc.CalledWithConfig.domain)
			assert.Equal(t, expectedClientID, mNewManagersFunc.CalledWithConfig.clientID)
			assert.Equal(t, expectedClientSecret, mNewManagersFunc.CalledWithConfig.clientSecret)
		})
	}
}
