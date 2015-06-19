package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/cloudfoundry-incubator/diego-ssh/cf-plugin/cmd"
	"github.com/cloudfoundry-incubator/diego-ssh/cf-plugin/models/app"
	"github.com/cloudfoundry-incubator/diego-ssh/cf-plugin/models/credential"
	"github.com/cloudfoundry-incubator/diego-ssh/cf-plugin/models/info"
	"github.com/cloudfoundry-incubator/diego-ssh/cf-plugin/options"
	"github.com/cloudfoundry-incubator/diego-ssh/cf-plugin/terminal"
	"github.com/cloudfoundry/cli/plugin"
	"golang.org/x/crypto/ssh"
)

type SSHPlugin struct {
	OutputWriter io.Writer
}

func (p *SSHPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "Diego-SSH",
		Version: plugin.VersionType{
			Major: 0,
			Minor: 1,
			Build: 0,
		},
		Commands: []plugin.Command{
			{
				Name:     "ssh",
				HelpText: "ssh to an application container instance",
				UsageDetails: plugin.Usage{
					Usage: options.SSHUsage(),
				},
			},
			{
				Name:     "enable-ssh",
				HelpText: "enable ssh on an application container instance",
				UsageDetails: plugin.Usage{
					Usage: cmd.EnableSSHUsage,
				},
			},
			{
				Name:     "disable-ssh",
				HelpText: "disable ssh on an application container instance",
				UsageDetails: plugin.Usage{
					Usage: cmd.DisableSSHUsage,
				},
			},
			{
				Name:     "ssh-enabled",
				HelpText: "reports whether SSH is enabled on an application container instance",
				UsageDetails: plugin.Usage{
					Usage: cmd.SSHEnabledUsage,
				},
			},
			{
				Name:     "allow-space-ssh",
				HelpText: "allows applications to use SSH within a space",
				UsageDetails: plugin.Usage{
					Usage: "cf allow-space-ssh SPACE_NAME",
				},
			},
			{
				Name:     "disallow-space-ssh",
				HelpText: "reports whether SSH is enabled on an application container instance",
				UsageDetails: plugin.Usage{
					Usage: "cf disallow-space-ssh SPACE_NAME",
				},
			},
			{
				Name:     "space-ssh-allowed",
				HelpText: "reports whether SSH is allowed in a space",
				UsageDetails: plugin.Usage{
					Usage: "cf space-ssh-allowed SPACE_NAME",
				},
			},
		},
	}
}

func (p *SSHPlugin) Run(cli plugin.CliConnection, args []string) {
	p.OutputWriter = os.Stdout
	appFactory := app.NewAppFactory(cli)

	switch args[0] {
	case "enable-ssh":
		err := cmd.EnableSSH(args, appFactory, p.OutputWriter)
		if err != nil {
			p.Fail(err.Error())
			return
		}
	case "disable-ssh":
		err := cmd.EnableSSH(args, appFactory, p.OutputWriter)
		if err != nil {
			p.Fail(err.Error())
			return
		}
	case "ssh-enabled":
		err := cmd.SSHEnabled(args, appFactory, p.OutputWriter)
		if err != nil {
			p.Fail(err.Error())
			return
		}
	case "ssh":
		opts := options.NewSSHOptions()
		err := opts.Parse(args)
		if err != nil {
			p.Fail(err.Error())
			fmt.Fprintf(p.OutputWriter, options.SSHUsage())
			return
		}

		secureShell := cmd.NewSecureShell(
			cmd.DefaultSecureDialer(),
			terminal.DefaultHelper(),
			cmd.DefaultListenerFactory(),
			30*time.Second,
			appFactory,
			info.NewInfoFactory(cli),
			credential.NewCredentialFactory(cli),
		)

		err = secureShell.Connect(opts)
		if err != nil {
			p.Fail(err.Error())
			return
		}
		defer secureShell.Close()

		err = secureShell.LocalPortForward()
		if err != nil {
			return
		}

		if opts.SkipRemoteExecution {
			err = secureShell.Wait()
		} else {
			err = secureShell.InteractiveSession()
		}

		if err == nil {
			return
		}

		if exitError, ok := err.(*ssh.ExitError); ok {
			exitStatus := exitError.ExitStatus()
			if sig := exitError.Signal(); sig != "" {
				fmt.Printf("Process terminated by signal: %s. Exited with %d.\n", sig, exitStatus)
			}
			os.Exit(exitStatus)
		} else {
			p.Fail(err.Error())
		}

	default:
		p.Fail("Invalid command")
	}
}

func main() {
	sshPlugin := &SSHPlugin{}
	plugin.Start(sshPlugin)
}

func (p *SSHPlugin) Fail(message string) {
	fmt.Fprintf(p.OutputWriter, "FAILED\n\n%s\n", message)
}