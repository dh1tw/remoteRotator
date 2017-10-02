package cmd

import (
	"fmt"
	"html/template"
	"os"

	"github.com/dh1tw/remoteRotator/discovery"
	"github.com/spf13/cobra"
)

var discoverCmd = &cobra.Command{
	Use:   "enumerate",
	Short: "discover and list all available rotators on the network",
	Long: `discover and list all available rotators on the network

The commend performs a mDNS query and will report all found rotators
with their network parameters and description.`,
	Run: discoverMDNS,
}

func init() {
	RootCmd.AddCommand(discoverCmd)
}

func discoverMDNS(cmd *cobra.Command, args []string) {

	fmt.Println("\n...discovering rotators (please wait)")
	rots := discovery.LookupRotators()

	if err := tmpl.Execute(os.Stdout, rots); err != nil {
		fmt.Println(err)
	}
}

var tmpl = template.Must(template.New("").Parse(
	`
Found {{. | len}} rotator(s) on this network

{{range .}}Rotator:
   Name:         {{.Name}}
   Description:  {{.Description}}
   URL:          {{.URL}}
   Host:         {{.Host}}{{if .AddrV6}}
   Address IPv6: {{.AddrV6}}{{else}}
   Address IPv4: {{.AddrV4}}{{end}}
   Port:         {{.Port}}
   
{{end}}
`,
))
