# GETTING STARTED — AhorraApp (Single Executable Guide, OpenCode + Windows)

> **This is the one and only guide to follow.** It takes you from a blank machine to a working
> MVP, using **OpenCode** (with your OpenCode Go subscription) as the AI agent, on **Windows**.
> There is no second document to cross-reference — everything you need is here, in order.
>
> **How to read it.** Do the steps strictly in order. Each step has:
> **(A)** the command or action, **(B)** what you should see if it worked, **(C)** what to do if
> it fails. Don't skip a step; later steps depend on earlier ones.
>
> **Time budget:** Phase 0 (installs) ~45–60 min the first time · Phase 1 ~15 min ·
> Phase 2 (build) is the bulk of the work, done epic by epic · Phase 3 (deploy) ~30 min, optional.
>
> **Symbols:** `$` = type this in PowerShell (don't type the `$`). **[YOU]** = you do it by hand.
> **[AGENT]** = you delegate it to OpenCode by pasting a prompt from `02_PROMPTS.md`.

---

# PHASE 0 — Install everything (one time)

> Goal: by the end of Phase 0, the six verification commands in step 0.8 all succeed.

## 0.1 — Accounts? Almost none. [YOU]

For the whole development stage you need **no cloud accounts** and pay nothing beyond the
OpenCode Go subscription you already have. The entire app — backend, database, cache, image
storage, OCR — runs locally in Docker.

- ✅ **OpenCode Go** — you already have this; it's your AI model provider.
- ⏳ **GitHub** — optional, only as an off-machine backup. Local Git works without it.
- ⏳ **Hetzner + domain** — only for Phase 3 (optional cloud deploy). Skip for now.

## 0.2 — Open a terminal [YOU]

This guide is for **Windows native** — no WSL, no Linux needed. Spec Kit ships PowerShell
scripts, and Flutter's Android tooling works best on native Windows.

Open **Windows Terminal** (recommended) or **PowerShell**: press Start, type "PowerShell",
open it.

**(B) Verify:** type `$ echo hello` → you see `hello`.

> 💡 Install Windows Terminal from the Microsoft Store for a nicer experience (tabs, better
> copy/paste). Optional.

## 0.3 — Install Git [YOU]

```
$ git --version
```
**(B)** You see `git version 2.x`. Skip ahead.
**(C) "is not recognized":** run `$ winget install --id Git.Git -e`, then close and reopen the
terminal and retry.

Set your identity (used for commits):
```
$ git config --global user.name "Your Name"
$ git config --global user.email "you@example.com"
```

## 0.4 — Install Go 1.23+ [YOU]

```
$ winget install --id GoLang.Go -e
```
Close and reopen the terminal, then:
```
$ go version
```
**(B)** `go version go1.23.x`.
**(C) "is not recognized":** close/reopen the terminal so PATH refreshes; retry.

## 0.5 — Install Docker Desktop [YOU]

```
$ winget install --id Docker.DockerDesktop -e
```
**Open the Docker Desktop app once** and leave it running (whale icon in the system tray). Then:
```
$ docker --version
$ docker run hello-world
```
**(B)** The second command prints "Hello from Docker!".
**(C) "Cannot connect to the Docker daemon":** Docker Desktop isn't running — open the app, wait
for "Docker is running", retry.

> 💡 Docker Desktop uses WSL2 as its engine under the hood, automatically. You run `docker`
> commands from normal PowerShell; you never open a Linux shell. If the installer offers the
> WSL2 backend, accept it — it's just Docker's engine.

## 0.6 — Install uv, OpenCode, and Spec Kit [YOU]

**uv** (Python tool manager that Spec Kit needs):
```
$ powershell -ExecutionPolicy ByPass -c "irm https://astral.sh/uv/install.ps1 | iex"
```
Close/reopen the terminal, then `$ uv --version` should print a version.

**OpenCode** (your AI agent):
```
$ winget install --id sst.opencode -e
```
Close/reopen the terminal, then `$ opencode --version` should print a version.

**Spec Kit** (the `specify` CLI, which drives the workflow):
```
$ uv tool install specify-cli --from git+https://github.com/github/spec-kit.git
```
Then `$ specify version` should print a version.
**(C) "specify is not recognized":** run `$ uv tool update-shell`, reopen the terminal, retry.

## 0.7 — Install Flutter (for the mobile app, Epic E6) [YOU]

Install the Windows version: https://docs.flutter.dev/get-started/install/windows . This also
installs **Android Studio** (Android SDK + emulator). Then:
```
$ flutter doctor
```
**(B)** Green checks for "Flutter" and "Android toolchain".
**(C)** Red items print their own fix command (e.g. `$ flutter doctor --android-licenses`). Fix
the Android-related ones. The "Xcode" red item is expected on Windows — it only matters for iOS,
which needs a Mac. You can fully build/test the **Android** app on Windows.

> 💡 You can do all the **backend** work (Phase 1 and most of Phase 2) without Flutter. If
> Flutter setup is slow, continue now and finish it before Epic E6.

## 0.8 — Verify Phase 0 [YOU]

All of these must print a version:
```
$ git --version
$ go version
$ docker --version
$ uv --version
$ opencode --version
$ specify version
```
All six succeed → Phase 0 done. ✅

---

# PHASE 1 — Create the project and connect OpenCode

> Goal: a clean project folder, Spec Kit wired for OpenCode, your design docs in place, and
> OpenCode connected to your Go models.

## 1.1 — (If restarting) back up the old attempt [YOU]

If you have a previous attempt to discard, rename it (don't lose it yet) and start fresh. From
the folder that contains your project folder:
```
$ cd D:\Desarrollo\2026-proyectos\Savemarket\src
$ Rename-Item ahorrapp ahorrapp_old_backup
```
*(Skip this step if this is a first-time setup.)*

## 1.2 — Create the project and initialize Git [YOU]

```
$ mkdir ahorrapp
$ cd ahorrapp
$ git init
```
**(B)** `$ git status` says "On branch main" (or master).

**From now on, run every command from inside this `ahorrapp` folder.** If you open a new
terminal later, return with `$ cd D:\Desarrollo\2026-proyectos\Savemarket\src\ahorrapp`.

## 1.3 — Initialize Spec Kit for OpenCode [YOU]

```
$ specify init --here --ai opencode
```
If it warns the folder isn't empty (because of `.git`), confirm yes or add `--force`. If it says
OpenCode isn't detected and you want to proceed, add `--ignore-agent-tools`. (If your Spec Kit
version uses the newer flag, the equivalent is `--integration opencode`; check with
`$ specify init --help`.)

**(B) Verify:**
```
$ ls .specify
$ ls .opencode\command
```
You should see the `.specify/` folder (Spec Kit's engine + memory) and `.opencode\command\`
containing the Spec Kit command files — these become the `/speckit.*` commands inside OpenCode.

**(C)** If `.opencode\command` is missing, re-run the init line above (confirm `specify version`
works first).

## 1.4 — Put the design documents in place [YOU]

Assuming the five `.md` files are in your Downloads folder:
```
$ mkdir docs
$ Copy-Item "$HOME\Downloads\00_GLOBAL_DESIGN.md" docs\ -Force
$ Copy-Item "$HOME\Downloads\01_GETTING_STARTED.md" docs\ -Force
$ Copy-Item "$HOME\Downloads\02_PROMPTS.md" docs\ -Force
$ Copy-Item "$HOME\Downloads\03_CONSTITUTION.md" docs\ -Force
```
Now place the constitution where the agent actually reads it (this **replaces** Spec Kit's
placeholder):
```
$ Copy-Item "$HOME\Downloads\03_CONSTITUTION.md" .specify\memory\constitution.md -Force
```
**(B) Verify:**
```
$ ls docs
$ Get-Content .specify\memory\constitution.md -TotalCount 3
```
You should see the design docs listed, and the constitution starting with "# AhorraApp
Constitution".

> Note: `03_CONSTITUTION.md` lives in **two** places on purpose — a readable copy in `docs\`, and
> the working copy at `.specify\memory\constitution.md` that the agent reads every phase.

## 1.5 — First clean commit [YOU]

```
$ git add .
$ git commit -m "chore: clean project init with Spec Kit (OpenCode) and design docs"
```

## 1.6 — Connect OpenCode to your Go models [YOU]

Start OpenCode from inside the project:
```
$ opencode
```
This opens OpenCode's terminal UI (TUI). Inside it:
1. Run `/connect`, select **OpenCode Go**, and paste your Go API key (get it from your OpenCode
   Zen console at https://opencode.ai if you don't have it handy). Keep the key private.
2. Run `/models` and pick a model. **Recommended for this project:** a strong agentic-coding
   model like **GLM** or **Kimi K2.7 Code** for the spec/plan/implement work; lighter models
   like **DeepSeek V4 Flash** or **Qwen Plus** stretch your usage limits for simple steps.

**(B) Verify:** ask OpenCode "say hello" in the TUI; it responds. If you get an auth error,
re-run `/connect` and re-paste the key.

**(C)** If `/connect` doesn't list OpenCode Go, confirm you're signed in to OpenCode Zen with an
active Go subscription, then retry.

> Your credentials are stored locally by OpenCode (in its auth file under your user profile),
> separate from the project — so the key never lands in Git.

## 1.7 — Confirm the workflow commands [YOU]

Still in the OpenCode TUI, type `/`. You should see `/speckit.constitution`, `/speckit.specify`,
`/speckit.plan`, `/speckit.tasks`, `/speckit.implement`. These five run the whole build.

**(C)** If they're missing, you likely started OpenCode outside the project folder — exit, `cd`
into `ahorrapp`, and run `$ opencode` again.

---

# PHASE 2 — Build the product, spec by spec

> The core of Spec-Driven Development. For each feature you repeat one loop:
>
> **`/speckit.specify` → review → `/speckit.plan` → review → `/speckit.tasks` →
> `/speckit.implement` → test → commit.**
>
> **Golden rule:** OpenCode does NOT write product code until you have read and **approved** the
> spec AND the plan. Approving = you read it, it matches your intent, it doesn't violate the
> constitution. That review is your job — it's what makes this reliable instead of "vibe coding".
>
> You type the `/speckit.*` commands **inside the OpenCode TUI** and paste the matching prompt
> body from **`docs/02_PROMPTS.md`**. Keep that file open beside you.

## 2.1 — Set the constitution [AGENT]

In the OpenCode TUI:
```
/speckit.constitution
```
…then paste **PROMPT 1**.
**[YOU] Review:** confirm the agent's consolidated constitution keeps the nine articles (Clean
Architecture with OCR and storage as replaceable ports, spec-first, tests, versioned REST
contracts, data/currency/normalization, simplicity and local-first, minimal security, ready
to grow without over-engineering, English as working language). Fix if needed, then have it
save.

## 2.2 — Specify the backend skeleton (Epic E1) [AGENT]

```
/speckit.specify
```
…then paste **PROMPT 2**.
**[YOU] Review the spec:** it must describe **behavior** (what `/health` does, what `docker
compose up` produces), not code. If it dictates implementation, tell the agent to keep it
behavioral. Approve when right.

## 2.3 — Plan the backend skeleton [AGENT]

```
/speckit.plan
```
…then paste **PROMPT 3**.
**[YOU] Review the plan:** it should use Go, PostgreSQL+PostGIS, Redis, MinIO, the Clean
Architecture folder layout, and Docker — and domain code must not depend on HTTP/DB. Approve.

## 2.4 — Generate tasks, then implement [AGENT]

```
/speckit.tasks
```
**[YOU] Skim the task list.** Then:
```
/speckit.implement
```
OpenCode writes the code, file by file.

## 2.5 — Test the backend skeleton [YOU]

From inside `ahorrapp` (a normal PowerShell window, not the TUI):
```
$ docker compose up --build
```
**(B)** Containers for API, PostgreSQL, Redis, and MinIO start without crashing. In another
PowerShell window (same folder):
```
$ curl http://localhost:8080/api/v1/health
```
You get a `200` describing the status of the DB and Redis. (Port may differ — check the agent's
output or `docker-compose.yml`.) MinIO's web console is usually at http://localhost:9001 — handy
for seeing uploaded receipt images later.

**(C)** If something crashes, copy the exact error and paste it to OpenCode: "docker compose up
fails with this error: <paste>. Fix it according to the plan." Re-run until `/health` returns 200.

Stop with `Ctrl+C`, then commit:
```
$ git add .
$ git commit -m "feat: backend skeleton (E1) passing health check"
```

## 2.6 — Build the remaining epics [AGENT + YOU]

Repeat the same loop for each epic, pasting the matching prompt from `02_PROMPTS.md`, testing,
and committing. **Do them top to bottom** — the order is required.

| Epic | What you build | Prompt | How you test it |
|------|----------------|--------|-----------------|
| **E2 — Auth** | Register/login with JWT | Reusable template in `02_PROMPTS.md` | Register a user with `curl`, get a token |
| **E3+E4+E5 — Receipts + OCR** | Upload, OCR, parse, editable review | **PROMPT 4**, then **PROMPT 4-bis** (plan) | Upload a real receipt; it reaches NEEDS_REVIEW with store + items |
| **E6 — Flutter app** | The mobile app | **PROMPT 5** | Run on your phone; scan a receipt; confirm |
| **E7+E8 — Price engine + ranking** | Averages + cheapest store | **PROMPT 6** | Search a product; see cheapest store |
| **E9 — Gamification** | Points for receipts | **PROMPT 7** | Confirm a receipt; earn points; resubmit grants none |
| **E10 — Hardening** | Rate limiting, validation, backups | Reusable template | Spam uploads; see them throttled |

Why the order: E2 before E3 (uploading needs a logged-in user); E3–E5 before E6 (the app needs
the API).

### Running the Flutter app (Epic E6 test) [YOU]
The agent generates the app in a subfolder (e.g. `mobile\`):
```
$ cd mobile
$ flutter pub get
$ flutter run
```
Connect your phone by USB (developer mode on) or start an emulator from Android Studio, then pick
the device when prompted. **(C)** No device found → `$ flutter devices`, connect a phone or start
an emulator. Return with `$ cd ..`.

## 2.7 — Local MVP checklist [YOU]

When all six are checked, your MVP works end to end on your machine — no cloud required:

- [ ] `docker compose up --build` starts API + Postgres + Redis + MinIO + OCR with no crashes.
- [ ] `curl .../api/v1/health` returns 200.
- [ ] I can register and log in from the Flutter app.
- [ ] I photograph a receipt and see the **editable summary**: store, date, line items.
- [ ] I correct an item, confirm, and **earn points**.
- [ ] I search a product and see **which store is cheaper**.

**This is the real milestone.** You can demo the whole product locally for $0. Phase 3 is
optional and only needed to put it online for other people.

---

# PHASE 3 — Deploy to the cloud (OPTIONAL)

> Skip this entirely until you want people outside your machine to use the app over the internet.
> This is the first point where you create cloud accounts and start paying (~$22/mo). If you're
> still developing locally, you're done after Phase 2.

## 3.1 — Create the server on Hetzner [YOU]

1. Hetzner Cloud Console → "New Project" → name it `ahorrapp`.
2. "Add Server": Location nearest your users · Image **Ubuntu 24.04** · Type **CX32** (~$8/mo) ·
   add your SSH key (or use the root password Hetzner emails you).
3. Note the server's **public IP**.

**(B) Verify** (replace `YOUR_IP`): `$ ssh root@YOUR_IP` → type `yes` to trust the host → you see
`root@ubuntu:~#`. Type `$ exit` to return.

## 3.2 — Install Coolify on the server [YOU]

Coolify is a web dashboard that deploys your Docker project with automatic HTTPS:
```
$ ssh root@YOUR_IP "curl -fsSL https://cdn.coollabs.io/coolify/install.sh | bash"
```
**(B)** When it finishes, open `http://YOUR_IP:8000` and create your Coolify admin account.
**(C)** Page won't load → wait 2–3 min, refresh, ensure `http://` and port `:8000`.

## 3.3 — Point your domain at the server [YOU]

In your registrar's DNS, add an **A record**: Host `api` → Value `YOUR_IP`.
**(B) Verify:** `$ ping api.yourdomain.com` resolves to `YOUR_IP` (wait a few minutes for DNS).

## 3.4 — Deploy through Coolify [YOU]

Push to GitHub first (create an empty repo on github.com):
```
$ git remote add origin https://github.com/YOUR_USERNAME/ahorrapp.git
$ git push -u origin main
```
Then in Coolify:
1. "New Resource" → connect GitHub → pick `ahorrapp`.
2. It detects `docker-compose.yml`; confirm the services (API, Postgres, Redis, MinIO/OCR).
3. Set the **environment variables** (DB URL, JWT secret, storage keys). Unsure? Ask the agent:
   "List every environment variable my backend needs for production and a safe example value."
4. On the API service, set domain `api.yourdomain.com` and enable HTTPS (Coolify gets a free
   Let's Encrypt cert automatically).
5. **Deploy.**

**(B)** Build logs finish, services "running", `https://api.yourdomain.com/api/v1/health` → 200.
**(C)** Build fails → copy Coolify's deploy logs to the agent for a fix; re-push and redeploy.

## 3.5 — Point the app at production [YOU + AGENT]

Tell the agent: "Change the Flutter app's API base URL to `https://api.yourdomain.com/api/v1` via
a config/env file, not hardcoded." Rebuild (`flutter run`) and confirm it talks to the live
server.

- [ ] Backend live at `https://api.yourdomain.com` with valid HTTPS.

---

# Troubleshooting (one place)

| Symptom | Likely cause | Fix |
|---------|--------------|-----|
| `is not recognized` after installing a tool | PATH not refreshed | Close/reopen PowerShell; for uv tools run `uv tool update-shell` |
| `Cannot connect to the Docker daemon` | Docker Desktop not running | Open Docker Desktop, wait for "running", retry |
| `/speckit.*` commands missing in OpenCode | Started OpenCode outside the project, or init didn't run | `cd` into `ahorrapp`, run `opencode`; if still missing, re-run `specify init --here --ai opencode` |
| `/connect` doesn't list OpenCode Go | Not signed in to Zen / Go inactive | Sign in to OpenCode Zen, confirm Go subscription, retry |
| Auth error running a command | Key not connected | Re-run `/connect`, re-paste the Go API key |
| Hitting usage limits mid-session | Model too heavy for the cap | Switch to a cheaper Go model via `/models` (DeepSeek V4 Flash, Qwen Plus), or enable "Use balance" in the Zen console |
| Generated code quality disappoints on a step | Model mismatch for the task | Switch to GLM or Kimi K2.7 Code via `/models` and re-run that step — files are in Git, so it's safe |
| `docker compose up` crashes | Config/code bug | Paste the exact error to the agent, ask it to fix per the plan |
| `flutter run` finds no device | No phone/emulator | `flutter devices`; connect a phone (USB debugging) or start an emulator |
| Coolify page won't load | Still starting / wrong URL | Wait 2–3 min; use `http://YOUR_IP:8000` |
| HTTPS cert fails in Coolify | DNS not propagated | Confirm `ping api.yourdomain.com` resolves, then retry |

---

# Between work sessions

- `git commit` after each working step. `git push` is optional (only if you set up a GitHub
  backup) — local commits already protect your history.
- When you return: `$ cd D:\Desarrollo\2026-proyectos\Savemarket\src\ahorrapp`, run `$ opencode`,
  and continue at the next unchecked epic.
- Keep `docs/00_GLOBAL_DESIGN.md` current if any decision changes — it's the project's memory
  that the agent re-reads each session.
- Once the new setup is solid, delete the backup: `$ Remove-Item ..\ahorrapp_old_backup -Recurse -Force`.

---

# A note on the model behind OpenCode Go

The Go plan's models are strong but are mostly open-source/Chinese-lab models (GLM, Kimi, Qwen,
MiniMax, DeepSeek) — not GPT/Claude/Gemini. For well-scoped, spec-driven tasks like these they
are perfectly capable. If a particular step underperforms, switch models with `/models` before
suspecting the spec. Because everything is file-based and committed to Git, you can even run the
same step with a different model and keep the better result — or open the project in another
agent entirely without losing any work.
