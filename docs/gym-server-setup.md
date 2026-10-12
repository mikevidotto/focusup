# Gym server setup

This guide puts your 5/3/1 program on a Raspberry Pi so you can see the day's weights and
log your AMRAP reps from your iPhone at the gym. The desktop app reads and writes the same
copy, so the Pi becomes the one place your workouts live.

```
iPhone (at the gym) ──Tailscale──┐
                                 ├──> Raspberry Pi: gymserver :8787 ──> workouts.json
FocusUp desktop (at home) ───────┘
```

You'll need:

- The Raspberry Pi, on and connected to your home network, with SSH access.
- This repo on your computer, with Go and Node/npm installed (the same setup `wails dev` uses).
- About 20 minutes.

Throughout this guide, **`pi`** in commands means the SSH address of your Pi (for example
`mike@raspberrypi.local` or `mike@192.168.1.50`). Replace it with yours.

---

## 1. Check the Pi

SSH in and run:

```sh
ssh pi
uname -m        # aarch64 = 64-bit OS, armv7l = 32-bit OS
whoami          # your username on the Pi; you'll need it in step 4
echo $HOME      # your home directory, e.g. /home/mike
exit
```

Write down all three values.

## 2. Build the server (on your computer)

From the repo root:

```sh
# If uname -m said aarch64:
scripts/build-gym.sh

# If it said armv7l:
GOARCH=arm GOARM=7 scripts/build-gym.sh
```

This builds the phone page and then the server, with the page built in. The result is a
single file, `build/gymserver`. The Pi doesn't need Go or Node installed.

## 3. Copy the server and your program to the Pi

```sh
scp build/gymserver pi:~/
ssh pi 'mkdir -p ~/.config/focusup'
scp ~/.config/focusup/workouts.json pi:~/.config/focusup/
```

The second copy moves your current cycles, logs and training maxes over. From now on the
Pi's copy is the real one. The file on your computer becomes an offline cache.

**Quick test.** Run the server by hand:

```sh
ssh -t pi '~/gymserver'
```

You should see `gymserver listening on :8787`. While it's running, open
`http://<pi's LAN address>:8787` in a browser on your computer, for example
`http://raspberrypi.local:8787`. Your next workout should appear. Press `Ctrl+C` to stop it.

Keep the `-t`. Without it, Ctrl+C only closes the SSH connection and the server keeps
running on the Pi. That leftover copy holds port 8787, and the service in step 4 can't
start. To check, run `ssh pi pgrep -a gymserver`, which should print nothing.

## 4. Run it as a service

This keeps the server running all the time and restarts it after a crash or a reboot.

1. Copy the unit file to the Pi:

   ```sh
   scp docs/gymserver.service pi:~/
   ```

2. On the Pi, put your username (from step 1) and home directory into it. The file
   assumes the user `pi` and `/home/pi`. Replace `mike` here with your username:

   ```sh
   ssh pi
   sed -i 's/^User=pi$/User=mike/; s#/home/pi#/home/mike#g' ~/gymserver.service
   cat ~/gymserver.service     # check that User= and ExecStart= look right
   ```

3. Install and start it:

   ```sh
   sudo mv ~/gymserver.service /etc/systemd/system/
   sudo systemctl daemon-reload
   sudo systemctl enable --now gymserver
   systemctl status gymserver    # should say "active (running)"
   ```

If the Pi has a firewall (for example `ufw`), allow the port with `sudo ufw allow 8787/tcp`.

## 5. Reach the Pi from outside your home: Tailscale

The server has **no login**. Anyone who can reach port 8787 can see and change your
program, so **don't** open the port on your router. Tailscale creates a private network
that only your own devices can join, and it's free for personal use.

1. **Create an account** at <https://tailscale.com> (signing in with Google, Apple or
   GitHub works).

2. **On the Pi:**

   ```sh
   curl -fsSL https://tailscale.com/install.sh | sh
   sudo tailscale up
   ```

   Open the login link it prints and sign in with the same account.

3. **On your iPhone:** install **Tailscale** from the App Store, sign in with the same
   account, and allow the VPN configuration when iOS asks. Leave it switched on. It's
   built to use very little battery.

