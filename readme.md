# Backend API for Open Movie Database — spec-kit Demo

This project demonstrates end-to-end AI-assisted development using [spec-kit](https://github.com/github/spec-kit): starting from a blank directory, a fully implemented golang based API was built entirely through structured agents in Claude Code. The steps below document exactly what was done — you can follow the same process to build your own project.

## Prerequisites

- [uv](https://docs.astral.sh/uv/) — Python package manager used to install spec-kit
- [VS Code](https://code.visualstudio.com/) with the [Claude Code](https://code.claude.com/docs/en/quickstart) extension

---

## Step 1 — Install spec-kit

```bash
uv tool install specify-cli --from git+https://github.com/github/spec-kit.git
```

## Step 2 — Initialize the project

This was run in the project directory:

```bash
specify init . --ai claude
```

This scaffolded the spec-kit structure and installed the Copilot custom agents.

## Step 3 — Open in VS Code

```bash
code .
```

## Step 4 — How to select a spec-kit agent

All spec-kit commands are available in the claude cli. Open Claude code CLI.

The agents available are:

| Command | Purpose |
|---|---|
| `speckit.constitution` | Define project-wide principles and governance |
| `speckit.specify` | Generate a feature specification from a description |
| `speckit.clarify` | Ask targeted questions to tighten an existing spec |
| `speckit.plan` | Produce a technical design and implementation plan |
| `speckit.analyze` | Check consistency across spec, plan, and tasks |
| `speckit.tasks` | Generate a dependency-ordered task list |
| `speckit.checklist` | Produce a custom quality checklist |
| `speckit.implement` | Execute tasks from `tasks.md` |
| `speckit.taskstoissues` | Convert tasks into GitHub Issues |

---

## Step 5 — Establish project principles
```
Fill the constitution with the bare miniumum requirements for a idiomatic golang based rest-api based backend. Follow Hexagonal architecture and clean code principles.
```

This update `.specify/memory/constitution.md`, which all subsequent agents respected.

---

## Step 6 — Write the feature specification

The **`speckit.specify`** agent was invoked with:

```
CineAPI is a JSON-based RESTful API for retrieving and managing information about movies. It serves as a comprehensive movie database service — similar in concept to the Open Movie Database API — with full support for user authentication, movie CRUD operations, and application observability.

Supported APIs
| Method | URL Pattern | Action |
|--------|-------------|--------|
| `GET` | `/v1/healthcheck` | Show application health and version information |
| `GET` | `/v1/movies` | Show the details of all movies |
| `POST` | `/v1/movies` | Create a new movie |
| `GET` | `/v1/movies/:id` | Show the details of a specific movie |
| `PATCH` | `/v1/movies/:id` | Update the details of a specific movie |
| `DELETE` | `/v1/movies/:id` | Delete a specific movie |
| `POST` | `/v1/users` | Register a new user |
| `PUT` | `/v1/users/activated` | Activate a specific user |
| `PUT` | `/v1/users/password` | Update the password for a specific user |
| `POST` | `/v1/tokens/authentication` | Generate a new authentication token |
| `POST` | `/v1/tokens/password-reset` | Generate a new password-reset token |
| `GET` | `/debug/vars` | Display application metrics |
```

This created `specs/<feature>/spec.md`.

---

## Step 7 — Create the implementation plan

The **`speckit.plan`** agent was invoked with:

```
I am going to use golang for restful API backend, use gorm and postgres database, use sample movie data. Use use chi router.
```

This created `specs/<feature>/plan.md` with a technical design tailored to the spec and the constitution.

---

## Step 9 — Generate the task list

The **`speckit.tasks`** agent was invoked with:

```
Execute
```

This created `specs/<feature>/tasks.md` with a dependency-ordered list of implementation tasks.

---

## Step 10 — Analyze consistency

The **`speckit.analyze`** agent was invoked with:

```
Execute
```

This performed a cross-artifact consistency check across `spec.md`, `plan.md`, and `tasks.md`. The agent identified gaps and offered to draft resolutions — both were accepted and applied.

> **Optional:** **`speckit.checklist`** can also be run at this stage to generate a custom quality checklist tailored to the feature.

---

## Step 11 — Implement

The **`speckit.implement`** agent was invoked with:

```
Execute
```

Large features are not always fully delivered in a single pass. For this project, two passes were needed:

**Pass 1** — `speckit.implement` was run with `Execute`. The agent completed a large portion of the tasks but reported completion prematurely.

**Pass 2** — `speckit.implement` was run again with `Execute`. The agent found failing tests and addressed them, then confirmed all tasks complete.

**Mop-up** — Additional agent-assisted Copilot work was needed to resolve runtime and integration issues stemming from context that was not provided to spec-kit upfront: JDK version, environment variable conventions, and API client configuration details. Providing that context in the plan prompt would have eliminated most of this work.

The number of passes will vary per project depending on scope and complexity.

---

## Result

The fully implemented LLM Performance Analytics Platform — a Spring Boot REST API with an embedded React dashboard, PostgreSQL persistence, Docker Compose deployment, and sample data — built from a one-sentence description in two implementation passes.

![Dashboard Screenshot](screenshot.png)