// Copyright 2023 Flant JSC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package app

import (
	"errors"
	"fmt"
	"net/url"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	MirrorModuleDirectory  = ""
	MirrorModuleSourcePath = ""
)

func DefineMirrorModulesFlags(cmd *kingpin.CmdClause) {
	cmd.Flag("modules-dir", "Path to modules directory.").
		Short('d').
		PlaceHolder("PATH").
		Required().
		Envar(configEnvName("MIRROR_MODULES_DIR")).
		StringVar(&MirrorModuleDirectory)
	cmd.Flag("module-source", "Path to ModuleSource YAML document describing where to pull modules from. Conflicts with --registry").
		Short('m').
		PlaceHolder("PATH").
		Envar(configEnvName("MIRROR_MODULES_SOURCE")).
		ExistingFileVar(&MirrorModuleSourcePath)
	cmd.Flag("registry", "Push modules to your private registry, specified as registry-host[:port][/path]. Conflicts with --module-source").
		Short('r').
		Envar(configEnvName("MIRROR_PRIVATE_REGISTRY")).
		StringVar(&MirrorRegistry)
	cmd.Flag("registry-login", "Username to log into your registry.").
		Short('u').
		PlaceHolder("LOGIN").
		Envar(configEnvName("MIRROR_USER")).
		StringVar(&MirrorRegistryUsername)
	cmd.Flag("registry-password", "Password to log into your registry.").
		Short('p').
		PlaceHolder("PASSWORD").
		Envar(configEnvName("MIRROR_PASS")).
		StringVar(&MirrorRegistryPassword)
	cmd.Flag("insecure", "Interact with registries over HTTP.").
		BoolVar(&MirrorInsecure)

	cmd.PreAction(func(c *kingpin.ParseContext) error {
		if err := validateRegistryCredentials(); err != nil {
			return err
		}

		if MirrorRegistry != "" {
			_, err := url.Parse("docker://" + MirrorRegistry)
			if err != nil {
				return fmt.Errorf("Malformed registry URL: %w", err)
			}
		}

		if MirrorModuleSourcePath == "" && MirrorRegistry == "" {
			return errors.New("One of --modules-source or --registry flags is required.")
		}

		if MirrorModuleSourcePath != "" && MirrorRegistry != "" {
			return errors.New("You have specified both --module-source and --registry flags. This is not how it works.\n\n" +
				"Leave only --module-source if you want to pull modules from ModuleSource.\n" +
				"Leave only --registry if you already pulled modules images and want to push it to your private registry.")
		}

		return nil
	})
}