4. **Find the Pi's address.** In the Tailscale app or at
   <https://login.tailscale.com/admin/machines>, find the Pi. It has a name (for example
   `raspberrypi`) and an address starting with `100.`. With MagicDNS turned on (the default
   for new accounts), the name works as an address.

5. **Test it the way it'll work at the gym.** On the iPhone, turn **Wi-Fi off** and open
   Safari at:

   ```
   http://raspberrypi:8787
   ```

   Use your Pi's Tailscale name, or the `100.x.y.z` address if the name doesn't work. Your
   next workout should load over cellular.

## 6. Add it to your iPhone home screen

In Safari, with the page open: tap the **Share** button, then **Add to Home Screen**, then
**Add**. It opens full-screen like an app, called "Gym".

Using it:

- The top card shows the next workout with every set and weight. The AMRAP set is
  highlighted.
- After your last set, enter the reps and tap **Log workout**. The date defaults to today.
  Change it if you're logging later.
- **Undo** at the bottom removes the most recent log. It asks you to confirm first.

## 7. Point FocusUp at the Pi

On your computer, which is on your home network:

1. Open FocusUp and go to the **Health** tab.
2. Press `s` to open the server field.
3. Enter the server address and save. Either address works from home:
   - `http://raspberrypi.local:8787` (LAN address)
   - `http://raspberrypi:8787` (Tailscale name, if Tailscale is on the computer too)
4. A line under the header should read **"Synced with http://…"**.

From then on:

- Workouts logged on the phone show up the next time the FocusUp window gets focus.
- If the Pi can't be reached, the page says so and shows the last copy it saw. Changes are
  paused until the Pi is back, so the two copies can't drift apart.
- To go back to keeping the program only on this computer, press `s` and clear the field.

You're done.

---

## Updating the server later

After changing the workout code or the phone page, rebuild and swap the binary. Copy it to
a new name first, because Linux won't overwrite a running program:

```sh
scripts/build-gym.sh                       # add GOARCH=arm GOARM=7 for a 32-bit Pi
scp build/gymserver pi:~/gymserver.new
ssh pi 'mv ~/gymserver.new ~/gymserver && sudo systemctl restart gymserver'
```

Reload the page on your phone. If the home-screen app shows an old version, close it from
the app switcher and open it again.

## Backups

Your whole program is one file on the Pi: `~/.config/focusup/workouts.json`. To copy it to
your computer:

```sh
scp pi:~/.config/focusup/workouts.json ~/workouts-backup-$(date +%F).json
```

Or keep daily backups on the Pi itself with a cron job (`crontab -e` on the Pi):

```
0 3 * * * mkdir -p ~/backups && cp ~/.config/focusup/workouts.json ~/backups/workouts-$(date +\%F).json
```

## Troubleshooting

| Problem | What to check |
| --- | --- |
| Phone page won't load at the gym | Tailscale is on in the iPhone app; the Pi shows as connected in the Tailscale admin page; try the `100.x.y.z` address instead of the name. |
| Won't load at home either | `ssh pi systemctl status gymserver`; logs: `ssh pi journalctl -u gymserver -n 50`. |
| Page says "phone page not built" | The binary was built without running the page build. Use `scripts/build-gym.sh`, not plain `go build`. |
| Service keeps restarting with `status=1/FAILURE` | Read the actual error with `journalctl -u gymserver -n 20 --no-pager`. The most common one is `address already in use`: a hand-started copy is still running. Run `pkill -x gymserver && sudo systemctl restart gymserver`. |
| Service won't start: `exec format error` (`status=203/EXEC`) | Wrong architecture. Rebuild using the `uname -m` result from step 1. |
| Service won't start: permission or file errors | `User=` and the paths in `/etc/systemd/system/gymserver.service` must match your Pi user (step 4). Then run `sudo systemctl daemon-reload && sudo systemctl restart gymserver`. |
| FocusUp says it can't reach the server | Open the same URL in a browser on the computer. If that fails, the problem is the network or the Pi, not FocusUp. |
| FocusUp doesn't show a workout logged on the phone | Click away from the FocusUp window and back. It re-reads when the window gets focus. |
| "workouts must be logged in order" | Something was logged from the other device in the meantime. Reload the page and try again. |
