package cmd

func init() {
	// Network - IP and Interfaces
	rootCmd.AddCommand(ipCmd)
	rootCmd.AddCommand(ifaceCmd)
	rootCmd.AddCommand(publicipCmd)

	// Network - Port and Connection
	rootCmd.AddCommand(portCmd)
	rootCmd.AddCommand(listenCmd)
	rootCmd.AddCommand(connCmd)
	rootCmd.AddCommand(netstatCmd)
	rootCmd.AddCommand(connectCmd)

	// Network - DNS and Lookup
	rootCmd.AddCommand(dnsCmd)
	rootCmd.AddCommand(lookupCmd)
	rootCmd.AddCommand(whoisCmd)

	// Network - Diagnostic
	rootCmd.AddCommand(pingCmd)
	rootCmd.AddCommand(tracerouteCmd)
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(portScanCmd)
	rootCmd.AddCommand(latencyTestCmd)

	// Network - Client
	rootCmd.AddCommand(tcpClientCmd)
	rootCmd.AddCommand(udpClientCmd)
	rootCmd.AddCommand(netcatCmd)
	rootCmd.AddCommand(socketTestCmd)

	// Network - HTTP
	rootCmd.AddCommand(httpCmd)
	rootCmd.AddCommand(httpDebugCmd)
	rootCmd.AddCommand(httpPostCmd)
	rootCmd.AddCommand(httpPutCmd)
	rootCmd.AddCommand(httpPatchCmd)
	rootCmd.AddCommand(httpDeleteCmd)
	rootCmd.AddCommand(http2TestCmd)
	rootCmd.AddCommand(http3TestCmd)
	rootCmd.AddCommand(httpHeadersCmd)
	rootCmd.AddCommand(curlCmd)
	rootCmd.AddCommand(websocketTestCmd)

	// Network - SSL and Security
	rootCmd.AddCommand(httpsCheckCmd)
	rootCmd.AddCommand(firewallCmd)
	rootCmd.AddCommand(crtLookupCmd)

	// Network - Traffic
	rootCmd.AddCommand(nettrafficCmd)
	rootCmd.AddCommand(bandwidthCmd)

	// Protocol Testing
	rootCmd.AddCommand(grpcTestCmd)

	// Process
	rootCmd.AddCommand(psCmd)
	rootCmd.AddCommand(pstreeCmd)
	rootCmd.AddCommand(pssearchCmd)
	rootCmd.AddCommand(psportsCmd)
	rootCmd.AddCommand(killCmd)
	rootCmd.AddCommand(monitorCmd)

	// System
	rootCmd.AddCommand(sysinfoCmd)
	rootCmd.AddCommand(cpuCmd)
	rootCmd.AddCommand(memCmd)
	rootCmd.AddCommand(diskCmd)
	rootCmd.AddCommand(uptimeCmd)
	rootCmd.AddCommand(whoamiCmd)
	rootCmd.AddCommand(hostnameCmd)

	// System - Services
	rootCmd.AddCommand(serviceListCmd)
	rootCmd.AddCommand(serviceStatusCmd)
	rootCmd.AddCommand(serviceStartCmd)
	rootCmd.AddCommand(serviceStopCmd)
	rootCmd.AddCommand(serviceRestartCmd)

	// System - Environment
	rootCmd.AddCommand(envListCmd)
	rootCmd.AddCommand(envGetCmd)

	// Utilities - Encoding
	rootCmd.AddCommand(base64EncodeCmd)
	rootCmd.AddCommand(base64DecodeCmd)
	rootCmd.AddCommand(urlEncodeCmd)
	rootCmd.AddCommand(urlDecodeCmd)
	rootCmd.AddCommand(hashMd5Cmd)
	rootCmd.AddCommand(hashSha256Cmd)
	rootCmd.AddCommand(hexEncodeCmd)
	rootCmd.AddCommand(hexDecodeCmd)

	// Utilities - Hashing
	rootCmd.AddCommand(hashFileCmd)

	// Utilities - Identity
	rootCmd.AddCommand(uuidGenCmd)
	rootCmd.AddCommand(uuidParseCmd)
	rootCmd.AddCommand(passwordGenCmd)
	rootCmd.AddCommand(passwordStrengthCmd)

	// Utilities - QR Code
	rootCmd.AddCommand(qrGenerateCmd)

	// Utilities - Time
	rootCmd.AddCommand(timeNowCmd)
	rootCmd.AddCommand(timeConvertCmd)
	rootCmd.AddCommand(cronNextCmd)

	// Utilities - Random
	rootCmd.AddCommand(randomStringCmd)

	// Utilities - Data
	rootCmd.AddCommand(jsonPrettyCmd)
	rootCmd.AddCommand(jsonValidateCmd)
	rootCmd.AddCommand(hexDumpCmd)
	rootCmd.AddCommand(fileInfoCmd)

	// Utilities - Report
	rootCmd.AddCommand(htmlReportCmd)

	// Utilities - Misc
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(agentCmd)
	rootCmd.AddCommand(docCmd)
}
