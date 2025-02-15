// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file except in compliance with the License.

package remove

import (
	"context"
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	kraftcloud "sdk.kraft.cloud"
	kcclient "sdk.kraft.cloud/client"
	kcvolumes "sdk.kraft.cloud/volumes"

	"kraftkit.sh/cmdfactory"
	"kraftkit.sh/config"
	"kraftkit.sh/internal/cli/kraft/cloud/utils"
	"kraftkit.sh/log"
)

type RemoveOptions struct {
	metro string
	token string
}

// Remove a KraftCloud persistent volume.
func Remove(ctx context.Context, opts *RemoveOptions, args ...string) error {
	if opts == nil {
		opts = &RemoveOptions{}
	}

	return opts.Run(ctx, args)
}

func NewCmd() *cobra.Command {
	cmd, err := cmdfactory.New(&RemoveOptions{}, cobra.Command{
		Short:   "Permanently delete a persistent volume",
		Use:     "remove UUID [UUID [...]]",
		Args:    cobra.MinimumNArgs(1),
		Aliases: []string{"rm"},
		Long: heredoc.Doc(`
			Permanently delete a persistent volume.
		`),
		Example: heredoc.Doc(`
			# Delete three persistent volumes
			$ kraft cloud volume rm UUID1 UUID2 UUID3
		`),
		Annotations: map[string]string{
			cmdfactory.AnnotationHelpGroup: "kraftcloud-vol",
		},
	})
	if err != nil {
		panic(err)
	}

	return cmd
}

func (opts *RemoveOptions) Pre(cmd *cobra.Command, _ []string) error {
	err := utils.PopulateMetroToken(cmd, &opts.metro, &opts.token)
	if err != nil {
		return fmt.Errorf("could not populate metro and token: %w", err)
	}

	return nil
}

func (opts *RemoveOptions) Run(ctx context.Context, args []string) error {
	auth, err := config.GetKraftCloudAuthConfig(ctx, opts.token)
	if err != nil {
		return fmt.Errorf("could not retrieve credentials: %w", err)
	}

	client := kraftcloud.NewVolumesClient(
		kraftcloud.WithToken(config.GetKraftCloudTokenAuthConfig(*auth)),
	)

	log.G(ctx).Infof("Deleting %d volume(s)", len(args))

	allUUIDs := true
	allNames := true
	for _, arg := range args {
		if utils.IsUUID(arg) {
			allNames = false
		} else {
			allUUIDs = false
		}
		if !(allUUIDs || allNames) {
			break
		}
	}

	var delResp *kcclient.ServiceResponse[kcvolumes.DeleteResponseItem]

	switch {
	case allUUIDs:
		if delResp, err = client.WithMetro(opts.metro).DeleteByUUIDs(ctx, args...); err != nil {
			return fmt.Errorf("deleting %d volume(s): %w", len(args), err)
		}
	case allNames:
		if delResp, err = client.WithMetro(opts.metro).DeleteByNames(ctx, args...); err != nil {
			return fmt.Errorf("deleting %d volume(s): %w", len(args), err)
		}
	default:
		for _, arg := range args {
			log.G(ctx).Infof("Deleting volume %s", arg)

			if utils.IsUUID(arg) {
				delResp, err = client.WithMetro(opts.metro).DeleteByUUIDs(ctx, arg)
			} else {
				delResp, err = client.WithMetro(opts.metro).DeleteByNames(ctx, arg)
			}
			if err != nil {
				return fmt.Errorf("could not delete volume %s: %w", arg, err)
			}
		}
	}
	if _, err = delResp.AllOrErr(); err != nil {
		return fmt.Errorf("removing volume(s): %w", err)
	}

	return nil
}
