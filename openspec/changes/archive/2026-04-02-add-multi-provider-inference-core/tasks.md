## 1. Define The Shared Inference Contract

- [x] 1.1 Define the provider-neutral request, event, tool-call, tool-result, usage, and stop-reason types that the runtime will consume.
- [x] 1.2 Define which provider-specific metadata is preserved on the shared event stream without becoming part of the core query control flow.
- [x] 1.3 Document the adapter responsibilities and package boundaries between core runtime logic and provider-specific transport logic.

## 2. Define Capability And Model Descriptors

- [x] 2.1 Specify the provider and model capability matrix for features such as tool calling, tool-argument streaming, structured output, reasoning visibility, and stateful response handling.
- [x] 2.2 Define the model descriptor format so provider-scoped model IDs and endpoint details stay out of core runtime constants.
- [x] 2.3 Specify runtime behavior when required capabilities are absent, including explicit failure versus supported degradation paths.

## 3. Specify Provider Adapter Families

- [x] 3.1 Specify the Anthropic adapter boundary and how Anthropic-native request and stream semantics map into the shared turn contract.
- [x] 3.2 Specify the OpenAI Responses adapter boundary and how response items and streaming events map into the shared turn contract.
- [x] 3.3 Specify the OpenAI-compatible adapter boundary, including the minimum supported contract and explicit non-guarantees relative to full Responses semantics.

## 4. Rebase The Runtime Plan

- [x] 4.1 Revise the headless session runner requirements so provider selection and streaming depend on the shared inference contract rather than Anthropic-specific transport assumptions.
- [x] 4.2 Revise the conformance baseline so it validates shared event semantics and adapter conformance, not just a single backend path.
- [x] 4.3 Update the bootstrap runtime plan to make multi-provider inference core a prerequisite for the first executable headless slice.
