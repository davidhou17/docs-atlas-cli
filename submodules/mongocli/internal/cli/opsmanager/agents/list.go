// Copyright 2020 MongoDB Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package agents

import (
	"context"
	"fmt"
	"strings"

	"github.com/mongodb/mongocli/internal/cli"
	"github.com/mongodb/mongocli/internal/cli/require"
	"github.com/mongodb/mongocli/internal/config"
	"github.com/mongodb/mongocli/internal/flag"
	"github.com/mongodb/mongocli/internal/store"
	"github.com/mongodb/mongocli/internal/usage"
	"github.com/spf13/cobra"
)

type ListOpts struct {
	cli.GlobalOpts
	cli.OutputOpts
	cli.ListOpts
	agentType string
	store     store.AgentLister
}

func (opts *ListOpts) initStore(ctx context.Context) func() error {
	return func() error {
		var err error
		opts.store, err = store.New(store.AuthenticatedPreset(config.Default()), store.WithContext(ctx))
		return err
	}
}

var listTemplate = `HOSTNAME	TYPE	STATE{{range .Results}}
{{.Hostname}}	{{.TypeName}}	{{.StateName}}{{end}}
`

func (opts *ListOpts) Run() error {
	r, err := opts.store.Agents(opts.ConfigProjectID(), opts.agentType, opts.NewListOptions())
	if err != nil {
		return err
	}

	return opts.Print(r)
}

// mongocli om agent(s) list [--projectId projectId].
func ListBuilder() *cobra.Command {
	validArgs := []string{"AUTOMATION", "MONITORING", "BACKUP"}
	opts := &ListOpts{}
	cmd := &cobra.Command{
		Use:       fmt.Sprintf("list <%s>", strings.Join(validArgs, "|")),
		Aliases:   []string{"ls"},
		Args:      require.ExactValidArgs(1),
		ValidArgs: validArgs,
		Short:     "List available MongoDB Agents for your project.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.PreRunE(
				opts.ValidateProjectID,
				opts.initStore(cmd.Context()),
				opts.InitOutput(cmd.OutOrStdout(), listTemplate),
			)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.agentType = strings.ToUpper(args[0])
			return opts.Run()
		},
	}

	cmd.Flags().IntVar(&opts.PageNum, flag.Page, cli.DefaultPage, usage.Page)
	cmd.Flags().IntVar(&opts.ItemsPerPage, flag.Limit, cli.DefaultPageLimit, usage.Limit)

	cmd.Flags().StringVar(&opts.ProjectID, flag.ProjectID, "", usage.ProjectID)
	cmd.Flags().StringVarP(&opts.Output, flag.Output, flag.OutputShort, "", usage.FormatOut)

	return cmd
}
