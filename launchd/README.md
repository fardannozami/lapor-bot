# launchd Setup for macOS

This directory contains a launchd plist template for running lapor-bot as a macOS service.

## Features

- **Auto-start on login** - Bot starts automatically when you log in
- **Crash recovery** - Automatically restarts if the process crashes
- **Proper logging** - Stdout/stderr written to log files
- **Throttle protection** - Prevents rapid restart loops (10 second minimum between restarts)

## Installation

1. **Copy and customize the plist:**
   ```bash
   # Copy template to LaunchAgents
   cp launchd/com.fardannozami.lapor-bot.plist ~/Library/LaunchAgents/

   # Edit and update paths
   # Replace /path/to/lapor-bot with your actual path
   # Replace YOUR_USERNAME with your macOS username
   nano ~/Library/LaunchAgents/com.fardannozami.lapor-bot.plist
   ```

2. **Create logs directory:**
   ```bash
   mkdir -p /path/to/lapor-bot/logs
   ```

3. **Stop any existing nohup process:**
   ```bash
   pkill -f lapor-bot
   ```

4. **Load the service:**
   ```bash
   launchctl load ~/Library/LaunchAgents/com.fardannozami.lapor-bot.plist
   ```

## Management Commands

```bash
# Check if service is running
launchctl list | grep lapor-bot

# Stop the service
launchctl stop com.fardannozami.lapor-bot

# Start the service
launchctl start com.fardannozami.lapor-bot

# Unload (disable) the service
launchctl unload ~/Library/LaunchAgents/com.fardannozami.lapor-bot.plist

# Reload after config changes
launchctl unload ~/Library/LaunchAgents/com.fardannozami.lapor-bot.plist
launchctl load ~/Library/LaunchAgents/com.fardannozami.lapor-bot.plist
```

## Viewing Logs

```bash
# View stdout log
tail -f /path/to/lapor-bot/logs/lapor-bot.out.log

# View stderr log
tail -f /path/to/lapor-bot/logs/lapor-bot.err.log

# View both
tail -f /path/to/lapor-bot/logs/*.log
```

## Troubleshooting

**Service won't start:**
- Check paths in the plist are correct and absolute
- Ensure the binary exists and is executable: `chmod +x ./lapor-bot`
- Check logs for errors

**Process keeps restarting:**
- Check stderr log for crash reasons
- The service waits 10 seconds between restarts (ThrottleInterval)

**Environment variables not working:**
- launchd doesn't source your shell profile
- The bot reads from `.env` file in WorkingDirectory
- Ensure `.env` exists with required variables
