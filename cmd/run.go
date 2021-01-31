// Copyright © 2020 Christian Fritz <mail@chr-fritz.de>
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

package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/chr-fritz/knx-exporter/pkg/knx"
	"github.com/chr-fritz/knx-exporter/pkg/metrics"
)

const RunPortParm = "exporter.port"
const RunConfigFileParm = "exporter.configFile"

type RunOptions struct {
	port       uint16
	configFile string
}

func NewRunOptions() *RunOptions {
	return &RunOptions{
		port: 8080,
	}
}

func NewRunCommand() *cobra.Command {
	runOptions := NewRunOptions()

	cmd := cobra.Command{
		Use:   "run",
		Short: "Run the exporter",
		Long:  `Run the exporter which exports the received values from all configured Group Addresses to prometheus.`,
		RunE:  runOptions.run,
	}
	cmd.Flags().Uint16VarP(&runOptions.port, "port", "p", 8080, "The port where all metrics should be exported.")
	cmd.Flags().StringVarP(&runOptions.configFile, "configFile", "f", "config.yaml", "The knx configuration file.")
	_ = viper.BindPFlag(RunPortParm, cmd.Flags().Lookup("port"))
	_ = viper.BindPFlag(RunConfigFileParm, cmd.Flags().Lookup("configFile"))
	return &cmd
}

func (i *RunOptions) run(cmd *cobra.Command, args []string) error {
	exporter := metrics.NewExporter(i.port)
	metricsExporter, err := knx.NewMetricsExporter(i.configFile)

	if err != nil {
		return err
	}

	collectors := metricsExporter.RegisterMetrics()
	exporter.MustRegister(collectors...)
	poller := knx.NewPoller(metricsExporter)

	go func() {
		if err := metricsExporter.Run(); err != nil {
			logrus.Warn(err)
		}
	}()
	poller.Run()

	defer func() {
		poller.Stop()
		metricsExporter.Close()
	}()
	return exporter.Run()
}

func init() {
	rootCmd.AddCommand(NewRunCommand())
}
