package cmd

import (
	"fmt"
	"html/template"
	"os"

	"github.com/dh1tw/remoteRotator/discovery"
	"github.com/spf13/cobra"
)

var enumerateCmd = &cobra.Command{
	Use:   "enumerate",
	Short: "discover and list all available rotators on the network",
	Long: `discover and list all available rotators on the network

This command performs a mDNS query on your local network and will report all
found rotators with their parameters.`,
	Run: enumerateMDNS,
}

func init() {
	RootCmd.AddCommand(enumerateCmd)
}

func enumerateMDNS(cmd *cobra.Command, args []string) {

	fmt.Println("\n...discovering rotators (please wait)")
	rots, err := discovery.LookupRotators()
	if err != nil {
		fmt.Println(err)
	}

	if err := tmpl.Execute(os.Stdout, rots); err != nil {
		fmt.Println(err)
	}
}

var tmpl = template.Must(template.New("").Parse(
	`
Found {{. | len}} rotator(s) on this network

{{range .}}Rotator:
   Name:         {{.Name}}
   URL:          {{.URL}}
   Host:         {{.Host}}{{if .AddrV4}}
   Address IPv4: {{.AddrV4}}{{else}}
   Address IPv6: {{.AddrV6}}{{end}}
   Port:         {{.Port}}

{{end}}
`,
))
