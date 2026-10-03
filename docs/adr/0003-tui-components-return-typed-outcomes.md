# TUI components return typed outcomes in the step

A TUI component reports what happened, such as a save request, a revealed Snippet, or a confirmed quit, by returning an outcome in the same `overlay.Step` that carries its next state. The overlay stack hands the outcome to the nearest parent, or out to the model, before the next message is read. Outcomes are members of the sealed union `outcome.Outcome`. The overlay stack is generic over the outcome type, so no outcome in its API is typed `any`.

## Considered Options

- **Outcomes as `tea.Cmd` messages.** The component returns a command that yields the outcome, and the model handles it on a later `Update`. Bubble Tea reads other messages in between, so a key press can land after the component changed state but before the model acts on it. That gap caused [#71](https://github.com/DannyFestor/TuiSnip/issues/71): a second `ctrl+s` arrived while a save was pending and created a duplicate Snippet. Returning the outcome in the step closes the gap. The component's new state and its outcome are applied in the same round.
- **Outcomes as `any`.** This was the first version of the overlay stack ([#72](https://github.com/DannyFestor/TuiSnip/issues/72)). The model type-switches on the outcomes it knows and silently drops the rest, so an outcome nobody handles compiles, passes lint, and does nothing when pressed.

## Consequences

- `gochecksumtype` fails every type switch over `outcome.Outcome` that misses a member, unless the switch has a `default:` case. `Model.concluded` has none, so a new outcome fails `make lint` until the model handles it ([linting](../standards/linting.md#type-switches-over-a-sealed-union-are-exhaustive)).
- Outcome types live in their own package, `tui/outcome`, so components can move into subpackages of `tui` without importing `tui`.
- Go can't infer the outcome type from a component passed as an `Overlay[O]`, so `tui` wraps `overlay.Stay` and `overlay.Close` in local `stay` and `closing` helpers.
