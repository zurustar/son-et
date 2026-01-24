# Glossary

This glossary defines common terms used across all son-et specification documents.

## Core Concepts

- **Interpreter**: The son-et system that parses and executes FILLY scripts through a virtual machine
- **FILLY_Script**: Legacy scripting language with C-like syntax used for creating interactive multimedia applications
- **TFY_Script**: A FILLY language script file with .tfy or .fil extension
- **OpCode**: An operation code structure containing a command name and arguments for VM execution
- **VM**: Virtual Machine - the execution engine that interprets OpCode sequences
- **Sequencer**: The VM component that manages OpCode execution and timing

## Execution Modes

- **Direct_Mode**: Execution mode where son-et interprets TFY files from a directory at runtime for development
- **Embedded_Mode**: Build mode where a specific project is embedded into the son-et executable at build time for distribution

## Graphics and Display

- **Virtual_Desktop**: A fixed 1280x720 rendering canvas that hosts multiple virtual windows
- **Virtual_Window**: A sub-region within the Virtual_Desktop that displays game content (typically 640x480)
- **Picture**: An image buffer that can be loaded, manipulated, and displayed
- **Cast**: A sprite object with transparency support and z-ordering

## Timing and Synchronization

- **MIDI_Sync_Mode**: Timing mode where script execution is synchronized to MIDI playback ticks
- **Time_Mode**: Timing mode where script execution is driven by the game loop at 60 FPS
- **Step**: A timing unit whose meaning depends on the current timing mode
- **mes_Block**: A message-driven execution block that responds to events

## Threading and Concurrency

- **Script_Goroutine**: A separate thread that executes the interpreted user script
- **Main_Thread**: The primary thread running the game loop (Update/Draw at 60 FPS)
- **Render_Mutex**: A synchronization lock protecting shared rendering state

## Audio

- **SoundFont**: A .sf2 file containing instrument samples for MIDI synthesis
