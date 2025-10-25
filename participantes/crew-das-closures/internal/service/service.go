package service

import (
	"fmt"

	"github.com/spf13/cobra"
)

func Service(cmd *cobra.Command, args []string) {
	fmt.Println("service called")
}
