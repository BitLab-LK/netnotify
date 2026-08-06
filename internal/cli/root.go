package cli

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/bitlab-dev/netnotify/internal/app"
	"github.com/bitlab-dev/netnotify/internal/config"
	"github.com/bitlab-dev/netnotify/internal/heartbeat"
	"github.com/bitlab-dev/netnotify/internal/logger"
)

var Version = "0.2.0"

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
	case "health", "test", "ping":
		c, err := config.Load(cfgFile)
		if err != nil {
			return err
		}
		if !c.Heartbeat.Enabled || c.Heartbeat.URL == "" {
			return fmt.Errorf("heartbeat is disabled or missing URL")
		}
		log := logger.New(c.Log)
		svc := heartbeat.New(c.Heartbeat, log)
		fmt.Printf("Sending test heartbeat ping to %s...\n", c.Heartbeat.URL)
		err = svc.Ping(context.Background(), c.Heartbeat.URL)
		if err == nil {
			fmt.Println("Heartbeat ping successful!")
		}
		return err
	case "install", "uninstall", "doctor", "config", "logs":
		fmt.Printf("%s command available in packaged installations\n", args[0])
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}
