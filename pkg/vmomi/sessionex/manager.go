package sessionex

import (
	"context"
	"net/url"

	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/soap"
)

func Login(ctx context.Context, endpoint string, username string, password string, noVerifySSL bool) (*vim25.Client, error) {
	u, err := soap.ParseURL(endpoint)
	if err != nil {
		return nil, err
	}

	sc := soap.NewClient(u, noVerifySSL)
	vc, err := vim25.NewClient(ctx, sc)
	if err != nil {
		return nil, err
	}

	sm := session.NewManager(vc)

	cred := url.UserPassword(username, password)
	err = sm.Login(ctx, cred)
	if err != nil {
		return nil, err
	}

	return vc, nil
}

func Logout(ctx context.Context, c *vim25.Client) error {
	sm := session.NewManager(c)
	return sm.Logout(ctx)
}
