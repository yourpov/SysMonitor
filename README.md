<div align="center" id="top">

# SysMonitor

</div>
<p align="center">
  <img alt="Top language" src="https://img.shields.io/github/languages/top/yourpov/SysMonitor?color=56BEB8">
  <img alt="Language count" src="https://img.shields.io/github/languages/count/yourpov/SysMonitor?color=56BEB8">
  <img alt="Repository size" src="https://img.shields.io/github/repo-size/yourpov/SysMonitor?color=56BEB8">
  <img alt="License" src="https://img.shields.io/github/license/yourpov/SysMonitor?color=56BEB8">
</p>

---

## About

**SysMonitor** is a Golang Discord bot that fetches **system statistics** like memory usage, CPU load, disk space, and system uptime which is nice for monitoring your system directly from your server’s chat.  

## Tech Stack

- [Go](https://golang.org/)  
- [DiscordGo](https://github.com/bwmarrin/discordgo)  
- [gopsutil](https://github.com/shirou/gopsutil)  

---

## Setup

### Configuration

Edit `config/config.json`:

```json
{
  "token": "",
  "prefix": "!",
  "status": "watching system"
}
```

### Bot Setup

- Create an application and bot in the [Discord Developer Portal](https://discord.com/developers/applications).  
- Copy your **Bot Token** and paste it into `config/config.json`.  
- Enable **Message Content Intent** under *Privileged Gateway Intents*.  
- Invite your bot to your server (with `bot` and `Send Messages` permissions)  

### Running

```bash
# Clone & enter project
git clone https://github.com/yourpov/SysMonitor
cd SysMonitor

# Install dependencies
go mod tidy

# Run bot
go run main.go
```
