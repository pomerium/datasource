package cognito

import (
	"context"
	"fmt"
	"sort"
	"sync"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"golang.org/x/exp/maps"

	"github.com/pomerium/datasource/pkg/directory"
)

// The Provider retrieves users and groups from cognito.
type Provider struct {
	cfg *config

	mu     sync.Mutex
	client *cognitoidentityprovider.Client
}

func New(options ...Option) *Provider {
	return &Provider{cfg: getConfig(options...)}
}

func (p *Provider) GetDirectory(ctx context.Context) ([]directory.Group, []directory.User, error) {
	client, err := p.getClient(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("cognito: error getting aws cognito client %w", err)
	}

	var userPoolIDs []string
	if p.cfg.userPoolID != "" {
		userPoolIDs = append(userPoolIDs, p.cfg.userPoolID)
	} else {
		userPoolIDs, err = listUserPoolIDs(ctx, client)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("cognito: error listing user pool ids: %w", err)
	}

	groupLookup := map[string]directory.Group{}
	userLookup := map[string]directory.User{}
	for _, userPoolID := range userPoolIDs {
		users, err := listUsers(ctx, client, userPoolID)
		if err != nil {
			return nil, nil, fmt.Errorf("cognito: error listing users in user pool: %w", err)
		}

		for _, u := range users {
			userLookup[u.ID] = u
		}

		groups, err := listGroups(ctx, client, userPoolID)
		if err != nil {
			return nil, nil, fmt.Errorf("cognito: error listing groups in user pool: %w", err)
		}

		for _, g := range groups {
			groupLookup[g.ID] = g
		}

		for groupID := range groupLookup {
			userIDs, err := listUserIDsInGroup(ctx, client, userPoolID, groupID)
			if err != nil {
				return nil, nil, fmt.Errorf("cognito: error listing user ids in group in user pool: %w", err)
			}
			for _, userID := range userIDs {
				if u, ok := userLookup[userID]; ok {
					u.GroupIDs = append(u.GroupIDs, groupID)
					userLookup[userID] = u
				}
			}
		}
	}

	groups := maps.Values(groupLookup)
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].ID < groups[j].ID
	})
	users := maps.Values(userLookup)
	sort.Slice(users, func(i, j int) bool {
		return users[i].ID < users[j].ID
	})
	return groups, users, nil
}

func (p *Provider) getClient(ctx context.Context) (*cognitoidentityprovider.Client, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.client != nil {
		return p.client, nil
	}

	var options []func(*awsconfig.LoadOptions) error
	if p.cfg.accessKeyID != "" || p.cfg.secretAccessKey != "" || p.cfg.sessionToken != "" {
		options = append(options, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(p.cfg.accessKeyID, p.cfg.secretAccessKey, p.cfg.sessionToken)))
	}
	options = append(options, awsconfig.WithHTTPClient(p.cfg.getHTTPClient()))

	cfg, err := awsconfig.LoadDefaultConfig(ctx, options...)
	if err != nil {
		return nil, err
	}

	if p.cfg.region != "" {
		cfg.Region = p.cfg.region
	}

	p.client = cognitoidentityprovider.NewFromConfig(cfg)
	return p.client, nil
}
