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

	"github.com/go-logr/stdr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deploy"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/updaters"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/wait"
)

// TODO: move elsewhere
const (
	DefaultSchedulerProfileName  = "topology-aware-scheduler"
	DefaultSchedulerResyncPeriod = 0 * time.Second
	DefaultUpdaterSyncPeriod     = 10 * time.Second
)

type internalOptions struct {
	rteConfigFile string
	plat          string
}

func ShowHelp(cmd *cobra.Command, args []string) error {
	fmt.Fprint(cmd.OutOrStderr(), cmd.UsageString())
	return nil
}

type NewCommandFunc func(ev *deployer.Environment, ko *deploy.Options) *cobra.Command

// NewRootCommand returns entrypoint command to interact with all other commands
func NewRootCommand(extraCmds ...NewCommandFunc) *cobra.Command {
	env := deployer.Environment{
		Ctx: context.Background(),
		Log: stdr.New(log.New(os.Stderr, "", log.LstdFlags)),
	}
	internalOpts := internalOptions{}
	commonOpts := deploy.Options{}

	root := &cobra.Command{
		Use:   "deployer",
		Short: "deployer helps setting up all the topology-aware-scheduling components on a kubernetes cluster",

		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return PostSetupOptions(&env, &commonOpts, &internalOpts)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return ShowHelp(cmd, args)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	InitFlags(root.PersistentFlags(), &commonOpts, &internalOpts)

	root.AddCommand(
		NewRenderCommand(&env, &commonOpts),
		NewValidateCommand(&env, &commonOpts),
		NewDeployCommand(&env, &commonOpts),
		NewRemoveCommand(&env, &commonOpts),
		NewSetupCommand(&env, &commonOpts),
		NewDetectCommand(&env, &commonOpts),
		NewImagesCommand(&env, &commonOpts),
	)
	for _, extraCmd := range extraCmds {
		root.AddCommand(extraCmd(&env, &commonOpts))
	}

	return root
}

func InitFlags(flags *pflag.FlagSet, commonOpts *deploy.Options, internalOpts *internalOptions) {
	flags.StringVarP(&internalOpts.plat, "platform", "P", "", "platform kind:version to deploy on (example kubernetes:v1.22)")
	flags.StringVar(&internalOpts.rteConfigFile, "rte-config-file", "", "inject rte configuration reading from this file.")

	flags.IntVarP(&commonOpts.Replicas, "replicas", "R", 1, "set the replica value - where relevant.")
	flags.DurationVarP(&commonOpts.WaitInterval, "wait-interval", "E", 2*time.Second, "wait interval.")
	flags.DurationVarP(&commonOpts.WaitTimeout, "wait-timeout", "T", 2*time.Minute, "wait timeout.")
	flags.BoolVar(&commonOpts.PullIfNotPresent, "pull-if-not-present", false, "force pull policies to IfNotPresent.")
	flags.StringVar(&commonOpts.UpdaterType, "updater-type", "RTE", "type of updater to deploy - RTE or NFD")
	flags.BoolVar(&commonOpts.UpdaterPFPEnable, "updater-pfp-enable", true, "toggle PFP support on the updater side.")
	flags.BoolVar(&commonOpts.UpdaterNotifEnable, "updater-notif-enable", true, "toggle event-based notification support on the updater side.")
	flags.BoolVar(&commonOpts.UpdaterCRIHooksEnable, "updater-cri-hooks-enable", true, "toggle installation of CRI hooks on the updater side.")
	flags.DurationVar(&commonOpts.UpdaterSyncPeriod, "updater-sync-period", DefaultUpdaterSyncPeriod, "tune the updater synchronization (nrt update) interval. Use 0 to disable.")
	flags.IntVar(&commonOpts.UpdaterVerbose, "updater-verbose", 1, "set the updater verbosiness.")
	flags.StringVar(&commonOpts.SchedProfileName, "sched-profile-name", DefaultSchedulerProfileName, "inject scheduler profile name.")
	flags.DurationVar(&commonOpts.SchedResyncPeriod, "sched-resync-period", DefaultSchedulerResyncPeriod, "inject scheduler resync period.")
	flags.IntVar(&commonOpts.SchedVerbose, "sched-verbose", 4, "set the scheduler verbosiness.")
	flags.BoolVar(&commonOpts.SchedCtrlPlaneAffinity, "sched-ctrlplane-affinity", true, "toggle the scheduler control plane affinity.")
}

func PostSetupOptions(env *deployer.Environment, commonOpts *deploy.Options, internalOpts *internalOptions) error {
	env.Log.V(3).Info("global polling interval=%v timeout=%v", commonOpts.WaitInterval, commonOpts.WaitTimeout)
	wait.SetBaseValues(commonOpts.WaitInterval, commonOpts.WaitTimeout)

	// if it is unknown, it's fine
	if internalOpts.plat == "" {
		commonOpts.UserPlatform = platform.Unknown
		commonOpts.UserPlatformVersion = platform.MissingVersion
	} else {
		fields := strings.FieldsFunc(internalOpts.plat, func(c rune) bool {
			return c == ':'
		})
		if len(fields) != 2 {
			return fmt.Errorf("unsupported platform spec: %q", internalOpts.plat)
		}
		commonOpts.UserPlatform, _ = platform.ParsePlatform(fields[0])
		commonOpts.UserPlatformVersion, _ = platform.ParseVersion(fields[1])
	}

	if internalOpts.rteConfigFile != "" {
		data, err := os.ReadFile(internalOpts.rteConfigFile)
		if err != nil {
			return err
		}
		commonOpts.RTEConfigData = string(data)
		env.Log.Info("RTE config: read", "bytes", len(commonOpts.RTEConfigData))
	}
	return validateUpdaterType(commonOpts.UpdaterType)
}

func validateUpdaterType(updaterType string) error {
	if updaterType != updaters.RTE && updaterType != updaters.NFD {
		return fmt.Errorf("%q is invalid updater type", updaterType)
	}
	return nil
}
