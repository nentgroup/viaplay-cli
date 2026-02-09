package gh

import (
	"context"

	"github.com/google/go-github/v74/github"
)

func (ghc *GitHubClient) GetUser(ctx context.Context, username string) (*github.User, error) {
	user, _, err := ghc.client.Users.Get(ctx, username)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserOrganizations returns the list of organizations the authenticated user belongs to
func (ghc *GitHubClient) GetUserOrganizations(ctx context.Context) ([]*github.Organization, error) {
	// Use ListOrgMemberships to get all organizations the user belongs to (even private ones)
	// with any role (member or admin)
	orgMemberships, _, err := ghc.client.Organizations.ListOrgMemberships(ctx, &github.ListOrgMembershipsOptions{
		State: "active",
		ListOptions: github.ListOptions{
			PerPage: 100, // Set a reasonable page size
		},
	})
	if err != nil {
		return nil, err
	}

	// Convert from membership to organization objects
	organizations := make([]*github.Organization, 0, len(orgMemberships))
	for _, membership := range orgMemberships {
		org := membership.GetOrganization()
		if org != nil {
			organizations = append(organizations, org)
		}
	}

	return organizations, nil
}
