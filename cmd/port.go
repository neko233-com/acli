package cmd

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var portCmd = &cobra.Command{
	Use:   "port [port]",
	Short: "Check if a port is in use",
	Long:  `Check if a specific port is being used and show which process is using it.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		port, err := strconv.Atoi(args[0])
		if err != nil || port < 1 || port > 65535 {
			fmt.Println("Invalid port number (1-65535)")
			os.Exit(1)
		}

		addr := fmt.Sprintf(":%d", port)
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			fmt.Printf("Port %d is IN USE\n", port)
		} else {
			listener.Close()
			fmt.Printf("Port %d is FREE\n", port)
		}
	},
}