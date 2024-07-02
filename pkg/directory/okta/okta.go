// Package okta contains the Okta directory provider.
package okta

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/okta/okta-sdk-golang/v2/okta/query"
)

// errors
var (
	ErrInvalidURL = errors.New("okta: invalid URL")
)

func listAllGroups(
	ctx context.Context,
	client *okta.Client,
	batchSize int,
) ([]*okta.Group, error) {
	groups, err := listAll(ctx, func(ctx context.Context) ([]*okta.Group, *okta.Response, error) {
		return client.Group.ListGroups(ctx, &query.Params{
			Limit: int64(batchSize),
		})
	})
	if err != nil {
		return nil, fmt.Errorf("error listing all groups: %w", err)
	}

	return groups, nil
}

func listChangedGroups(
	ctx context.Context,
	client *okta.Client,
	lastUpdated, lastMembershipUpdated time.Time,
	batchSize int,
) ([]*okta.Group, error) {
	groups, err := listAll(ctx, func(ctx context.Context) ([]*okta.Group, *okta.Response, error) {
		filter := fmt.Sprintf(`lastUpdated gt "%s" or lastMembershipUpdated gt "%s"`,
			lastUpdated.UTC().Format(filterDateFormat),
			lastMembershipUpdated.UTC().Format(filterDateFormat))

		return client.Group.ListGroups(ctx, &query.Params{
			Limit:  int64(batchSize),
			Filter: filter,
		})
	})
	if err != nil {
		return nil, fmt.Errorf("error listing changed groups: %w", err)
	}

	return groups, nil
}

func listGroupUsers(
	ctx context.Context,
	client *okta.Client,
	groupID string,
	batchSize int,
) ([]*okta.User, error) {
	users, err := listAll(ctx, func(ctx context.Context) ([]*okta.User, *okta.Response, error) {
		return client.Group.ListGroupUsers(ctx, groupID, &query.Params{
			Limit: int64(batchSize),
		})
	})
	if err != nil {
		return nil, fmt.Errorf("error listing group users: %w", err)
	}

	return users, nil
}

func listAll[T any](ctx context.Context, f func(ctx context.Context) ([]T, *okta.Response, error)) ([]T, error) {
	els, res, err := f(ctx)
	if err != nil {
		return nil, err
	}

	for res.HasNextPage() {
		var next []T
		res, err = res.Next(ctx, &next)
		if err != nil {
			return nil, err
		}
		els = append(els, next...)
	}

	return els, nil
}
