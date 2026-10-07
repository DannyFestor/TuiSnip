# Standards

How code is written in this repo. Before writing or reviewing code, read the sections the table below names for your change, not whole files.

| Standard | Covers |
|---|---|
| [architecture](architecture.md) | Hexagonal layers, package layout, Actions, capabilities, error wrapping, where context and logging go, the test tiers, TUI components, the files a feature touches |
| [code](code.md) | Conventions no linter checks: validation into value objects, context, concurrency, logging, comments, generated code |
| [testing](testing.md) | Choosing a tier, naming, table tests, test doubles, feature, e2e, fuzz, and property tests |
| [linting](linting.md) | Why the non-obvious golangci-lint rules exist, the thresholds, and the `nolint` format |
| [database](database.md) | Schema conventions, migrations, sqlc queries, transactions, connections, start-up |

## Read before

A change usually matches several rows. Read every section they name. To read one section alone, `grep -n '^#' <file>` gives the line where it starts. It ends before the next heading of the same or a higher level.

| When you | Read |
|---|---|
| write any Go | architecture: [Vocabulary](architecture.md#vocabulary). code: [Vocabulary](code.md#vocabulary), [Files and packages](code.md#files-and-packages), [Comments](code.md#comments). linting: [Small functions, happy path](linting.md#small-functions-happy-path), [Blank lines](linting.md#blank-lines), [Struct literals are complete](linting.md#struct-literals-are-complete), [No globals, no hidden state](linting.md#no-globals-no-hidden-state), [Comments and naming](linting.md#comments-and-naming) |
| add a feature that crosses layers | architecture: [Adding a feature](architecture.md#adding-a-feature), then the sections each step links |
| add a package | architecture: [Layout](architecture.md#layout), [Who owns what](architecture.md#who-owns-what), [Dependency rule](architecture.md#dependency-rule), [Adding a package](architecture.md#adding-a-package). code: [Files and packages](code.md#files-and-packages) |
| import across layers or add a third-party import | architecture: [Dependency rule](architecture.md#dependency-rule), [Enforcement](architecture.md#enforcement). linting: [Imports](linting.md#imports) |
| add or change an Action | architecture: [Actions](architecture.md#actions), [Interfaces](architecture.md#interfaces). code: [Actions](code.md#actions) |
| add or change an entity, a value object, or an input rule | code: [Domain types](code.md#domain-types) |
| add or change a TUI component, an outcome, or a Binding | architecture: [TUI components](architecture.md#tui-components), [Interfaces](architecture.md#interfaces), and step 6 of [Adding a feature](architecture.md#adding-a-feature) for a Binding. linting: [Type switches over a sealed union are exhaustive](linting.md#type-switches-over-a-sealed-union-are-exhaustive) |
| handle a click or the wheel | architecture: [Clicks arrive in the component's own cells](architecture.md#clicks-arrive-in-the-components-own-cells). testing: [TUI unit tests](testing.md#tui-unit-tests) |
| return, wrap, or show an error | architecture: [Errors](architecture.md#errors). code: [Errors](code.md#errors). linting: [Errors](linting.md#errors) |
| take or pass a context, or set a deadline | architecture: [Context](architecture.md#context). code: [Context](code.md#context) |
| log | architecture: [Logging](architecture.md#logging). code: [Logging](code.md#logging). linting: [Logging](linting.md#logging) |
| reach shared state from a `tea.Cmd` | code: [Concurrency](code.md#concurrency) |
| add an enum, a mock, a query, or another generated file | architecture: [Generated code and SQL](architecture.md#generated-code-and-sql). code: [Generated code](code.md#generated-code). linting: [Generated code](linting.md#generated-code) |
| write any test | architecture: [Tests](architecture.md#tests). testing: [Choosing the tier](testing.md#choosing-the-tier), [Naming](testing.md#naming), [Structure](testing.md#structure), [What to cover](testing.md#what-to-cover). linting: [Tests](linting.md#tests) |
| write a unit test | testing: [Test doubles](testing.md#test-doubles) |
| write a TUI unit test | testing: [TUI unit tests](testing.md#tui-unit-tests) |
| write a feature test | testing: [Feature tests](testing.md#feature-tests), [testkit](testing.md#testkit). database: [Tests](database.md#tests) |
| write an e2e test | testing: [e2e tests](testing.md#e2e-tests) |
| write a fuzz or property test | testing: [Fuzz and property tests](testing.md#fuzz-and-property-tests) |
| ask which lint rule applies, or fix a lint finding | linting: [Everything on, by exclusion](linting.md#everything-on-by-exclusion), then the section for the concern: [Small functions, happy path](linting.md#small-functions-happy-path), [Blank lines](linting.md#blank-lines), [Struct literals are complete](linting.md#struct-literals-are-complete), [Type switches over a sealed union are exhaustive](linting.md#type-switches-over-a-sealed-union-are-exhaustive), [No globals, no hidden state](linting.md#no-globals-no-hidden-state), [Errors](linting.md#errors), [Imports](linting.md#imports), [Logging](linting.md#logging), [Tests](linting.md#tests), [Comments and naming](linting.md#comments-and-naming), [Generated code](linting.md#generated-code), [Formatting](linting.md#formatting). A rule none of them explains is only in `.golangci.yml` |
| add a `//nolint` or change `.golangci.yml` | linting: [Everything on, by exclusion](linting.md#everything-on-by-exclusion), [`nolint`](linting.md#nolint) |
| change the schema or add a migration | database: [Stack](database.md#stack), [Schema](database.md#schema), [Migrations](database.md#migrations) |
| add or change a query or a repository | database: [Queries](database.md#queries), [Repositories](database.md#repositories), [Reading entities](database.md#reading-entities), [Writing and transactions](database.md#writing-and-transactions), [Errors](database.md#errors) |
| change how the database opens or starts up | database: [Connections](database.md#connections), [Start-up](database.md#start-up) |

Related documents:

- [`GLOSSARY.md`](../../GLOSSARY.md): the domain terms. They are binding for identifiers, test names, and log keys.
- [`docs/spec/`](../spec/): what the product does. [v1](../spec/v1.md), [config](../spec/config.md), [UI](../spec/ui.md).
- [`docs/toolchain.md`](../toolchain.md): each tool and library, why it was chosen, and what was rejected.
- [`docs/adr/`](../adr/): decisions that are hard to reverse.
