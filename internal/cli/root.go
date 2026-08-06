package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/bitlab-dev/netnotify/internal/app"
	"github.com/bitlab-dev/netnotify/internal/config"
	"github.com/bitlab-dev/netnotify/internal/domain"
	"github.com/bitlab-dev/netnotify/internal/provider/gowa"
)

var Version = "0.1.0"

func Execute() error {
	cfgFile := ""
	fs := flag.NewFlagSet("netnotify", flag.ExitOnError)
	fs.StringVar(&cfgFile, "config", "", "config file")
	fs.StringVar(&cfgFile, "c", "", "config file")
	_ = fs.Parse(os.Args[1:])
	args := fs.Args()
	if len(args) == 0 {
		c, err := config.Load(cfgFile)
		if err != nil {
			return err
		}
		return app.Run(c)
	}
	switch args[0] {
	case "version":
		fmt.Println(Version)
		return nil
	case "validate":
		_, err := config.Load(cfgFile)
		if err == nil {
			fmt.Println("configuration valid")
		}
		return err
	case "providers":
		fmt.Println("gowa")
		return nil
	case "sources":
		fmt.Println("netdata")
		return nil
	case "health":
		c, err := config.Load(cfgFile)
		if err != nil {
			return err
		}
		if c.Providers.GoWA.Enabled {
			return gowa.New(c.Providers.GoWA).Health(context.Background())
		}
		return nil
	case "send":
		c, err := config.Load(cfgFile)
		if err != nil {
			return err
		}
		n := domain.NewNotification()
		n.Provider = "gowa"
		n.Severity = domain.SeverityInfo
		n.Title = "netnotify test"
		n.Text = "netnotify test message"
		if os.Getenv("NETNOTIFY_DRY_RUN") != "" {
			b, _ := json.Marshal(n)
			fmt.Println(string(b))
			return nil
		}
		return gowa.New(c.Providers.GoWA).Notify(context.Background(), n)
	case "install", "uninstall", "doctor", "test", "config", "logs":
		fmt.Printf("%s command available in packaged installations\n", args[0])
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}
