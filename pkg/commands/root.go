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
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-logr/stdr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform/detect"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/updaters"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/wait"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests"
	schedmanifests "github.com/k8stopologyawareschedwg/deployer/pkg/manifests/sched"
	"github.com/k8stopologyawareschedwg/deployer/pkg/options"
)

type internalOptions struct {
	verbose                     int
	replicas                    int
	rteConfigFile               string
	schedScoringStratConfigFile string
	schedCacheParamsConfigFile  string
	plat                        string
}

func ShowHelp(cmd *cobra.Command, args []string) error {
	fmt.Fprint(cmd.OutOrStderr(), cmd.UsageString())
	return nil
}

type NewCommandFunc func(ev *deployer.Environment, ko *options.Options) *cobra.Command

// NewRootCommand returns entrypoint command to interact with all other commands
func NewRootCommand(env *deployer.Environment, extraCmds ...NewCommandFunc) *cobra.Command {
	internalOpts := internalOptions{}
	commonOpts := options.Options{}

	root := &cobra.Command{
		Use:   "deployer",
		Short: "deployer helps setting up all the topology-aware-scheduling components on a kubernetes cluster",

		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return PostSetupOptions(env, &commonOpts, &internalOpts)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return ShowHelp(cmd, args)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	InitFlags(root.PersistentFlags(), &commonOpts, &internalOpts)

	root.AddCommand(
		NewRenderCommand(env, &commonOpts),
		NewValidateCommand(env, &commonOpts),
		NewDeployCommand(env, &commonOpts),
		NewRemoveCommand(env, &commonOpts),
		NewSetupCommand(env, &commonOpts),
		NewDetectCommand(env, &commonOpts),
		NewImagesCommand(env, &commonOpts),
	)
	for _, extraCmd := range extraCmds {
		root.AddCommand(extraCmd(env, &commonOpts))
	}

	return root
}

func InitFlags(flags *pflag.FlagSet, commonOpts *options.Options, internalOpts *internalOptions) {
	flags.IntVarP(&internalOpts.verbose, "verbose", "v", 1, "set the tool verbosity.")
	flags.StringVarP(&internalOpts.plat, "platform", "P", "", "platform kind:version to deploy on (example kubernetes:v1.22)")
	flags.StringVar(&internalOpts.rteConfigFile, "rte-config-file", "", "inject rte configuration reading from this file.")
	flags.StringVar(&internalOpts.schedScoringStratConfigFile, "sched-scoring-strat-config-file", "", "inject scheduler scoring strategy configuration reading from this file.")
	flags.StringVar(&internalOpts.schedCacheParamsConfigFile, "sched-cache-params-config-file", "", "inject scheduler fine cache params configuration reading from this file.")
	flags.IntVarP(&internalOpts.replicas, "replicas", "R", 1, "set the replica value - where relevant.")

	flags.DurationVarP(&commonOpts.WaitInterval, "wait-interval", "E", 2*time.Second, "wait interval.")
	flags.DurationVarP(&commonOpts.WaitTimeout, "wait-timeout", "T", 2*time.Minute, "wait timeout.")
	flags.BoolVar(&commonOpts.PullIfNotPresent, "pull-if-not-present", false, "force pull policies to IfNotPresent.")
	flags.StringVar(&commonOpts.UpdaterType, "updater-type", "RTE", "type of updater to deploy - RTE or NFD")
	flags.BoolVar(&commonOpts.UpdaterPFPEnable, "updater-pfp-enable", true, "toggle PFP support on the updater side.")
	flags.BoolVar(&commonOpts.UpdaterNotifEnable, "updater-notif-enable", false, "toggle event-based notification support on the updater side.")
	flags.BoolVar(&commonOpts.UpdaterCRIHooksEnable, "updater-cri-hooks-enable", false, "toggle installation of CRI hooks on the updater side.")
	flags.BoolVar(&commonOpts.UpdaterCustomSELinuxPolicy, "updater-custom-selinux-policy", false, "toggle installation of selinux policy on the updater side. off by default")
	flags.DurationVar(&commonOpts.UpdaterSyncPeriod, "updater-sync-period", manifests.DefaultUpdaterSyncPeriod, "tune the updater synchronization (nrt update) interval. Use 0 to disable.")
	flags.IntVar(&commonOpts.UpdaterVerbose, "updater-verbose", manifests.DefaultUpdaterVerbose, "set the updater verbosiness.")
	flags.StringVar(&commonOpts.SchedProfileName, "sched-profile-name", schedmanifests.DefaultProfileName, "inject scheduler profile name.")
	flags.DurationVar(&commonOpts.SchedResyncPeriod, "sched-resync-period", schedmanifests.DefaultResyncPeriod, "inject scheduler resync period.")
	flags.IntVar(&commonOpts.SchedVerbose, "sched-verbose", schedmanifests.DefaultVerbose, "set the scheduler verbosiness.")
	flags.BoolVar(&commonOpts.SchedCtrlPlaneAffinity, "sched-ctrlplane-affinity", schedmanifests.DefaultCtrlPlaneAffinity, "toggle the scheduler control plane affinity.")
	flags.StringVar(&commonOpts.SchedLeaderElectResource, "sched-leader-elect-resource", schedmanifests.DefaultLeaderElectResource, "leader election resource namespaced name \"namespace/name\"")
}

func PostSetupOptions(env *deployer.Environment, commonOpts *options.Options, internalOpts *internalOptions) error {
	stdr.SetVerbosity(internalOpts.verbose) // MUST be the very first thing

	env.Log.V(3).Info("global polling settings", "interval", commonOpts.WaitInterval, "timeout", commonOpts.WaitTimeout)
	wait.SetBaseValues(commonOpts.WaitInterval, commonOpts.WaitTimeout)

	if internalOpts.replicas < 0 {
		err := env.EnsureClient()
		if err != nil {
			return err
		}

		env.Log.V(4).Info("autodetecting replicas from control plane")
		info, err := detect.ControlPlaneFromLister(env.Ctx, env.Cli)
		if err != nil {
			return err
		}
		commonOpts.Replicas = info.NodeCount
		env.Log.V(3).Info("autodetected control plane nodes, set replicas accordingly", "controlPlaneNodes", info.NodeCount)
	} else {
		commonOpts.Replicas = internalOpts.replicas
	}

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
	if internalOpts.schedScoringStratConfigFile != "" {
		data, err := os.ReadFile(internalOpts.schedScoringStratConfigFile)
		if err != nil {
			return err
		}
		commonOpts.SchedScoringStratConfigData = string(data)
		env.Log.Info("Scheduler Scoring Strategy config: read", "bytes", len(commonOpts.SchedScoringStratConfigData))
	}
	if internalOpts.schedCacheParamsConfigFile != "" {
		data, err := os.ReadFile(internalOpts.schedCacheParamsConfigFile)
		if err != nil {
			return err
		}
		commonOpts.SchedCacheParamsConfigData = string(data)
		env.Log.Info("Scheduler Cache Parameters config: read", "bytes", len(commonOpts.SchedCacheParamsConfigData))
	}

	return validateUpdaterType(commonOpts.UpdaterType)
}

func validateUpdaterType(updaterType string) error {
	if updaterType != updaters.RTE && updaterType != updaters.NFD {
		return fmt.Errorf("%q is invalid updater type", updaterType)
	}
	return nil
}
