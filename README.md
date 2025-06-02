# nevr-common
the runtime framework for the NEVR game service

This codebase defines the runtime API and protocol interface used by Nakama EVR.

It is structured similarly to the `nakama-common` repository.

The code is broken up into packages for different parts:

- `gameapi`: The game-engine API definitions, such as control messages, session data and user bones.
- `rtapi`: The runtime API definitions, including the frame structure and connectivity statistics.
