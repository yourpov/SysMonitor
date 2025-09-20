package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	logger "github.com/yourpov/logrite"
)

type Config struct {
	Token  string `json:"token"`
	Prefix string `json:"prefix"`
	Status string `json:"status"`
}

// loadConfig returns the config
func loadConfig() Config {
	f, err := os.Open("config/config.json")
	if err != nil {
		logger.Error("open config/config.json: %v", err)
		os.Exit(1)
	}
	defer f.Close()

	var c Config
	if err := json.NewDecoder(f).Decode(&c); err != nil {
		logger.Error("parse config/config.json: %v", err)
		os.Exit(1)
	}
	if c.Prefix == "" {
		c.Prefix = "!"
	}
	if c.Token == "" {
		logger.Error("missing bot token in config/config.json")
		os.Exit(1)
	}
	return c
}

// formatBytes formats bytes to a readable string
func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

// safer hostname (fallback if host.Info() fails)
func hostNameSafe() string {
	if hi, err := host.Info(); err == nil && hi != nil && hi.Hostname != "" {
		return hi.Hostname
	}
	if hn, err := os.Hostname(); err == nil && hn != "" {
		return hn
	}
	return "unknown"
}

// buildEmbed collects system stats and returns a Discord embed
func buildEmbed() (*discordgo.MessageEmbed, error) {
	// Memory
	vm, _ := mem.VirtualMemory()

	cps, _ := cpu.Percent(time.Second, false)
	cpuPct := 0.0
	if len(cps) > 0 {
		cpuPct = cps[0]
	}

	hostname := hostNameSafe()
	uptime, err := host.Uptime()
	if err != nil {
		uptime = 0
	}

	rootPath := "/"
	if runtime.GOOS == "windows" {
		drive := os.Getenv("SystemDrive")
		if drive == "" {
			drive = "C:"
		}
		rootPath = drive + `\`
	}
	var used, total uint64
	var usedPct float64
	if du, err := disk.Usage(rootPath); err == nil && du != nil {
		used, total, usedPct = du.Used, du.Total, du.UsedPercent
	}

	fields := []*discordgo.MessageEmbedField{
		{Name: "Total Memory", Value: formatBytes(vm.Total), Inline: true},
		{Name: "Free Memory", Value: formatBytes(vm.Available), Inline: true},
		{Name: "Used Memory", Value: formatBytes(vm.Used), Inline: true},

		{Name: "Memory %", Value: fmt.Sprintf("%.1f%%", vm.UsedPercent), Inline: true},
		{Name: "CPU Usage", Value: fmt.Sprintf("%.1f%%", cpuPct), Inline: true},
		{Name: "Disk Used", Value: func() string {
			if total == 0 {
				return "n/a"
			}
			return fmt.Sprintf("%s / %s (%.1f%%)", formatBytes(used), formatBytes(total), usedPct)
		}(),
			Inline: true,
		},

		{Name: "Go / OS", Value: fmt.Sprintf("%s / %s-%s", runtime.Version(), runtime.GOOS, runtime.GOARCH), Inline: true},
		{Name: "Uptime", Value: Uptime(uptime), Inline: true},
	}

	return &discordgo.MessageEmbed{
		Title:     "System Statistics!",
		Color:     0x5865F2,
		Author:    &discordgo.MessageEmbedAuthor{Name: hostname + " (@" + strings.ToLower(hostname) + ")"},
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: "https://avatars.githubusercontent.com/u/59181303?v=4"},
		Fields:    fields,
		Footer:    &discordgo.MessageEmbedFooter{Text: "https://github.com/yourpov/SysMonitor"},
	}, nil
}

// Uptime converts seconds to "Xd Yh Zm"
func Uptime(sec uint64) string {
	d := time.Duration(sec) * time.Second
	days := d / (24 * time.Hour)
	d -= days * 24 * time.Hour
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute

	parts := []string{}
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if h > 0 {
		parts = append(parts, fmt.Sprintf("%dh", h))
	}
	if m > 0 {
		parts = append(parts, fmt.Sprintf("%dm", m))
	}

	if len(parts) == 0 {
		return "err"
	}
	return strings.Join(parts, " ")
}

func main() {
	logger.SetConfig(logger.Config{
		ShowIcons:    true,
		UppercaseTag: true,
		UseColors:    true,
	})

	cfg := loadConfig()
	dg, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		logger.Error("discord session: %v", err)
		os.Exit(1)
	}

	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

	if cfg.Status != "" {
		dg.AddHandlerOnce(func(s *discordgo.Session, r *discordgo.Ready) {
			_ = s.UpdateWatchStatus(0, cfg.Status)
		})
	}

	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}
		if m.Content == cfg.Prefix+"stats" {
			embed, _ := buildEmbed()
			_, _ = s.ChannelMessageSendEmbed(m.ChannelID, embed)
		}
	})

	appURL := "https://discord.com/developers/applications"
	if u, err := dg.User("@me"); err == nil && u != nil {
		appURL = fmt.Sprintf("https://discord.com/developers/applications/%s/bot", u.ID)
	}
	if err := dg.Open(); err != nil {
		if strings.Contains(err.Error(), "4014") ||
			strings.Contains(strings.ToLower(err.Error()), "disallowed intent") {
			logger.Error("Missing Message Content Intent. Enable it here: %s", appURL)
			os.Exit(1)
		}
		log.Fatal("open:", err)
	}
	defer dg.Close()

	logger.Success("Bot running")
	logger.Info("type %sstats", cfg.Prefix)
	select {}
}
