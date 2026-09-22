# Changelog

All notable changes to **Turbo Fortran (`tf`) / Turbo F77 (`tf77`)** will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.90.1] - 2026-09-22

### Fixed
- **Default File Extension in `tf`**: Set default extension to `.f90` when saving and compiling untitled buffers in `bin/tf` (Modern Fortran mode).
- **Free Format Verification**: Improved syntax highlighting and format checks for Modern Fortran (F90+) free-form source files.

## [0.90] - 2026-09-21

### Added
- **Submenu Mnemonic Hotkeys**: Direct single-letter execution in drop-down menus (e.g. `File` ➔ `N` New, `O` Open, `S` Save, `A` Save As).
- **Word-by-Word Navigation**: Move cursor word-by-word with `Ctrl + Left/Right` (and macOS `Option + Left/Right`).
- **macOS Option Meta Key Guide**: Added documentation and tips for configuring Option as Meta key in macOS Terminal.app and iTerm2.
- **MIT License**: Added official open-source MIT License.
- **Language-Separated Documentation**: Split `README.md` into dedicated English (`README.md`) and Korean (`README.ko.md`) with reciprocal navigation links.
- **Authentic Terminal Screenshots**: Embedded high-resolution native terminal screenshots in documentation.
- **Windows Defender Notice**: Added guidance for Windows SmartScreen false-positive warnings on unsigned binaries.

## [0.89] - 2026-09-18

### Added
- **Automated Multi-Platform Release CI/CD**: GitHub Actions workflow and `scripts/build_release.sh` building pure static binaries for both `tf` and `tf77` across macOS (Apple Silicon), Linux (amd64), and Windows (amd64) with SHA256 checksums.
- **Mouse Support**: Mouse click, drag block selection, and mouse wheel scrolling across editor, menu bar, and dialogs.
- **Unsaved Changes Confirmation**: Borland-style modal alert dialog prompting to save or discard changes when opening files or exiting.
- **User Screen Live Streaming**: Stream live subprocess output to the `Alt+F5` User Screen buffer during debugger sessions.

### Fixed
- **Debugger Stability**: Synchronized GDB with `target-async off`, isolated multi-file breakpoints, and added an advisory note for macOS Apple LLDB variable inspection limitations.

## [0.88] - 2026-09-14

### Added
- **Dual Binary Architecture**:
  - `bin/tf`: Default Modern Fortran (F90+) free-form editor with modern keyword/operator highlighting and column guides disabled.
  - `bin/tf77`: Default Classic FORTRAN 77 fixed-form editor with column 6 continuation and column 72 boundary guides (`│`).
  - Automatic format mode switching based on file extension (`.for`, `.f`, `.f77` vs `.f90`, `.f95`, etc.).
- **Multi-Backend Interactive Debugger**:
  - **Internal Engine (Pure Go)**: Built-in interpreter for instant single-file stepping and 100% Watches variable inspection.
  - **Native Engine (LLDB/GDB)**: Automatic subprocess backend for multi-file companion modules (`math_sub.for`, `io_sub.for`).
  - Engine switcher in `Debug ➔ Engine`.
- **Compiler Toolchain Integration**: Priority 1 for `gfortran` with modular `-J <tmpDir>` isolation; automatic Open Watcom 2.0 (`wfl386`) support when `$WATCOM` is set.
- **Classic Borland Turbo Vision UI**: Signature Turbo Blue canvas (`#0000A8`), double-line box borders (`╔═╗`), 3D text drop shadows, top pull-down menu bar, bottom hotkey bar, and `Alt+F5` User Screen.
- **Retro PC Speaker Sound Effects**: Dual-tone compilation success chime, failure buzz, and debugger step pings with audio toggle (`Options ➔ Sound`).
- **Atomic File Operations**: Safe atomic saves via swap files, UTF-8 BOM auto-stripping, and non-text binary file protection.
- **Cross-Platform Clipboard**: Seamless integration with system clipboard (`Ctrl+C`, `Ctrl+X`, `Ctrl+V`, `Ctrl+A`).
