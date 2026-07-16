# Engine Kernelization Memo

This document records the current architectural judgment about future Engine kernelization work.

It is not a commitment to start implementation now.
Its purpose is to preserve the design direction and boundary decisions so they are not lost before the Engine feature set is mature enough.

## 1. Current conclusion

GameAgentEngine is suitable for future kernelization work.

The correct target is not:

- compile the current HTTP service directly into a DLL / so / dylib and treat that as the final architecture

The correct target is:

- refactor the Engine into a host-agnostic runtime core
- let the host process own communication, persistence, and LLM invocation paths
- keep Engine focused on semantic runtime, orchestration, continuity, and execution rules

In short, the goal is to turn Engine from a service-owned runtime into an embeddable inference kernel.

## 2. Why this is feasible in the current repository

The current repository already has meaningful layering.

- `internal/engine`: inference pipeline, context build, orchestration, world tick continuity, action and memory flow
- `internal/service`: world management, transactional workflows, persistence-side orchestration
- `internal/store`: database access and persistence implementation
- `internal/api`: HTTP surface
- `internal/llm`: provider-side LLM access
- `internal/external`: external dispatch adapters
- `cmd/gameagentengine`: process bootstrap and server assembly

That means future kernelization is primarily a boundary and ownership refactor, not a rewrite from zero.

## 3. Primary motivations

The intended value of kernelization is broader than replacing HTTP with a local dynamic library call.

### 3.1 Communication goes from service IO to in-process ABI

The expected direction is:

- replace inter-process JSON request/response exchange with in-process binary message exchange
- bypass HTTP routing, middleware, and JSON encode/decode overhead on hot paths
- reduce string-heavy payload assembly and repeated object reconstruction

Important note:

The gain does not come only from changing JSON to binary.
The gain depends on also making the host-kernel boundary coarse-grained enough.

If the future boundary still exposes many tiny CRUD-style calls, the system will only replace many small JSON round-trips with many small ABI round-trips.
That would dilute most of the benefit.

### 3.2 Persistence goes from Engine-owned CRUD to host-owned state bridging

The intended direction is:

- Engine should stop assuming ownership of database access
- Engine should declare what state it needs and what state changes it produces
- the host process should decide where data comes from, how it is cached, and how it is persisted

This allows the host to bridge state using its own storage model, such as:

- in-memory runtime state
- SQLite
- custom save systems
- ECS-backed state
- platform-specific authority services

Important note:

The future design should not turn into a high-frequency demand-driven state RPC system between host and kernel.
The preferred main path is:

- snapshot-in
- patch-out

That means the host feeds a sufficiently complete state slice or runtime snapshot into the kernel, and the kernel returns patches, effects, plans, memory updates, and pending continuation state.

Demand-driven bridge calls should remain supplemental, not the dominant execution path.

### 3.3 LLM invocation goes from Engine-owned provider flow to host-owned inference pipeline

The intended direction is:

- Engine prepares prompts, context, contracts, and expected structured output
- the host process performs the real model invocation
- the host uses its own network path, inference middleware, connection pools, internal routing, dedicated links, and fallback strategy

This is especially important for environments where the host already has a private inference pipeline or an internal high-speed service route.

The correct abstraction is not "Engine owns the model call".
The correct abstraction is closer to:

- Engine produces an inference specification
- host executes it
- host returns a structured result envelope

This can reduce deployment coupling and often improves real latency more than localizing the HTTP service layer alone.

## 4. Target ownership model

### 4.1 Current tendency

Today the system tends toward the following ownership model:

- Engine owns API exposure
- Engine owns persistence path
- Engine owns LLM provider path
- external systems integrate around the Engine service

### 4.2 Future kernelized tendency

The future kernelized model should invert that ownership:

- host owns IO
- host owns persistence implementation
- host owns LLM pipeline access
- Engine owns semantic runtime rules and execution logic only

This inversion is the real point of kernelization.
The dynamic library form is only one packaging option.

## 5. What should remain inside the kernel

The future kernel should preserve the parts of Engine that are its real product value.

