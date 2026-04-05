## ADDED Requirements

### Requirement: Initial tool surface
The first Go rewrite slice SHALL expose `Bash`, `Read`, and `Edit` as the supported local tools for headless execution.

#### Scenario: Runtime registers the initial tools
- **WHEN** the headless runtime prepares the tool catalog for a session
- **THEN** `Bash`, `Read`, and `Edit` are available to the model
- **AND** unsupported later-phase tools are not advertised as if they were implemented

### Requirement: Read-before-edit safety
The Go `Edit` implementation SHALL enforce the same read-before-write safety model expected by the rewrite specs for existing files.

#### Scenario: Edit is attempted without a prior read
- **WHEN** the assistant requests an edit against an existing file that has not been read in the current tool context
- **THEN** the edit is rejected
- **AND** the tool result explains that the file must be read before it can be edited

### Requirement: Permission-gated shell execution
The Go `Bash` implementation SHALL preserve permission-gated execution semantics so that write-capable or otherwise unsafe shell commands are not silently executed.

#### Scenario: Bash command requires approval
- **WHEN** the assistant requests a shell command that is not classified as a safe read-only action under the active permission mode
- **THEN** the runtime refuses or pauses execution according to the configured permission policy
- **AND** the command is not run as an implicit allow

### Requirement: Shared tool runtime contract
The initial tools SHALL execute through a common Go tool interface that can be reused by later tools and future interactive entrypoints.

#### Scenario: Query loop invokes a tool
- **WHEN** the headless query loop dispatches a supported tool call
- **THEN** the tool is executed through the shared runtime contract rather than bespoke per-call logic in the CLI layer
- **AND** the returned result can be fed back into the ongoing session without tool-specific special casing in `cmd/holy`
