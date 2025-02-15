// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2024, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file except in compliance with the License.

package up

import (
	"context"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"kraftkit.sh/cmdfactory"
	"kraftkit.sh/internal/cli/kraft/compose/create"
	"kraftkit.sh/internal/cli/kraft/compose/start"
	"kraftkit.sh/log"
	"kraftkit.sh/packmanager"
)

type UpOptions struct {
	composefile string
}

func NewCmd() *cobra.Command {
	cmd, err := cmdfactory.New(&UpOptions{}, cobra.Command{
		Short:   "Run a compose project",
		Use:     "up [FLAGS]",
		Args:    cobra.NoArgs,
		Aliases: []string{},
		Example: heredoc.Doc(`
			# Run a compose project
			$ kraft compose up
		`),
		Annotations: map[string]string{
			cmdfactory.AnnotationHelpGroup: "compose",
		},
	})
	if err != nil {
		panic(err)
	}

	return cmd
}

func (opts *UpOptions) Pre(cmd *cobra.Command, _ []string) error {
	ctx, err := packmanager.WithDefaultUmbrellaManagerInContext(cmd.Context())
	if err != nil {
		return err
	}

	cmd.SetContext(ctx)

	if cmd.Flag("file").Changed {
		opts.composefile = cmd.Flag("file").Value.String()
	}

	log.G(cmd.Context()).WithField("composefile", opts.composefile).Debug("using")
	return nil
}

func (opts *UpOptions) Run(ctx context.Context, _ []string) error {
	createOptions := create.CreateOptions{
		Composefile: opts.composefile,
	}

	if err := createOptions.Run(ctx, []string{}); err != nil {
		return err
	}

	startOptions := start.StartOptions{
		Composefile: opts.composefile,
	}

	return startOptions.Run(ctx, []string{})
}
