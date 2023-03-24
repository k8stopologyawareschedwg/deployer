/*
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2021 Red Hat, Inc.
 */

package commands

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/k8stopologyawareschedwg/deployer/pkg/clientutil"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/updaters"
)

// TODO: move elsewhere
const (
	DefaultSchedulerProfileName  = "topology-aware-scheduler"
	DefaultSchedulerResyncPeriod = 0 * time.Second
)

type CommonOptions struct {
	Debug               bool
	UserPlatform        platform.Platform
	UserPlatformVersion platform.Version
	Log                 logr.Logger
	DebugLog            logr.Logger
	Replicas            int
	RTEConfigData       string
	PullIfNotPresent    bool
	UpdaterType         string
	UpdaterPFPEnable    bool
	OCIHookNotifier     bool
	OCIHookListing      bool
	rteConfigFile       string
	plat                string
	platVer             string
	schedProfileName    string
	schedResyncPeriod   time.Duration
}

func ShowHelp(cmd *cobra.Command, args []string) error {
	fmt.Fprint(cmd.OutOrStderr(), cmd.UsageString())
	return nil
}

type NewCommandFunc func(ko *CommonOptions) *cobra.Command

// NewRootCommand returns entrypoint command to interact with all other commands
func NewRootCommand(extraCmds ...NewCommandFunc) *cobra.Command {
	commonOpts := &CommonOptions{}

	root := &cobra.Command{
		Use:   "deployer",
		Short: "deployer helps setting up all the topology-aware-scheduling components on a kubernetes cluster",

		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return PostSetupOptions(commonOpts)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return ShowHelp(cmd, args)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	InitFlags(root.PersistentFlags(), commonOpts)

	root.AddCommand(
		NewRenderCommand(commonOpts),
		NewValidateCommand(commonOpts),
		NewDeployCommand(commonOpts),
		NewRemoveCommand(commonOpts),
		NewSetupCommand(commonOpts),
		NewDetectCommand(commonOpts),
		NewImagesCommand(commonOpts),
	)
	for _, extraCmd := range extraCmds {
		root.AddCommand(extraCmd(commonOpts))
	}

	return root
}

func InitFlags(flags *pflag.FlagSet, commonOpts *CommonOptions) {
	flags.BoolVarP(&commonOpts.Debug, "debug", "D", false, "enable debug log")
	flags.StringVarP(&commonOpts.plat, "platform", "P", "", "platform kind:version to deploy on (example kubernetes:v1.22)")
	flags.IntVarP(&commonOpts.Replicas, "replicas", "R", 1, "set the replica value - where relevant.")
	flags.BoolVar(&commonOpts.PullIfNotPresent, "pull-if-not-present", false, "force pull policies to IfNotPresent.")
	flags.StringVar(&commonOpts.rteConfigFile, "rte-config-file", "", "inject rte configuration reading from this file.")
	flags.StringVar(&commonOpts.UpdaterType, "updater-type", "RTE", "type of updater to deploy - RTE or NFD")
	flags.BoolVar(&commonOpts.UpdaterPFPEnable, "updater-pfp-enable", true, "toggle PFP support on the updater side.")
	flags.StringVar(&commonOpts.schedProfileName, "sched-profile-name", DefaultSchedulerProfileName, "inject scheduler profile name.")
	flags.DurationVar(&commonOpts.schedResyncPeriod, "sched-resync-period", DefaultSchedulerResyncPeriod, "inject scheduler resync period.")
	flags.BoolVar(&commonOpts.OCIHookNotifier, "oci-hook-notifier", true, "toggle support for the notifier OCI hook.")
	flags.BoolVar(&commonOpts.OCIHookListing, "oci-hook-listing", false, "toggle support for the listing OCI hook.")
}

func PostSetupOptions(commonOpts *CommonOptions) error {
	// we abuse the logger to have a common interface and the timestamps
	commonOpts.Log = stdr.New(log.New(os.Stderr, "", log.LstdFlags))
	if commonOpts.Debug {
		commonOpts.DebugLog = commonOpts.Log.WithName("DEBUG")
	} else {
		commonOpts.DebugLog = logr.Discard()
	}

	// if it is unknown, it's fine
	if commonOpts.plat == "" {
		commonOpts.UserPlatform = platform.Unknown
		commonOpts.UserPlatformVersion = platform.MissingVersion
	} else {
		fields := strings.FieldsFunc(commonOpts.plat, func(c rune) bool {
			return c == ':'
		})
		if len(fields) != 2 {
			return fmt.Errorf("unsupported platform spec: %q", commonOpts.plat)
		}
		commonOpts.UserPlatform, _ = platform.ParsePlatform(fields[0])
		commonOpts.UserPlatformVersion, _ = platform.ParseVersion(fields[1])
	}

	if commonOpts.rteConfigFile != "" {
		data, err := os.ReadFile(commonOpts.rteConfigFile)
		if err != nil {
			return err
		}
		commonOpts.RTEConfigData = string(data)
		commonOpts.DebugLog.Info("RTE config: read", "bytes", len(commonOpts.RTEConfigData))
	}
	return validateUpdaterType(commonOpts.UpdaterType)
}

func validateUpdaterType(updaterType string) error {
	if updaterType != updaters.RTE && updaterType != updaters.NFD {
		return fmt.Errorf("%q is invalid updater type", updaterType)
	}
	return nil
}

func environFromOpts(commonOpts *CommonOptions) (*deployer.Environment, error) {
	cli, err := clientutil.New()
	if err != nil {
		return nil, err
	}
	return &deployer.Environment{
		Ctx: context.Background(),
		Cli: cli,
		Log: commonOpts.Log,
	}, nil
}
