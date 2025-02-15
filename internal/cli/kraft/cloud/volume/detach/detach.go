// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file except in compliance with the License.

package detach

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
	"kraftkit.sh/iostreams"
)

type DetachOptions struct {
	Auth   *config.AuthConfig       `noattribute:"true"`
	Client kcvolumes.VolumesService `noattribute:"true"`

	metro string
	token string
}

func NewCmd() *cobra.Command {
	cmd, err := cmdfactory.New(&DetachOptions{}, cobra.Command{
		Short:   "Detach a volume from an instance",
		Use:     "detach [FLAGS] UUID",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"det"},
		Long:    "Detach a volume from an instance.",
		Example: heredoc.Doc(`
			# Detach a volume from an instance
			$ kraft cloud volume detach 77d0316a-fbbe-488d-8618-5bf7a612477a
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

func (opts *DetachOptions) Pre(cmd *cobra.Command, _ []string) error {
	err := utils.PopulateMetroToken(cmd, &opts.metro, &opts.token)
	if err != nil {
		return fmt.Errorf("could not populate metro and token: %w", err)
	}

	return nil
}

func (opts *DetachOptions) Run(ctx context.Context, args []string) error {
	var err error

	if opts.Auth == nil {
		opts.Auth, err = config.GetKraftCloudAuthConfig(ctx, opts.token)
		if err != nil {
			return fmt.Errorf("could not retrieve credentials: %w", err)
		}
	}

	if opts.Client == nil {
		opts.Client = kraftcloud.NewVolumesClient(
			kraftcloud.WithToken(config.GetKraftCloudTokenAuthConfig(*opts.Auth)),
		)
	}

	var detachResp *kcclient.ServiceResponse[kcvolumes.DetachResponseItem]
	if utils.IsUUID(args[0]) {
		detachResp, err = opts.Client.WithMetro(opts.metro).DetachByUUID(ctx, args[0])
	} else {
		detachResp, err = opts.Client.WithMetro(opts.metro).DetachByName(ctx, args[0])
	}
	if err != nil {
		return fmt.Errorf("could not detach volume: %w", err)
	}
	detach, err := detachResp.FirstOrErr()
	if err != nil {
		return fmt.Errorf("could not detach volume: %w", err)
	}

	_, err = fmt.Fprintln(iostreams.G(ctx).Out, detach.UUID)
	return err
}
