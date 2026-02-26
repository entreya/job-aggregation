# 🇮🇳 Self-Hosted Runner Setup Guide (Oracle Cloud Free Tier)

## Overview
Run the scraper on a **free forever** Oracle Cloud ARM VM in Mumbai, India. This completely bypasses GitHub's Azure datacenter IP blocks and geo-restrictions on `recruitment.nic.in`.

## 💰 Cost
**$0/month — forever.** Oracle Cloud's "Always Free" tier includes:
- **ARM A1 Flex**: Up to 4 OCPUs + 24GB RAM
- **200GB** block storage
- **10TB/month** outbound data
- **Mumbai, India** datacenter 🇮🇳

---

## 🚀 Step 1: Create Oracle Cloud Account

1. Go to [cloud.oracle.com](https://cloud.oracle.com) and sign up.
2. **⚠️ Select Home Region → India South (Mumbai)**. This cannot be changed later!
3. A credit card is required for identity verification — **you will not be charged**.
4. (Recommended) After signup, upgrade to "Pay As You Go" to avoid capacity issues. You still only pay if you **explicitly** create paid resources.

---

## 📦 Step 2: Create ARM VM Instance

1. Go to **Compute → Instances → Create Instance**.
2. Configure:
   | Setting       | Value                              |
   |--------------|-----------------------------------|
   | **Name**      | `job-scraper-runner`               |
   | **Image**     | Ubuntu 22.04 Minimal (aarch64)     |
   | **Shape**     | VM.Standard.A1.Flex (Ampere ARM)   |
   | **OCPUs**     | 1 (more than enough for cron)      |
   | **Memory**    | 6 GB                               |
   | **Boot Vol.** | 50 GB                              |
3. Add your **SSH public key**.
4. Click **Create**.

> **Tip:** If you get "Out of Capacity" errors, retry at different times of day. ARM instances are popular. Upgrading to Pay As You Go often helps.

5. SSH into your instance:
   ```bash
   ssh ubuntu@<your-instance-ip>
   ```

---

## 🔧 Step 3: Install Dependencies (Go & Chromium)

Since Oracle's free tier uses ARM (aarch64), Google Chrome is **not available**. We use Chromium instead.

```bash
# 1. Update System
sudo apt update && sudo apt upgrade -y
sudo apt install -y curl wget git

# 2. Install Go 1.23+ (ARM64)
wget https://go.dev/dl/go1.23.0.linux-arm64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.23.0.linux-arm64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile
source ~/.profile
go version  # Verify

# 3. Install Chromium (ARM-compatible browser for Chromedp)
sudo apt install -y chromium-browser
chromium-browser --version  # Verify

# 4. (Optional) Set CHROME_PATH for the scraper
echo 'export CHROME_PATH=/usr/bin/chromium-browser' >> ~/.profile
source ~/.profile
```

---

## 🏃 Step 4: Configure GitHub Actions Runner (ARM64)

1. Go to your GitHub Repo → **Settings** → **Actions** → **Runners**.
2. Click **New self-hosted runner**.
3. Select **Linux** and **ARM64**.
4. Run the commands provided by GitHub on your VPS:

```bash
# Create folder
mkdir actions-runner && cd actions-runner

# Download ARM64 Runner (use the URL provided by GitHub)
curl -o actions-runner-linux-arm64-2.x.tar.gz -L <GITHUB_PROVIDED_URL>

# Extract
tar xzf ./actions-runner-linux-arm64-2.x.tar.gz

# Configure
./config.sh --url https://github.com/entreya/job-aggregation --token <YOUR_TOKEN>
```

5. **Install as Service** (runs 24/7 across reboots):
   ```bash
   sudo ./svc.sh install
   sudo ./svc.sh start
   ```
6. Go back to GitHub Settings. The runner should appear as **"Idle"**.

---

## 🔄 Step 5: Use the VPS Workflow

The workflow `.github/workflows/scraper-vps.yml` will:
1. Auto-detect Chromium (or Chrome) on the runner.
2. Pass the browser path via `CHROME_PATH` env to the scraper.
3. Run every 6 hours via cron.
4. Commit scrape results back to the repo.
5. Create a GitHub Issue if it fails.

```yaml
runs-on: self-hosted  # Routes to your Oracle Cloud VM
```

---

## 🛠️ Maintenance

| Task                  | Details                                           |
|----------------------|--------------------------------------------------|
| **Runner Updates**    | Auto-updates itself                              |
| **Scraper Schedule**  | Every 6 hours via cron in `scraper-vps.yml`      |
| **Logs**              | Visible in GitHub Actions tab                     |
| **VM Reboot**         | Runner service auto-starts via `svc.sh`          |
| **Oracle Dashboard**  | Monitor at [cloud.oracle.com](https://cloud.oracle.com) |

---

## ⚠️ ARM-Specific Notes

| Item                  | x64 (Paid VPS)               | ARM (Oracle Free)                |
|----------------------|-----------------------------|----------------------------------|
| **Browser**           | `google-chrome`              | `chromium-browser`               |
| **Go Binary**         | `go1.23.0.linux-amd64`      | `go1.23.0.linux-arm64`          |
| **Runner Binary**     | `actions-runner-linux-x64`   | `actions-runner-linux-arm64`    |
| **Go Cross-Compile**  | `GOARCH=amd64`               | `GOARCH=arm64`                  |
