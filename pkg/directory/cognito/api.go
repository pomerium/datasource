package cognito

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/google/uuid"

	"github.com/pomerium/datasource/pkg/directory"
)

func listGroups(
	ctx context.Context,
	client cognitoidentityprovider.ListGroupsAPIClient,
	userPoolID string,
) ([]directory.Group, error) {
	var groups []directory.Group

	paginator := cognitoidentityprovider.NewListGroupsPaginator(client, &cognitoidentityprovider.ListGroupsInput{
		UserPoolId: aws.String(userPoolID),
	})
	for paginator.HasMorePages() {
		res, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, g := range res.Groups {
			groups = append(groups, directory.Group{
				ID:   getGroupID(g),
				Name: getGroupName(g),
			})
		}
	}

	return groups, nil
}

func listUserPoolIDs(
	ctx context.Context,
	client cognitoidentityprovider.ListUserPoolsAPIClient,
) ([]string, error) {
	var userPoolIDs []string

	paginator := cognitoidentityprovider.NewListUserPoolsPaginator(client, &cognitoidentityprovider.ListUserPoolsInput{
		MaxResults: aws.Int32(60),
	})
	for paginator.HasMorePages() {
		res, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, p := range res.UserPools {
			userPoolIDs = append(userPoolIDs, *p.Id)
		}
	}

	return userPoolIDs, nil
}

func listUsers(
	ctx context.Context,
	client cognitoidentityprovider.ListUsersAPIClient,
	userPoolID string,
) ([]directory.User, error) {
	var users []directory.User

	paginator := cognitoidentityprovider.NewListUsersPaginator(client, &cognitoidentityprovider.ListUsersInput{
		UserPoolId: aws.String(userPoolID),
	})
	for paginator.HasMorePages() {
		res, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, u := range res.Users {
			users = append(users, directory.User{
				ID:          getUserID(u),
				DisplayName: getUserDisplayName(u),
				Email:       getUserEmail(u),
			})
		}
	}

	return users, nil
}

func listUserIDsInGroup(
	ctx context.Context,
	client cognitoidentityprovider.ListUsersInGroupAPIClient,
	userPoolID string,
	groupID string,
) ([]string, error) {
	var userIDs []string

	paginator := cognitoidentityprovider.NewListUsersInGroupPaginator(client, &cognitoidentityprovider.ListUsersInGroupInput{
		UserPoolId: aws.String(userPoolID),
		GroupName:  aws.String(groupID),
	})
	for paginator.HasMorePages() {
		res, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, u := range res.Users {
			userIDs = append(userIDs, getUserID(u))
		}
	}

	return userIDs, nil
}

func getGroupID(group types.GroupType) string {
	if group.GroupName == nil {
		return uuid.NewString()
	}
	return *group.GroupName
}

func getGroupName(group types.GroupType) string {
	if group.GroupName == nil {
		return uuid.NewString()
	}
	return *group.GroupName
}

func getUserID(u types.UserType) string {
	if v, ok := getUserAttribute(u.Attributes, "sub"); ok {
		return v
	}

	// this really shouldn't happen
	if u.Username == nil {
		return uuid.NewString()
	}

	return *u.Username
}

func getUserDisplayName(u types.UserType) string {
	if v, ok := getUserAttribute(u.Attributes, "name"); ok {
		return v
	}

	return ""
}

func getUserEmail(u types.UserType) string {
	if v, ok := getUserAttribute(u.Attributes, "email"); ok {
		return v
	}

	return ""
}

func getUserAttribute(attributes []types.AttributeType, name string) (value string, ok bool) {
	for _, attr := range attributes {
		if attr.Name != nil && attr.Value != nil &&
			*attr.Name == name {
			return *attr.Value, true
		}
	}

	return "", false
}
