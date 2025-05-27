package cognito_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/pomerium/datasource/pkg/directory"
	"github.com/pomerium/datasource/pkg/directory/cognito"
)

type roundTripperFunc func(r *http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestCognito(t *testing.T) {
	t.Parallel()

	httpClient := &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			rec := httptest.NewRecorder()

			switch r.Header.Get("x-amz-target") {
			case "AWSCognitoIdentityProviderService.ListGroups":
				var obj cognitoidentityprovider.ListGroupsInput
				err := json.NewDecoder(r.Body).Decode(&obj)
				if err != nil {
					return nil, err
				}

				err = json.NewEncoder(rec).Encode(&cognitoidentityprovider.ListGroupsOutput{
					Groups: []types.GroupType{
						{GroupName: aws.String("GROUP1")},
						{GroupName: aws.String("GROUP2")},
					},
				})
				if err != nil {
					return nil, err
				}
			case "AWSCognitoIdentityProviderService.ListUserPools":
				var obj cognitoidentityprovider.ListUserPoolsInput
				err := json.NewDecoder(r.Body).Decode(&obj)
				if err != nil {
					return nil, err
				}

				err = json.NewEncoder(rec).Encode(&cognitoidentityprovider.ListUserPoolsOutput{
					UserPools: []types.UserPoolDescriptionType{
						{Id: aws.String("USER-POOL-1")},
						{Id: aws.String("USER-POOL-2")},
					},
				})
				if err != nil {
					return nil, err
				}
			case "AWSCognitoIdentityProviderService.ListUsers":
				var obj cognitoidentityprovider.ListUsersInput
				err := json.NewDecoder(r.Body).Decode(&obj)
				if err != nil {
					return nil, err
				}

				err = json.NewEncoder(rec).Encode(&cognitoidentityprovider.ListUsersOutput{
					Users: []types.UserType{
						{
							Username: aws.String("USERx1"),
							Attributes: []types.AttributeType{
								{Name: aws.String("sub"), Value: aws.String("USER1")},
								{Name: aws.String("name"), Value: aws.String("user-1")},
								{Name: aws.String("email"), Value: aws.String("user1@example.com")},
							},
						},
						{
							Username: aws.String("USER2"),
							Attributes: []types.AttributeType{
								{Name: aws.String("name"), Value: aws.String("user-2")},
								{Name: aws.String("email"), Value: aws.String("user2@example.com")},
							},
						},
						{
							Username: aws.String("USER3"),
							Attributes: []types.AttributeType{
								{Name: aws.String("name"), Value: aws.String("user-3")},
								{Name: aws.String("email"), Value: aws.String("user3@example.com")},
							},
						},
					},
				})
				if err != nil {
					return nil, err
				}
			case "AWSCognitoIdentityProviderService.ListUsersInGroup":
				var obj cognitoidentityprovider.ListUsersInGroupInput
				err := json.NewDecoder(r.Body).Decode(&obj)
				if err != nil {
					return nil, err
				}

				err = json.NewEncoder(rec).Encode(map[string]*cognitoidentityprovider.ListUsersInGroupOutput{
					"GROUP1": {
						Users: []types.UserType{
							{Username: aws.String("USER1")},
							{Username: aws.String("USER2")},
							{Username: aws.String("USER3")},
						},
					},
					"GROUP2": {
						Users: []types.UserType{
							{Username: aws.String("USER3")},
						},
					},
				}[*obj.GroupName])
				if err != nil {
					return nil, err
				}
			default:
				rec.WriteHeader(http.StatusNotFound)
			}

			return rec.Result(), nil
		}),
	}

	httptest.NewRecorder()
	c := cognito.New(
		cognito.WithAccessKeyID("ACCESS_KEY_ID"),
		cognito.WithHTTPClient(httpClient),
		cognito.WithLogger(zerolog.New(zerolog.NewTestWriter(t))),
		cognito.WithRegion("us-east-1"),
		cognito.WithSecretAccessKey("SECRET_ACCESS_KEY"),
		cognito.WithSessionToken("SESSION_TOKEN"))

	groups, users, err := c.GetDirectory(t.Context())
	assert.NoError(t, err)
	assert.Equal(t, []directory.Group{
		{ID: "GROUP1", Name: "GROUP1"},
		{ID: "GROUP2", Name: "GROUP2"},
	}, groups)
	assert.Equal(t, []directory.User{
		{
			ID:          "USER1",
			DisplayName: "user-1",
			Email:       "user1@example.com",
			GroupIDs:    []string{"GROUP1"},
		},
		{
			ID:          "USER2",
			DisplayName: "user-2",
			Email:       "user2@example.com",
			GroupIDs:    []string{"GROUP1"},
		},
		{
			ID:          "USER3",
			DisplayName: "user-3",
			Email:       "user3@example.com",
			GroupIDs:    []string{"GROUP1", "GROUP2"},
		},
	}, users)
}
