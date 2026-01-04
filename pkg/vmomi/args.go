package vmomi

import (
	"context"
	"errors"

	"github.com/9506hqwy/vmomi-exporter/pkg/flag"
)

func GetTarget(ctx context.Context) (url, user, password string, noVerifySSL bool, err error) {
	url, ok := ctx.Value(flag.TargetUrlKey{}).(string)
	if !ok {
		return "", "", "", false, errors.New("target_url not found in context")
	}

	user, ok = ctx.Value(flag.TargetUserKey{}).(string)
	if !ok {
		return "", "", "", false, errors.New("target_user not found in context")
	}

	password, ok = ctx.Value(flag.TargetPasswordKey{}).(string)
	if !ok {
		return "", "", "", false, errors.New("target_password not found in context")
	}

	noVerifySSL, ok = ctx.Value(flag.TargetNoVerifySSLKey{}).(bool)
	if !ok {
		return "", "", "", false, errors.New("target_no_verify_ssl not found in context")
	}

	return url, user, password, noVerifySSL, nil
}
