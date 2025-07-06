# Online store consultant

## Add your files

- [ ] [Create](https://docs.gitlab.com/ee/user/project/repository/web_editor.html#create-a-file) or [upload](https://docs.gitlab.com/ee/user/project/repository/web_editor.html#upload-a-file) files
- [ ] [Add files using the command line](https://docs.gitlab.com/topics/git/add_files/#add-files-to-a-git-repository) or push an existing Git repository with the following command:

```
cd existing_repo
git remote add origin https://gitlab.pg.innopolis.university/r.muliukin/online-store-consultant.git
git branch -M main
git push -uf origin main
```

## Integrate with your tools

- [ ] [Set up project integrations](https://gitlab.pg.innopolis.university/r.muliukin/online-store-consultant/-/settings/integrations)

## Collaborate with your team

- [ ] [Invite team members and collaborators](https://docs.gitlab.com/ee/user/project/members/)
- [ ] [Create a new merge request](https://docs.gitlab.com/ee/user/project/merge_requests/creating_merge_requests.html)
- [ ] [Automatically close issues from merge requests](https://docs.gitlab.com/ee/user/project/issues/managing_issues.html#closing-issues-automatically)
- [ ] [Enable merge request approvals](https://docs.gitlab.com/ee/user/project/merge_requests/approvals/)
- [ ] [Set auto-merge](https://docs.gitlab.com/user/project/merge_requests/auto_merge/)

## Test and Deploy

Use the built-in continuous integration in GitLab.

- [ ] [Get started with GitLab CI/CD](https://docs.gitlab.com/ee/ci/quick_start/)
- [ ] [Analyze your code for known vulnerabilities with Static Application Security Testing (SAST)](https://docs.gitlab.com/ee/user/application_security/sast/)
- [ ] [Deploy to Kubernetes, Amazon EC2, or Amazon ECS using Auto Deploy](https://docs.gitlab.com/ee/topics/autodevops/requirements.html)
- [ ] [Use pull-based deployments for improved Kubernetes management](https://docs.gitlab.com/ee/user/clusters/agent/)
- [ ] [Set up protected environments](https://docs.gitlab.com/ee/ci/environments/protected_environments.html)

***


## Name
Online store consulant.

## Description
Our project provides the best solution to the pain point of online consultations in the shops. Our team offers the chat-bot that acts as a professional, answers in a simple user-friendly language without complex terms, remembers the shop stock, recall the dialogue history, and even communicates in a way that is indistinguishable from a human.


### Background

Initially, the consultants that worked locally at a specific shop had to answer to both online and offline clients. As a result, the waiting time increased dramitically while the response quality decreased accordingly.

### Features
```
* The consultant will only answer questions about the products or direct the client to those topics.
* The possibility of applying filters.
* The service works without using a VPN.
* The possibility of calling a human consultant.
* Support for additional functionality for registered users.
* Service analytics (processing speed, quality of service).
```

## Development
Our policies:
```
* We used the Trunk-based development with short-lived branches attached to a specific issue.
* At least one team member must have provided the useful feedback regarding the MR.
* The QA was performed via pipelining each time the changes in the codebase were introduced.
```

### Kanban board
The link to our Kanban board:
https://gitlab.pg.innopolis.university/r.muliukin/online-store-consultant/-/boards

#### Columns
There are entry criteria for each column. An issue can be closed when it reaches the ```Done``` column.

#### To Do
```
[Entry Criteria]
* The issue is unambigiously formulated as the issue form template.
* The issue is unanimously estimated in story points by all team members.
* The label is attached to the issue.
* The issue is desided to be in the sprints.
```

#### In Progress
```
[Entry Criteria]
* The issue description was revised to provide missing details.
* The issue was added to the current sprint.
```

#### Ready to Deploy
```
[Entry criteria]
* The MR attached to the issues passes the pipeline stage.
* The MR is approaved by at least one team member.
* All acceptance criterias are satisfied.
* All issues on which it depends are done.
```

#### Closed/Done
```
[Entry criteria]
* Also the previous development stages are passed.
* The code is merged into the main branch.
```

### **Git workflow**

---

#### Issue Management

- **Creating Issues**  
  Use our predefined issue templates when creating a new issue:
    - [Bug Report Template](./ISSUE_TEMPLATE/bug_report.md)
    - [Feature Request Template](./ISSUE_TEMPLATE/feature_request.md)
    - [Task Template](./ISSUE_TEMPLATE/task.md)

- **Labelling Issues**  
  Use consistent labels to categorize issues:
    - `bug`, `feature`, `task`, `enhancement`
    - `priority: high`, `priфority: medium`, `priority: low`
    - `status: in progress`, `status: blocked`, `status: ready for review`

- **Assigning Issues**  
  Team members self-assign issues or are assigned by the project lead. Each issue should have at least one assignee responsible for completion.


#### **Branching Rules**
- **`main` branch**: Protected, used only for stable releases.
- **Feature branches**: Created from `main` for each task using the naming convention:
  feature/ISSUE_ID-short-description
#### **Commit Message Format**
- Follow **Conventional Commits**: type(issue-id): description
  **Types:**
    - `feat`: New feature
    - `fix`: Bug fix
    - `docs`: Documentation changes
    - `refactor`: Code improvements (no new features)
    - `test`: Test-related changes
#### **Merge Request (MR) Process**
- When a task is ready, create an MR from the feature branch to `main`.
- **MR Title Format:** type(#issue-id): short-description
  **Types:**
    - `feature`: New features or user-facing functionality.
    - `bugfix`: Bug fixes or critical patches.
    - `documentation`: Documentation updates (READMEs, comments, wikis).
    - `refactor`: Code improvements (non-breaking, no new features).
    - `testing`: Test additions/improvements (unit, integration, e2e)
      **MR Description Template:**
      Description  
      [Briefly describe changes]

  Changes  
  [List key changes]

  Testing Steps  
  [How to test the changes]

  Closes issue-id

#### **Code Review Rules**
- **Minimum 1 reviewer** required before merging.
- Reviewers should:
    - Check for code quality, logic errors, and edge cases.
    - Leave **constructive comments** (e.g., _"Add error handling for empty input in line 45"_).
    - The author updates the branch if changes are requested.
    - MR is merged only after:
    - All comments are resolved.
    - CI/CD pipelines pass (if configured).

#### Gitgraph Diagram

![Git workflow diagram](./docs/images/diagram.png)
---

### Secret management
We thoroughly look after our secret data such as telegram bot token and DeepSeek API key. We prioriotize the safety in our project, that's why we use .env file to keep all these data. The bot token and DeepSeek API is shared only with the team members in Telegram PM. The data sharing happens iff the keeper of this secret data is sure that he/she is not being contacted by a fraudster.

## Quality assurance
### Quality attribute scenarios
See detailed quality attribute scenarios [here](docs/quality-assurance/quality-attribute-scenarios.md)

### Automated tests
We use the following tools and practices for automated quality assurance:

- **Tools used**:
  - [`sqlmock`](https://github.com/DATA-DOG/go-sqlmock) – for mocking database queries.
  - [`httptest`](https://pkg.go.dev/net/http/httptest) – for testing HTTP handlers.
  - `testing` – Go standard testing framework.
  - `go test` – to run test suites via CLI or CI.

- **Types of tests implemented**:
  - **Unit tests** – for individual HTTP handlers (`addProduct`, `addMessage`, `createSession`), verifying logic in isolation and
  including DB calls through mocks.
  - **User acceptance tests** - testing the work of the whole project by simulating user's
interaction with it.
  - **Integration-style tests** – simulate end-to-end HTTP interaction via `httptest`, including DB calls.
  - **User assistance behavior checks** – verify that the system gives meaningful help when queries are vague or invalid.
- **Test file locations**:
  - `unit_tests.go` – main unit test suite for session logic and database operations.
  - `DBBasicRequests-unit_test.go` – unit tests for some database requests such as `addProduct`, `addMessage`, etc.
  - `DBHandler_test.go` - integration test suite for database operations.
All tests can be executed using:
```bash
go test ./...
```
Tests run automatically on push due to pipeline (view [Build and deployment](#Build-and-deployment))
## Build and deployment

### Continuous Integration
Our project uses GitLab CI to automate the **linting**, **building**, **testing**, and **vulnerability scanning** processes. This ensures high code quality and early detection of issues.

#### CI pipeline file
[`gitlab-ci.yml`](./.gitlab-ci.yml)

#### Static analysis & testing tools used in CI
| Tool               | Purpose                                                                 |
|--------------------|-------------------------------------------------------------------------|
| `staticcheck`      | Performs advanced static analysis of Go code to catch bugs and issues.  |
| `govulncheck`      | Detects known vulnerabilities in Go modules and standard library usage. |
| `go test`          | Runs unit and integration tests.                                        |
| `gocover-cobertura`| Converts Go coverage profiles to Cobertura XML format for reporting.    |
| `Trivy`            | Scans Docker images for vulnerabilities, CVEs, and misconfigurations.   |

#### View all CI pipeline runs
- [CI/CD Pipelines](https://gitlab.pg.innopolis.university/r.muliukin/online-store-consultant/-/pipelines)

### Continuous Deployment
Continuous deployment (CD) is **not yet enabled**.

However, the CI pipeline **builds and pushes Docker images** to a Harbor registry, making them ready for deployment:

- `$HARBOR_HOST/$HARBOR_PROJECT/backend:$CI_JOB_ID`
- `$HARBOR_HOST/$HARBOR_PROJECT/bot:$CI_JOB_ID`

## Architecture

### Static view

Here you can find the diagram: [Component diagram](https://gitlab.pg.innopolis.university/r.muliukin/online-store-consultant/-/blob/main/docs/architecture/static-view/component-diagram.png?ref_type=heads), and here is the code for PlantUML: [Code for the diagram](docs/architecture/static-view/component-diagram.txt)

In this diagram we have decomposed the application into key components:
1. Web API – handles REST requests from the client.
2. Session Manager – responsible for creating/storing anonymous and authorized sessions.
3. Message Store – dynamically created tables anonymous_messages_<id> and user_messages_<id>.
4. DeepSeek Adapter is a component that proxies requests to the external DeepSeek API.
5. Postgres DB – storage of all metadata (tables anonymous_sessions, user_sessions, users, popular_products).

#### Coupling and Cohesion

Our component breakdown maximizes cohesion by grouping related functionality together - the Session Manager owns only session‐lifecycle concerns, the Message Store handles only persistence of chat messages, and the DeepSeek Adapter is responsible only for proxying AI calls. Coupling between components is kept intentionally low: each interaction happens over a well-defined interface (e.g. HTTP for the Web API, CRUD or SELECT queries for the database, and a single “fetchContext” call to DeepSeek), meaning changes in one component rarely ripple into others.

#### Maintainability:

Modularity: by splitting functionality into discrete components, we can replace or upgrade one piece (for example, swapping out Postgres for another store) without touching business logic.

Reusability: core services (Session Manager, Message Store, DeepSeek Adapter) are self-contained and expose generic interfaces, so they could be reused in a different front-end with minimal wiring.

Analyzability: each component’s responsibilities are narrow and documented, making it straightforward to trace the root cause of issues. Logs and metrics can be viewed for each component.

Modifiability: adding new features only requires touching one adapter or one service, not the entire codebase.

Testability: we cover each component with unit tests (e.g. Go’s testing, Testify, and go-sqlmock for database mocks) and integration tests (using Testcontainers). The clear boundaries and mockable interfaces mean we can exercise session logic and AI calls in isolation, ensuring high confidence in changes.

### Dynamic view

Here you can find the diagram: [Sequence diagram](https://gitlab.pg.innopolis.university/r.muliukin/online-store-consultant/-/blob/main/docs/architecture/dynamic-view/sequence-diagram.png?ref_type=heads), and here is the code for PlantUML: [Code for the diagram](docs/architecture/dynamic-view/sequence-diagram.txt)

In this scenario an anonymous user converts to an authenticated session, bringing along all their previous messages and context. It exercises the following components and steps:
1. Web API accepts the POST /api/login call with {login, password}.
2. Database is queried for the matching user_id and any prior session_id.
3. Session Service: allocates a fresh user_sessions entry for the logged-in user, copies all rows from anonymous_messages_<oldSessID> into user_messages_<newSessID>, drops the old anonymous message table and deletes the anonymous session record.
4. Web API returns a JSON response back to the client: { "response": "You have successfully logged in" }.

We measured end-to-end latency of this flow (anonymous→login, data migration, DeepSeek call, HTTP response) on our production instance from 5 sequential requests. The average time is 1.26 seconds.

### Deployment view

Here you can find the diagram: [Deployment diagram](https://gitlab.pg.innopolis.university/r.muliukin/online-store-consultant/-/blob/main/docs/architecture/deployment-view/deployment-diagram.png?ref_type=heads), and here is the code for PlantUML: [Code for the diagram](docs/architecture/deployment-view/deployment-diagram.txt)

#### Description:
We combine three services on a single Docker bridge network via docker-compose.yml:
1. PostgreSQL (postgres-container), persisting all user, session and message data.
2. Go Application (go-app-container), exposing the HTTP API (port 8080), managing sessions, business logic, and AI calls.
3. Telegram Bot (tg-bot-container), which stores state in the same Postgres DB, and invokes the HTTP API for core operations.

At runtime the bot and app communicate directly with the database over TCP port 5432. The bot also makes HTTPS calls to the Go API, and the Go API makes outbound HTTPS requests to DeepSeek’s external service for AI completions. This deployment using Docker-Compose only ensures that the customer will also be able to deploy the service using docker-compose up.

#### Legend: 
Solid arrows denote HTTPS connections, dashed arrows denote internal TCP connections; all containers live inside a single Docker bridge network (the big box), while the external DeepSeek API and the client sit outside.

## Badges
On some READMEs, you may see small images that convey metadata, such as whether or not all the tests are passing for the project. You can use Shields to add some to your README. Many services also have instructions for adding a badge.

## Visuals
Depending on what you are making, it can be a good idea to include screenshots or even a video (you'll frequently see GIFs rather than actual videos). Tools like ttygif can help, but check out Asciinema for a more sophisticated method.

## Installation
```

1) Clone the git repository into your IDE that supports Go programming language and docker deployment.
2) Create .env with two environment variables in the root of the project with the API_KEY and BOT_TOKEN fields.
3) Put in these fields valid DeepSeek API key retreived from https://platform.deepseek.com and bot token for your created bot via @BotFather in Telegram respectively.
4) Having installed Docker on your computer, run docker compose up --build in the IDE Terminal and enjoy the service running.
```

## Usage
To use our product you should:
```
* /start chat with our bot avalaible as @consultant_radad_bot in Telegram
* After that, you should follow the instructions that the bot asks you to do (language selection, registration, product selection, session starting)
* Finally, you can maintain the conversation with the consultant typing any messages you want or use any avalaible commands
Please Note:
* You can view list of all commands avalaible via Telegram widget "Menu"
```

## Support
In case of any misunderstanding and/or technical issues write in the PM to our developer using alias from /help command

## Roadmap
If you have ideas for releases in the future, it is a good idea to list them in the README.

## Contributing
State if you are open to contributions and what your requirements are for accepting them.

For people who want to make changes to your project, it's helpful to have some documentation on how to get started. Perhaps there is a script that they should run or some environment variables that they need to set. Make these steps explicit. These instructions could also be useful to your future self.

You can also document commands to lint the code or run tests. These steps help to ensure high code quality and reduce the likelihood that the changes inadvertently break something. Having instructions for running tests is especially helpful if it requires external setup, such as starting a Selenium server for testing in a browser.

## Authors and acknowledgment
Show your appreciation to those who have contributed to the project.

## License
For open source projects, say how it is licensed.

## Project status
If you have run out of energy or time for your project, put a note at the top of the README saying that development has slowed down or stopped completely. Someone may choose to fork your project or volunteer to step in as a maintainer or owner, allowing your project to keep going. You can also make an explicit request for maintainers.