- world semantic model
- context building and reduction rules
- relation and memory selection logic
- pipeline modes and multi-round orchestration
- structured output parsing and validation
- action planning and execution rules
- world tick continuity and narrative state transitions
- pending continuation and resume state machines
- embedding-safe interaction contracts

## 6. What should become host bridges

The following capabilities should move toward explicit host-provided bridge interfaces.

- LLM invocation
- state read and write
- external action execution
- runtime task dispatch
- scheduler and clock access
- logging, metrics, and trace sinks
- lock and transaction policy

These should not remain hidden inside a server-only runtime assumption.

## 7. What the host-kernel boundary should prefer

The future boundary should prefer coarse-grained semantic operations rather than tiny persistence-style calls.

Examples of the correct shape:

- `invoke_dialogue`
- `advance_world_tick`
- `apply_world_snapshot`
- `run_inference_round`
- `resume_pending_effect`
- `commit_state_patch`

Examples of the wrong default shape:

- many per-field updates
- many per-component fetches
- relation-by-relation remote access
- memory-by-memory host callbacks during normal execution

The kernel should consume structured runtime state and return structured runtime results.

## 8. Recommended message model direction

The future interface should move away from exposing current HTTP DTOs or database models as the stable kernel contract.

The stable direction should instead revolve around runtime messages such as:

- `WorldSnapshot`
- `RuntimeContext`
- `InferenceSpec`
- `InferenceResult`
- `StatePatch`
- `ActionEffect`
- `PendingContinuation`

These names are directional only, not final API commitments.

## 9. Major expected benefits

If implemented correctly, kernelization can provide the following benefits.

- lower communication overhead on hot paths
- better alignment with engine-side authority and save systems
- direct access to host-owned LLM infrastructure and private service links
- cleaner reuse across Unity, Unreal Engine, and Godot as different hosts over one runtime core
- lower deployment friction in embedded scenarios

## 10. Major risks and constraints

Kernelization is feasible, but it is not low-cost.

### 10.1 Interface design risk

The largest risk is designing the wrong host-kernel protocol.

If the future contract is too fine-grained, the project will preserve most of the current boundary cost in a different form.
If the contract simply mirrors current HTTP DTOs or GORM-facing models, the result will be an embedded service, not a true kernel.

### 10.2 Observability regression risk

Once LLM invocation and persistence ownership move to the host, the kernel will no longer automatically own the full trace of:

- request path
- retry behavior
- latency
- token usage
- provider fallback
- environment-specific failures

That means the host must return enough metadata with inference results and state application results for meaningful diagnosis.

### 10.3 Runtime model migration cost

Several current conveniences are built around the service model, including:

- automatic migrations
- callback persistence
- runtime task recovery
- centralized log sink
- world lock handling
- database retry policy

These do not disappear, but they stop being implicit Engine ownership and become explicit bridge policy.

### 10.4 Multi-engine packaging risk

Even after the runtime core is well designed, Unity, Unreal Engine, and Godot will still need separate host bindings and lifecycle integration.
That work should be treated as a later phase, not as the first definition of success.

## 11. Recommended sequencing

Kernelization should not start immediately.

The current correct sequencing is:

1. finish maturing Engine kernel features and contracts first
2. keep stabilizing interaction, intent, continuity, callback, and runtime task semantics
3. postpone kernelization implementation until those contracts are sufficiently mature
4. when implementation begins, validate the host-agnostic runtime boundary before doing full multi-engine binding work

The first real implementation milestone should be a local embedding proof of concept for the runtime boundary, not immediate Unity + UE + Godot production binding work.

## 12. Pre-implementation checkpoint

Before kernelization begins, the project should re-evaluate at least the following questions.

- Which latency is actually the primary bottleneck now?
- Which calls are truly high-frequency?
- Which state must remain host-authoritative?
- Which state should remain Engine-authoritative?
- Which flows must be synchronous and which can stay resumable or deferred?
- Which observability fields must the host return for inference and persistence results?

If these questions are still unclear, implementation should not start.

## 13. Final standing decision

The current standing decision is:

- keep this memo as design guidance only
- do not start kernelization implementation yet
- revisit this document after Engine functionality and contracts are mature enough

At that point, the project can turn this memo into a concrete architecture and execution plan.
