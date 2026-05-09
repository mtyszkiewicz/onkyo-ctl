package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/mtyszkiewicz/eiscp/internal/config"
	"github.com/mtyszkiewicz/eiscp/internal/pkg/eiscp"
	"github.com/urfave/cli/v3"
)

var client *eiscp.EISCPClient

func main() {
	cmd := &cli.Command{
		Name:  "onkyo",
		Usage: "Onkyo TX-L20D client",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "host",
				Aliases: []string{"H"},
				Usage:   "Onkyo host ip address",
				Sources: cli.EnvVars("ONKYO_HOST"),
			},
			&cli.StringFlag{
				Name:    "port",
				Aliases: []string{"P"},
				Usage:   "Onkyo host port",
				Sources: cli.EnvVars("ONKYO_PORT"),
			},
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			cfg, err := config.Load(os.Getenv("ONKYO_CONFIG"))
			if err != nil {
				return nil, fmt.Errorf("error loading config: %w", err)
			}

			host := cmd.String("host")
			if host == "" || host == "127.0.0.1" {
				host = cfg.Onkyo.Host
			}

			port := cmd.String("port")
			if port == "" || port == "60128" {
				port = cfg.Onkyo.Port
			}

			client, err = eiscp.NewEISCPClient(host, port, cfg.InputCodes(), cfg.InputNames())
			if err != nil {
				return nil, fmt.Errorf("error connecting to server: %w", err)
			}
			return nil, nil
		},
		After: func(ctx context.Context, cmd *cli.Command) error {
			if client != nil && client.Conn != nil {
				return client.Conn.Close()
			}
			return nil
		},
		EnableShellCompletion: true,
		Commands: []*cli.Command{
			{
				Name:  "chat",
				Usage: "Chat with onkyo using raw eiscp messages",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return StartChatSession(client)
				},
			},
			{
				Name:  "power",
				Usage: "Control device power",
				Commands: []*cli.Command{
					{
						Name:  "on",
						Usage: "Turn device on",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							return client.PowerOn()
						},
					},
					{
						Name:  "off",
						Usage: "Turn device off",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							return client.PowerOff()
						},
					},
				},
			},
			{
				Name:  "volume",
				Usage: "Control volume settings",
				Commands: []*cli.Command{
					{
						Name:  "query",
						Usage: "Query current volume level",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							result, err := client.QueryVolume()
							fmt.Print(result)
							return err
						},
					},
					{
						Name:  "set",
						Usage: "Set volume level",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							if cmd.Args().Len() != 1 {
								return fmt.Errorf("usage: volume set <level>")
							}
							level, err := strconv.Atoi(cmd.Args().First())
							if err != nil {
								return fmt.Errorf("invalid volume level: %w", err)
							}
							return client.SetMasterVolume(level)
						},
					},
					{
						Name:  "up",
						Usage: "Increase volume",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							return client.VolumeUp()
						},
					},
					{
						Name:  "down",
						Usage: "Decrease volume",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							return client.VolumeDown()
						},
					},
				},
			},
			{
				Name:  "subwoofer",
				Usage: "Control subwoofer settings",
				Commands: []*cli.Command{
					{
						Name:  "query",
						Usage: "Query current subwoofer level",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							result, err := client.QuerySubwooferLevel()
							fmt.Print(result)
							return err
						},
					},
					{
						Name:  "set",
						Usage: "Set subwoofer level",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							if cmd.Args().Len() != 1 {
								return fmt.Errorf("usage: subwoofer set <level>")
							}
							level, err := strconv.Atoi(cmd.Args().First())
							if err != nil {
								return fmt.Errorf("invalid subwoofer level: %w", err)
							}
							return client.SetSubwooferLevel(level)
						},
					},
				},
			},
			{
				Name:  "source",
				Usage: "Control input source",
				Commands: []*cli.Command{
					{
						Name:  "query",
						Usage: "Query current input source",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							result, err := client.QueryInputSelector()
							if err != nil {
								return err
							}
							fmt.Println(result)
							return nil
						},
					},
					{
						Name:  "set",
						Usage: "Set input source",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							if cmd.Args().Len() != 1 {
								return fmt.Errorf("usage: source set <source>")
							}

							source := strings.ToLower(cmd.Args().First())
							return client.SetInputSelector(source)
						},
					},
					{
						Name:  "list",
						Usage: "List available input sources",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							fmt.Println("Available sources:")
							for _, name := range client.ListInputs() {
								fmt.Printf("  - %s\n", name)
							}
							return nil
						},
					},
				},
			},
			{
				Name:  "brightness",
				Usage: "Set brightness level",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if cmd.Args().Len() != 1 {
						return fmt.Errorf("usage: brightness (0|1|2)")
					}
					level, err := strconv.Atoi(cmd.Args().First())
					if err != nil {
						return fmt.Errorf("invalid brightness level: %w", err)
					}
					return client.SetBrightness(level)
				},
			},
			{
				Name: "blink",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return client.AnimateBlink()
				},
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return cli.ShowAppHelp(cmd)
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}