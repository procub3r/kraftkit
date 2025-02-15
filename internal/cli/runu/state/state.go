// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file except in compliance with the License.
package state

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/opencontainers/runc/libcontainer/utils"
	"github.com/spf13/cobra"

	"kraftkit.sh/cmdfactory"
	libcontainer "kraftkit.sh/libmocktainer"
	"kraftkit.sh/log"
)

const (
	flagRoot = "root"
)

// StateOptions implements the OCI "state" command.
type StateOptions struct {
	rootDir string
}

func NewCmd() *cobra.Command {
	cmd, err := cmdfactory.New(&StateOptions{}, cobra.Command{
		Short: "Output the state of a unikernel",
		Args:  cobra.ExactArgs(1),
		Use:   "state <unikernel-id>",
		Long:  "The state command outputs current state information for a unikernel.",
		Example: heredoc.Doc(`
			# Output the state of a unikernel
			$ runu state my-unikernel
		`),
	})
	if err != nil {
		panic(err)
	}

	return cmd
}

func (opts *StateOptions) Pre(cmd *cobra.Command, args []string) error {
	opts.rootDir = cmd.Flag(flagRoot).Value.String()
	if opts.rootDir == "" {
		return fmt.Errorf("state directory (--%s flag) is not set", flagRoot)
	}

	return nil
}

func (opts *StateOptions) Run(ctx context.Context, args []string) (retErr error) {
	defer func() {
		// Make sure the error is written to the configured log destination, so
		// that the message gets propagated through the caller (e.g. containerd-shim)
		if retErr != nil {
			log.G(ctx).Error(retErr)
		}
	}()

	cID := args[0]

	c, err := libcontainer.Load(opts.rootDir, cID)
	if err != nil {
		return fmt.Errorf("loading container from saved state: %w", err)
	}

	status, err := c.Status()
	if err != nil {
		return fmt.Errorf("getting container status: %w", err)
	}

	state, err := c.State()
	if err != nil {
		return fmt.Errorf("getting container state: %w", err)
	}

	pid := state.BaseState.InitProcessPid
	if status == libcontainer.Stopped {
		pid = 0
	}

	bundle, annotations := utils.Annotations(state.Config.Labels)

	cs := containerState{
		OCIVersion:  state.BaseState.Config.Version,
		ID:          state.BaseState.ID,
		Status:      status.String(),
		Bundle:      bundle,
		Pid:         pid,
		Annotations: annotations,
		RootFS:      state.BaseState.Config.Rootfs,
		Created:     state.BaseState.Created,
	}

	data, err := json.MarshalIndent(cs, "", "  ")
	if err != nil {
		return fmt.Errorf("serializing container state: %w", err)
	}

	_, _ = os.Stdout.Write(data)

	return nil
}

// containerState is a JSON-serializable representation of the OCI runtime state.
// https://github.com/opencontainers/runtime-spec/blob/v1.1.0/schema/state-schema.json
// https://github.com/opencontainers/runtime-spec/blob/v1.1.0/runtime.md#state
type containerState struct {
	// The version of OCI Runtime Specification that the document complies with
	OCIVersion string `json:"ociVersion"`
	// Container unique ID
	ID string `json:"id"`
	// Runtime state of the container
	Status string `json:"status"`
	// Absolute path to the container bundle directory
	Bundle string `json:"bundle"`
	// ID of the container's init process
	Pid int `json:"pid"`
	// User defined annotations associated with the container
	Annotations map[string]string `json:"annotations,omitempty"`

	/* Additional attributes, for runc compatibility */

	// Absolute path to the directory containing the container's root filesystem
	RootFS string `json:"rootfs"`
	// Creation time of the container in UTC
	Created time.Time `json:"created"`
}
