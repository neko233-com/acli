package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

type incidentReport struct {
	Time   string        `json:"time"`
	Health healthResult  `json:"health"`
	Net    netInfoResult `json:"net"`
	Ports  []portEntry   `json:"ports,omitempty"`
	Top    topFrame      `json:"top"`
}

var incidentCmd = &cobra.Command{
	Use:   "incident",
	Short: "Collect one-shot incident diagnostics",
	Long:  `Collect health, network, ports, and top process diagnostics as JSON.`,
	Run: func(cmd *cobra.Command, args []string) {
		report, err := collectIncident()
		if err != nil {
			exitErr(err)
		}
		data, _ := json.MarshalIndent(report, "", "  ")
		fmt.Println(string(data))
	},
}

var alertCmd = &cobra.Command{
	Use:   "alert",
	Short: "Run threshold checks repeatedly",
	Long:  `Run check rules repeatedly and emit JSON events when checks fail.`,
	Run: func(cmd *cobra.Command, args []string) {
		interval, _ := cmd.Flags().GetDuration("interval")
		once, _ := cmd.Flags().GetBool("once")
		if interval <= 0 {
			interval = 30 * time.Second
		}
		for {
			results := runAlertChecks(cmd)
			for _, result := range results {
				if !result.OK {
					event := map[string]any{"time": time.Now().Format(time.RFC3339), "alert": result}
					data, _ := json.Marshal(event)
					fmt.Println(string(data))
				}
			}
			if once {
				for _, result := range results {
					if !result.OK {
						os.Exit(1)
					}
				}
				return
			}
			time.Sleep(interval)
		}
	},
}

func init() {
	alertCmd.Flags().StringSlice("port", nil, "TCP check host:port (repeatable)")
	alertCmd.Flags().StringSlice("dns", nil, "DNS check domain (repeatable)")
	alertCmd.Flags().StringSlice("http", nil, "HTTP check URL (repeatable)")
	alertCmd.Flags().Float64("disk-max", 0, "Fail if any disk used percent exceeds this value")
	alertCmd.Flags().Duration("interval", 30*time.Second, "Check interval")
	alertCmd.Flags().Bool("once", false, "Run once and exit non-zero on failure")
}

func collectIncident() (incidentReport, error) {
	netInfo, err := collectNetInfo("all")
	if err != nil {
		return incidentReport{}, err
	}
	ports, _ := collectPorts()
	top, err := collectTopFrame(20, "memory", "")
	if err != nil {
		return incidentReport{}, err
	}
	return incidentReport{
		Time:   time.Now().Format(time.RFC3339),
		Health: collectHealth(nil),
		Net:    netInfo,
		Ports:  ports,
		Top:    top,
	}, nil
}

func runAlertChecks(cmd *cobra.Command) []checkResult {
	ports, _ := cmd.Flags().GetStringSlice("port")
	domains, _ := cmd.Flags().GetStringSlice("dns")
	urls, _ := cmd.Flags().GetStringSlice("http")
	diskMax, _ := cmd.Flags().GetFloat64("disk-max")
	results := []checkResult{}
	for _, item := range ports {
		results = append(results, runPortCheck(item))
	}
	for _, domain := range domains {
		results = append(results, runDNSCheck(domain))
	}
	for _, url := range urls {
		results = append(results, runHTTPCheck(url))
	}
	if diskMax > 0 {
		results = append(results, runDiskCheck(diskMax)...)
	}
	return results
}
