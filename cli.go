package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type cli struct {
	binName      string
	flags        cliFlags
	flagsOrder   []string
	hideDefValue []string
}

type cliFlags struct {
	protocol   string // Protocol (TCP or UDP)
	remotePort int    // Remote port
	localHost  string // Local host
	localPort  int    // Local port
	serverMode bool   // Server mode
	count      int    // Number of echo requests to send
	pattern    string // Pattern to be sent for echo
	timeout    string // Connection timeout
	deadline   string // Echo timeout
	echoPeriod string // Ping period in milliseconds
	debug      bool
	version    bool
}

type cliUsage struct {
	name     string
	flagType string
	usage    string
	defValue string
}

var cliUsageTemplate = `{{.Name}}

Usage:
   {{.BinName}} [options] [destination]

Options:
{{.Options}}
`

func (c *cli) execute() (*config, error) {
	c.binName = filepath.Base(os.Args[0])
	c.setFlags()
	c.setUsage()

	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(0)
	}

	flag.Parse()

	if c.flags.version {
		c.version()
		os.Exit(0)
	}

	conf := config{}
	err := conf.parse(c.flags, flag.Args())

	return &conf, err
}

func (c *cli) setFlags() {
	flag.StringVar(&c.flags.protocol, "p", "", "TCP or UDP `protocol`")
	flag.IntVar(&c.flags.remotePort, "r", -1, "Remote `port` number")
	flag.StringVar(&c.flags.localHost, "o", "", "Local `host` for client/server")
	flag.IntVar(&c.flags.localPort, "l", 0, "Local `port` number for client/server")
	flag.BoolVar(&c.flags.serverMode, "s", false, "Server mode enabled")
	flag.IntVar(&c.flags.count, "c", 5, "Number of echo requests to send (0 = infinite) 'count'")
	flag.StringVar(&c.flags.pattern, "a", c.getDefaultPattern(), "`Pattern` to be sent for echo")
	flag.StringVar(&c.flags.timeout, "t", "10s", "Connection `timeout`")
	flag.StringVar(&c.flags.deadline, "w", "5s", "Read/Write `deadline`")
	flag.StringVar(&c.flags.echoPeriod, "i", "100ms", "Time `interval` between sending each echo request")
	flag.BoolVar(&c.flags.debug, "d", false, "Print various debugging information")
	flag.BoolVar(&c.flags.version, "v", false, "Print program version and exit")

	c.flagsOrder = []string{"p", "r", "o", "l", "s", "c", "a", "t", "w", "i", "d", "v"}
	c.hideDefValue = []string{"r", "v"}
}

func (c *cli) setUsage() {
	flag.Usage = func() {
		tpl, err := template.New("cliUsage").Parse(cliUsageTemplate)
		if err != nil {
			panic(err)
		}

		err = tpl.Execute(flag.CommandLine.Output(), struct {
			Name    string
			BinName string
			Options string
		}{
			Name:    name,
			BinName: c.binName,
			Options: c.getOptions(),
		})

		if err != nil {
			panic(err)
		}
	}
}

func (c *cli) getOptions() string {
	opts := []cliUsage{}

	nPad, tPad := 0, 0
	for _, n := range c.flagsOrder {
		f := flag.Lookup(n)

		flagType, usage := c.getTypeAndUsage(f.Usage)
		opts = append(opts, cliUsage{
			name:     "-" + f.Name,
			flagType: flagType,
			usage:    usage,
			defValue: c.getDefaultValue(f.DefValue, f.Name),
		})

		if len := len(f.Name); len > nPad {
			nPad = len
		}

		if len := len(flagType); len > tPad {
			tPad = len
		}
	}

	opts = append(opts, cliUsage{
		name:  "-h",
		usage: "Print this help text and exit",
	})

	return c.formatOptions(opts, nPad, tPad)
}

func (c *cli) getTypeAndUsage(usage string) (string, string) {
	char := "`"
	if strings.Contains(usage, "'") {
		char = "'"
	}

	s := strings.Index(usage, char)
	if s == -1 {
		return "", usage
	}

	ns := usage[s+len(char):]
	e := strings.Index(ns, char)
	if e == -1 {
		return "", usage
	}

	if char == "'" {
		usage = strings.ReplaceAll(usage, char+ns[:e]+char, "")
	} else {
		usage = strings.ReplaceAll(usage, char, "")
	}

	return strings.ToLower(ns[:e]), strings.TrimSpace(usage)
}

func (c *cli) getDefaultValue(val, name string) string {
	if len(val) == 0 {
		return ""
	}

	for _, n := range c.hideDefValue {
		if name == n {
			return ""
		}
	}

	return fmt.Sprintf(" (default: %s)", val)
}

func (c *cli) formatOptions(options []cliUsage, nPad, tPad int) string {
	usage := ""
	for _, o := range options {
		t := nPad + 2 - len(o.name)
		u := tPad + 3 - len(o.flagType)

		usage = strings.Join([]string{
			usage,
			strings.Repeat(" ", 3), o.name,
			strings.Repeat(" ", t), o.flagType,
			strings.Repeat(" ", u), o.usage,
			o.defValue,
			"\n",
		}, "")
	}

	return strings.TrimSuffix(usage, "\n")
}

func (c *cli) getDefaultPattern() string {
	hostname, err := os.Hostname()
	if err != nil || len(hostname) == 0 {
		return fmt.Sprintf("%s %s", name, version)
	}

	return hostname
}

func (c *cli) version() {
	fmt.Fprintf(flag.CommandLine.Output(), "%s version %s", name, version)
}
