# Turbo Fortran / Turbo F77

> **Retro Borland Turbo Vision TUI IDE for Modern and Classic Fortran**

**Turbo Fortran** delivers two self-contained Go binaries — `tf` (Modern Fortran / Free-Form) and `tf77` (Classic FORTRAN 77 / Fixed-Form) — faithfully recreating the legendary look, feel, and ergonomics of Borland's **Turbo Pascal** and **Turbo C** from the golden DOS era.

Built with `tcell`, both binaries share the iconic classic Turbo Blue canvas (`#0000A8`), double-line box frames (`╔═╗`), top pull-down menus, bottom status bar, `Alt+F5` User Screen, PC speaker sound effects, an interactive multi-backend debugger with a real-time Watches window, and smart multi-file build support powered by **gfortran** (1st priority) with automatic **Open Watcom 2.0** support when `$WATCOM` is configured.

---

## Two Binaries, One Codebase

| Binary | Identity | Default File | Default Mode | Column Guides |
|:---|:---|:---|:---|:---|
| `bin/tf` | **Turbo Fortran** | `NONAME00.F90` | Free-Form (F90+) | Off |
| `bin/tf77` | **Turbo F77** | `NONAME00.FOR` | Fixed-Form (F77) | On (col 6, 72) |

Both binaries **auto-switch mode** based on the file extension whenever you open or save a file:
- **Fixed-Form** (`.for`, `.f`, `.f77`, `.ftn`, `.inc`): Columns 1-5 label, column 6 continuation, column 7+ statement, 72-column right guide active.
- **Free-Form** (`.f90`, `.f95`, `.f03`, `.f08`, `.f18`): No column restrictions, `!` comments at any position, column guides disabled, modern keyword and operator highlighting.

---

## Key Features

* **Authentic Borland Turbo Vision TUI**:
  * Classic Turbo Blue editor canvas (`#0000A8`) with double-line borders (`╔═╗`, `║ ║`, `╚═╝`)
  * 3D drop shadows and classic window headers (`[■] 1 NONAME00.F90 [▲]`)
  * Pull-down menu bar (`File`, `Edit`, `Search`, `Run`, `Compile`, `Debug`, `Options`, `Window`, `Help`)
  * Bottom hotkey bar with real-time column position indicator

* **Alt + Hotkey Direct Menu Navigation**:
  * Open any top menu instantly with `Alt+F` (File), `Alt+E` (Edit), `Alt+S` (Search), `Alt+R` (Run), `Alt+C` (Compile), `Alt+D` (Debug), `Alt+O` (Options), `Alt+W` (Window), `Alt+H` (Help)
  * macOS Option key translation support (`Option+F = ƒ`, `Option+E = ´`, etc.) and `Esc` prefix navigation

* **Dual-Mode Syntax Highlighting**:
  * **F77 Fixed-Form**: Column-aware tokenizer (label, continuation, statement, inline comment). Keywords: `PROGRAM`, `SUBROUTINE`, `FUNCTION`, `DO`, `GOTO`, `COMMON`, `EQUIVALENCE`, etc.
  * **F90+ Free-Form**: Full modern keyword set: `module`, `use`, `contains`, `type`, `allocatable`, `intent`, `recursive`, `pure`, `elemental`, `select`, `cycle`, `exit`, `where`, `forall`, `interface`, `pointer`, `target`, etc.
  * Modern operators: `==`, `/=`, `<=`, `>=`, `=>`, `::` highlighted in magenta
  * Modern array and string intrinsics: `sum`, `matmul`, `size`, `shape`, `trim`, `allocated`, `associated`, etc.

* **FORTRAN 77 Column Guides (Fixed-Form only)**:
  * Subtle vertical guide lines (`│`, `#2E60C8`) at **Column 6** (Continuation) and **Column 72** (Statement limit)
  * Intelligent rendering: only visible on blank cells, never obscures code
  * Toggle via `Options → Column Guides: ON/OFF`

* **Smart Multi-File Compilation & Auto-Dependency Resolution**:
  * Automatically detects and compiles all companion source files (`.for`, `.f`, `.f77`, `.f90`, `.f95`, etc.) in the same folder
  * Filters out duplicate `PROGRAM` entry points; collects only subroutines, functions, and modules
  * gfortran builds isolate `.mod` files in a temporary directory (`-J <tmpDir>`) to prevent workspace pollution

* **Multi-Backend Interactive Debugger & Watches Window**:
  * **Default Engine (Internal F77)**: Built-in pure-Go interpreter engine for single-file routines (`hello.for`, `fibonacci.for`, `stats.for`) — zero external dependencies, sub-millisecond step execution, and 100% real-time variable/array inspection in the Watches window.
  * **Native Engine (LLDB / GDB / Open Watcom)**: Subprocess backend for compiling and debugging native binaries. Automatically used for **multi-file projects** (e.g. `main.for` with companion files `io_sub.for`, `math_sub.for`), linking all routines and enabling seamless cross-file breakpoints, file auto-switching, and stepping.
  * **Engine Toggle**: Switch engines anytime via `Debug ➔ Engine: Internal (F77) / Native (LLDB/GDB)`.
  * **Note on Apple LLDB (macOS) Watch Inspection**: Apple LLDB on macOS does not ship with a Fortran TypeSystem plugin (`frame variable` cannot decode Fortran types). When debugging multi-file projects via LLDB on macOS, cross-file breakpoints and stepping work seamlessly, but the Watches window and status bar will display an advisory note: `[Note: macOS LLDB has limited Fortran variable inspection support]`. On Linux/Windows, GDB inspects Fortran variables natively. For full variable inspection on macOS, single-file code runs via the built-in **Internal** engine.
  * **F4**: Toggle breakpoint (full-width red bar `●`)
  * **F5**: Start debugging / Continue to next breakpoint
  * **F7**: Trace Into, **F8**: Step Over, **Ctrl+F2**: Reset
  * Current execution pointer highlighted with a full-width yellow bar (`►`)
  * Bottom **Watches Window**: monitors variable names, types, and values in real time

* **Alt+F5 User Screen**:
  * Seamlessly toggle to a fullscreen console viewer to inspect program standard output and exit codes
  * Press any key to return to the IDE

* **Compiler Toolchain Integration**:
  * **Priority 1**: `gfortran` (Homebrew, system), `flang-new`, `flang`, `ifx`, `ifort`
  * **Priority 2** (when `$WATCOM` is set): `wfl386`, `wfc386`, `wfl`
  * Borland modal "Compiling..." statistics box (file, compiler, lines, errors, warnings, elapsed time)
  * Diagnostic parser supporting Open Watcom and standard Unix/GCC formats

* **Retro PC Speaker Sound Effects**:
  * Two-tone chime on success, buzz on errors, ping on debugger breakpoints
  * Toggle via `Options → Sound: ON/OFF`

---

## Keyboard Shortcuts

| Shortcut | Action | Description |
|:---|:---|:---|
| **Alt + F** | File Menu | Open File menu (`New`, `Open`, `Save`, `Exit`) |
| **Alt + E** | Edit Menu | Open Edit menu (`Undo`, `Cut`, `Copy`, `Paste`) |
| **Alt + S** | Search Menu | Open Search menu (`Find`, `Again`, `Go to Line`) |
| **Alt + R** | Run Menu | Open Run menu (`Run`, `User Screen`) |
| **Alt + C** | Compile Menu | Open Compile menu (`Compile`, `Make`) |
| **Alt + D** | Debug Menu | Open Debug menu (`Cont`, `Step`, `Trace`, `BP`) |
| **Alt + O** | Options Menu | Open Options menu (`Line Nums`, `Guides`, `Sound`) |
| **Alt + W** | Window Menu | Open Window menu (`Tile`, `Cascade`, `Close`) |
| **Alt + H** | Help Menu | Open Help menu (`About`) |
| **F1** | Help / About | Display IDE information and help |
| **F2** | Save | Save current buffer (or Save As if new) |
| **F3** | Open | Open file selection dialog |
| **F4** | **Breakpoint** | Toggle breakpoint on current line (`●`) |
| **F5** | **Debug / Continue** | Start debugging or continue execution |
| **F7** | **Trace Into** | Single step into subroutine/function |
| **F8** | **Step Over** | Step over current statement |
| **Ctrl + F2** | **Reset Debugger** | Reset and terminate active debug session |
| **Ctrl + F9** | **Run** | Build, execute, and view output in User Screen |
| **Alt + F9** | **Compile** | Compile target with "Compiling..." dialog |
| **F9** | Make | Build with companion modules |
| **Alt + F5** | **User Screen** | Toggle full-screen program output |
| **Ctrl + F** | **Find** | Search text in active buffer |
| **Ctrl + L** | **Search Again** | Repeat search for next match |
| **Alt + G** | **Go to Line** | Jump to line number |
| **Alt + L** | **Line Numbers** | Toggle line numbers |
| **F10** | Menu Bar | Focus top pull-down menu bar |
| **Alt + X** | Exit | Exit IDE |
| **Shift + Arrows** | **Select Block** | Select text block |
| **Ctrl + C** / **Ctrl + Ins** | **Copy** | Copy selected text to clipboard |
| **Ctrl + X** / **Shift + Del** | **Cut** | Cut selected text to clipboard |
| **Ctrl + V** / **Shift + Ins** | **Paste** | Paste clipboard text |
| **Esc** | Close / Cancel | Dismiss modal dialog / clear selection |

---

## Build and Installation

### 1. Pre-built Binaries (GitHub Releases)

Download ready-to-use standalone executables (`tf` and `tf77`) for your platform from [GitHub Releases](https://github.com/MaroShim/TurboF77/releases):
* **macOS**: `tf77-v0.90-darwin-arm64.tar.gz` (Apple Silicon M-series)
* **Linux**: `tf77-v0.90-linux-amd64.tar.gz` (64-bit)
* **Windows**: `tf77-v0.90-windows-amd64.zip` (64-bit)

> [!NOTE]
> **Windows Defender / SmartScreen Notice**:
> Since these open-source binaries are newly compiled without expensive commercial code-signing certificates, Windows Defender or SmartScreen may occasionally flag them as unrecognized or a false positive.
> If a Windows SmartScreen popup appears, click **"More info" ➔ "Run anyway"** (추가 정보 ➔ 실행) or add an exclusion to run safely. You can also build directly from source using the Go compiler below.

### 2. Prerequisites
* Go 1.20 or newer
* (Optional) `gfortran` (recommended), `flang`, or **Open Watcom 2.0** for compilation
* (Optional) `lldb` (macOS) or `gdb` (Linux) for native source-level debugging

### 3. Installation via Go (Recommended)

```bash
# Install Turbo F77 (Classic F77)
go install github.com/MaroShim/tf77/cmd/tf77@latest

# Install Turbo Fortran (Modern F90+)
go install github.com/MaroShim/tf77/cmd/tf@latest
```

Ensure `$GOPATH/bin` (or `~/go/bin`) is in your `$PATH`. You can then launch `tf77` or `tf` from anywhere:

```bash
tf77
# or
tf
```

### 4. Build from Source


```bash
# Build both binaries
go build -o bin/tf   ./cmd/tf
go build -o bin/tf77 ./cmd/tf77

# Launch Turbo Fortran (F90+ default mode)
./bin/tf

# Launch Turbo F77 (F77 fixed-form default mode)
./bin/tf77

# Open a specific file (auto-detects mode from extension)
./bin/tf  examples/modern_fibonacci.f90
./bin/tf77 examples/fibonacci.for
```

---

## Compiler Setup

### gfortran (Recommended — macOS / Linux)

```bash
# macOS (Homebrew)
brew install gfortran

# Ubuntu / Debian
sudo apt install gfortran
```

`tf` and `tf77` auto-detect `gfortran` and use it with `-g` debug symbols. On macOS, `lldb` is used as the native debugger automatically.

### Open Watcom 2.0 (Classic F77, when `$WATCOM` is set)

```bash
export WATCOM=/opt/watcom
export PATH=$WATCOM/binl64:$PATH
```

If `$WATCOM` is set, Open Watcom takes priority over other compilers. Without it, `gfortran` is always tried first.

---

## Example Projects

**Classic Fortran 77 (Fixed-Form)**
* `main.for`: Default starter template
* `examples/hello.for`: Classic Hello World
* `examples/fibonacci.for`: Fibonacci sequence generator
* `examples/stats.for`: Array statistics and RMS calculation
* `examples/modular/`: Multi-file modular program
  * `main.for`: Main entry point (`PROGRAM MODULAR`)
  * `math_sub.for`: Arithmetic function (`INTEGER FUNCTION CALCSUM`)
  * `io_sub.for`: Formatted output subroutine (`SUBROUTINE PRINTSUM`)

**Modern Fortran (Free-Form, F90+)**
* `examples/modern_fibonacci.f90`: Fibonacci using `ALLOCATABLE` array and modern `DO` loops

---
---

# Turbo Fortran / Turbo F77 — 한국어 안내

> **현대 및 클래식 포트란을 위한 볼랜드 터보 비전 레트로 TUI IDE**

**Turbo Fortran**은 두 개의 독립 Go 바이너리를 제공합니다 — `tf` (모던 포트란 / 자유 형식, F90+)와 `tf77` (클래식 FORTRAN 77 / 고정 형식). 90년대 볼랜드(Borland)의 전설적인 **Turbo Pascal**과 **Turbo C** 특유의 터보 비전 UI(파란색 에디터 캔버스 `#0000A8`, 이중선 프레임 `╔═╗`, 풀다운 메뉴바, `Alt+F5` User Screen, 사운드 FX)에 **gfortran** 기반 컴파일러 및 `lldb`/`gdb` 네이티브 디버거를 결합한 레트로 터미널 개발 환경(TUI IDE)입니다.

---

## 두 개의 바이너리, 하나의 코드베이스

| 바이너리 | 브랜드 | 기본 파일 | 기본 모드 | 구분선 |
|:---|:---|:---|:---|:---|
| `bin/tf` | **Turbo Fortran** | `NONAME00.F90` | 자유 형식(F90+) | 비활성 |
| `bin/tf77` | **Turbo F77** | `NONAME00.FOR` | 고정 형식(F77) | 활성 (6열, 72열) |

두 바이너리 모두 파일을 열거나 저장할 때 **확장자에 따라 자동으로 모드를 전환**합니다:
- **고정 형식** (`.for`, `.f`, `.f77`, `.ftn`, `.inc`): 1~5열 레이블, 6열 연속행, 7열+ 문장, 72열 우측 구분선 활성화
- **자유 형식** (`.f90`, `.f95`, `.f03`, `.f08`, `.f18`): 열 제약 없음, 임의 위치 `!` 주석, 구분선 비활성, 모던 키워드 및 연산자 강조

---

## 주요 특징

* **Classic Borland Turbo Vision UI**:
  * 시그니처 터보 블루 에디터 캔버스 (`#0000A8`) 및 이중선 박스 드로잉 (`╔═╗`, `║ ║`, `╚═╝`)
  * 입체 텍스트 그림자(Drop Shadow) 및 레트로 윈도우 헤더 (`[■] 1 NONAME00.F90 [▲]`)
  * 상단 풀다운 메뉴바 및 하단 핫키 바 완전 구현

* **Alt + 첫 글자 메뉴 즉시 호출**:
  * `Alt+F`(File), `Alt+E`(Edit), `Alt+S`(Search), `Alt+R`(Run), `Alt+C`(Compile), `Alt+D`(Debug), `Alt+O`(Options), `Alt+W`(Window), `Alt+H`(Help)
  * macOS Option 키 유니코드 매핑(`Option+F = ƒ`, `Option+E = ´` 등) 및 `Esc` 프리픽스 완벽 지원

* **이중 모드 구문 강조(Dual-Mode Syntax Highlighting)**:
  * **F77 고정 형식**: 열 기반 토크나이저 (레이블, 연속행, 문장, 인라인 주석 각각 별도 강조)
  * **F90+ 자유 형식**: 모던 키워드 전체 지원 (`module`, `use`, `contains`, `type`, `allocatable`, `intent`, `recursive`, `pure`, `elemental`, `select`, `cycle`, `exit`, `where`, `forall`, `interface`, `pointer`, `target` 등)
  * 모던 연산자 강조: `==`, `/=`, `<=`, `>=`, `=>`, `::`
  * F90+ 배열/문자열 내장함수 강조: `sum`, `matmul`, `size`, `shape`, `trim`, `allocated`, `associated` 등

* **FORTRAN 77 세로 구분선 (고정 형식 전용)**:
  * **6열** (연속행 경계) 및 **72열** (문장 끝 한계)에 소프트 블루(`│`) 세로 가이드선
  * 코드가 없는 빈 공백 셀에만 표시되며 사용자의 소스코드를 가리지 않음
  * `Options → Column Guides: ON/OFF` 토글 지원

* **스마트 다중 소스 파일 자동 탐색 및 빌드**:
  * 메인 파일이 있는 폴더에서 보조 모듈(`.for`, `.f`, `.f77`, `.f90`, `.f95` 등) 파일을 자동 감지하여 일괄 빌드
  * gfortran 빌드 시 `.mod` 파일을 임시 디렉터리(`-J <tmpDir>`)에 격리하여 작업 폴더 오염 방지

* **다중 백엔드 인터랙티브 디버거 & Watches Window**:
  * **기본 엔진 (Internal F77)**: 단일 파일 실습 및 알고리즘용 내장 순수 Go 인터프리터 엔진 (`hello.for`, `fibonacci.for`, `stats.for`) — 외부 툴 불필요, 0.001초 미만의 초고속 스텝 실행, 하단 Watches 창에서 포트란 변수명, 타입, 현재 값을 **100% 완벽하게 실시간 추적**.
  * **네이티브 엔진 (LLDB / GDB / Open Watcom)**: 실제 컴파일된 바이너리를 구동하는 서브프로세스 백엔드. 서브루틴/함수 파일이 분리된 **다중 파일 프로젝트**(`main.for` + `io_sub.for`, `math_sub.for`) 디버깅 시 자동 적용되어 전체 파일을 일괄 링크하고 파일 간 브레이크포인트, 에디터 파일 자동 전환 및 스텝 인투/오버 완벽 지원.
  * **디버거 엔진 즉시 전환**: 상단 메뉴 `Debug ➔ Engine: Internal (F77) / Native (LLDB/GDB)` 항목을 통해 언제든 자유롭게 전환 가능.
  * **macOS Apple LLDB 변수 감시 안내**: macOS의 기본 Apple LLDB에는 Fortran TypeSystem 플러그인이 포함되어 있지 않아 LLDB 자체적으로 포트란 로컬 변수 타입을 디코딩하지 못합니다 (`frame variable` 지원 한계). macOS에서 다중 파일 디버깅 시 브레이크포인트 및 스텝 이동은 정상 작동하나, Watches 창 및 상태 표시줄에 안내 메시지(`[Note: macOS LLDB has limited Fortran variable inspection support]`)가 표시됩니다. (Linux/Windows 환경의 GDB는 네이티브 포트란 변수 추적을 완벽 지원합니다.) 단일 파일의 완벽한 변수 감시가 필요할 경우 내장 **Internal** 엔진으로 실행됩니다.
  * **F4**: 브레이크포인트 설정/해제(`●`), **F5**: 디버깅 시작/계속, **F7**: Trace Into, **F8**: Step Over, **Ctrl+F2**: 리셋
  * 실행 중 현재 라인: 노란색 바(`►`) 강조
  * 하단 **Watches 윈도우**: 변수명, 타입, 값 실시간 감시

* **Alt+F5 User Screen**:
  * 컴파일된 포트란 프로그램의 실행 결과를 전체 화면 콘솔 뷰어로 전환하여 확인
  * 아무 키나 누르면 즉시 IDE로 복귀

* **컴파일러 툴체인 연동**:
  * **1순위**: `gfortran` (Homebrew, 시스템), `flang-new`, `flang`, `ifx`, `ifort`
  * **2순위** (`$WATCOM` 설정 시): `wfl386`, `wfc386`, `wfl`
  * 볼랜드 특유의 "Compiling..." 팝업 박스 (대상 파일, 컴파일러, 라인 수, 에러/워닝, 경과 시간)
  * Open Watcom 및 표준 Unix/GCC 진단 메시지 파싱 및 원클릭 에러 이동

* **볼랜드 레트로 사운드 이펙트**:
  * 컴파일 성공 시 2단 비프음, 실패 시 에러 버즈음, 디버거 스텝 시 핑음
  * `Options → Sound: ON/OFF` 토글 지원

---

## 단축키 안내

| 단축키 | 기능 | 설명 |
|:---|:---|:---|
| **Alt + F** | **File Menu** | File 메뉴 열기 (`New`, `Open`, `Save`, `Exit`) |
| **Alt + E** | **Edit Menu** | Edit 메뉴 열기 (`Undo`, `Cut`, `Copy`, `Paste`) |
| **Alt + S** | **Search Menu** | Search 메뉴 열기 (`Find`, `Again`, `Go to Line`) |
| **Alt + R** | **Run Menu** | Run 메뉴 열기 (`Run`, `User Screen`) |
| **Alt + C** | **Compile Menu** | Compile 메뉴 열기 (`Compile`, `Make`) |
| **Alt + D** | **Debug Menu** | Debug 메뉴 열기 (`Cont`, `Step`, `Trace`, `BP`) |
| **Alt + O** | **Options Menu** | Options 메뉴 열기 (`Line Nums`, `Guides`, `Sound`) |
| **Alt + W** | **Window Menu** | Window 메뉴 열기 (`Tile`, `Cascade`, `Close`) |
| **Alt + H** | **Help Menu** | Help 메뉴 열기 |
| **F1** | Help / About | IDE 정보 및 도움말 |
| **F2** | Save | 현재 버퍼 저장 / 다른 이름으로 저장 |
| **F3** | Open | 파일 브라우저 열기 |
| **F4** | **Breakpoint** | 현재 라인에 브레이크포인트(`●`) 설정/해제 |
| **F5** | **Debug / Continue** | 디버깅 시작 / 다음 브레이크포인트까지 실행 |
| **F7** | **Trace Into** | 한 줄씩 실행 (서브루틴 내부 진입) |
| **F8** | **Step Over** | 한 줄씩 실행 (함수/루프 건너뜀) |
| **Ctrl + F2** | **Reset Debugger** | 디버깅 세션 리셋 및 종료 |
| **Ctrl + F9** | **Run** | 빌드, 실행 후 User Screen에 결과 표시 |
| **Alt + F9** | **Compile** | "Compiling..." 통계 모달과 함께 빌드 |
| **F9** | Make | 보조 모듈 파일들과 함께 전체 빌드 |
| **Alt + F5** | **User Screen** | 프로그램 실행 결과 전체화면 토글 |
| **Ctrl + F** | **Find** | 문자열 검색 다이얼로그 |
| **Ctrl + L** | **Search Again** | 다음 일치 항목 찾기 |
| **Alt + G** | **Go to Line** | 특정 라인 번호로 이동 |
| **Alt + L** | **Line Numbers** | 라인 번호 표시 on/off |
| **F10** | Menu Bar | 상단 풀다운 메뉴바 포커스 |
| **Alt + X** | Exit | IDE 종료 |
| **Shift + 방향키** | **Select Block** | 텍스트 블록 선택 |
| **Ctrl + C** / **Ctrl + Ins** | **Copy** | 클립보드로 복사 |
| **Ctrl + X** / **Shift + Del** | **Cut** | 클립보드로 잘라내기 |
| **Ctrl + V** / **Shift + Ins** | **Paste** | 커서 위치에 붙여넣기 |
| **Esc** | Close | 활성 모달 닫기 / 선택 영역 해제 |

---

## 설치 및 빌드 방법

### 1. 사전 빌드된 바이너리 다운로드 (GitHub Releases)

[GitHub Releases](https://github.com/MaroShim/TurboF77/releases)에서 OS별로 빌드된 독립 실행 파일(`tf`, `tf77` 포함)을 즉시 다운로드하여 사용할 수 있습니다:
* **macOS (Apple Silicon)**: `tf77-v0.90-darwin-arm64.tar.gz`
* **Linux (64-bit)**: `tf77-v0.90-linux-amd64.tar.gz`
* **Windows (64-bit)**: `tf77-v0.90-windows-amd64.zip`

> [!NOTE]
> **Windows Defender / SmartScreen 오진 안내**:
> 유료 상용 코드 서명(Code Signing) 인증서가 적용되지 않은 순수 오픈소스 바이너리 특성상, 윈도우 디펜더(Windows Defender) 또는 SmartScreen에서 바이러스/위험 파일로 오진(False Positive)할 수 있습니다.
> SmartScreen 경고 창이 나타날 경우 **"추가 정보" ➔ "실행"**을 누르시거나 백신 예외 처리를 하시면 안전하게 실행하실 수 있습니다. 오진이 염려되시는 경우 아래의 Go 소스 빌드 방식을 이용하시면 소스로부터 신뢰할 수 있는 바이너리를 직접 컴파일하여 사용하실 수 있습니다.

### 2. Go 명령어로 직접 설치 (권장)

소스 코드를 별도로 clone하지 않고 터미널에서 즉시 설치하여 사용할 수 있습니다:

```bash
# Turbo F77 설치 (클래식 F77 고정 형식 에디터)
go install github.com/MaroShim/tf77/cmd/tf77@latest

# Turbo Fortran 설치 (현대 F90+ 자유 형식 에디터)
go install github.com/MaroShim/tf77/cmd/tf@latest
```

`$GOPATH/bin` (또는 `~/go/bin`)이 `$PATH` 환경 변수에 등록되어 있다면 어디서든 실행할 수 있습니다:

```bash
tf77
# 또는
tf
```

### 3. 소스 코드에서 직접 빌드 및 실행


```bash
# 두 바이너리 모두 빌드
go build -o bin/tf   ./cmd/tf
go build -o bin/tf77 ./cmd/tf77

# Turbo Fortran 실행 (F90+ 자유 형식 기본 모드)
./bin/tf

# Turbo F77 실행 (F77 고정 형식 기본 모드)
./bin/tf77

# 특정 파일 열기 (확장자에 따라 모드 자동 전환)
./bin/tf  examples/modern_fibonacci.f90
./bin/tf77 examples/fibonacci.for
```

---

## 컴파일러 설정

### gfortran (권장 — macOS / Linux)

```bash
# macOS (Homebrew)
brew install gfortran

# Ubuntu / Debian
sudo apt install gfortran
```

`tf`와 `tf77` 모두 `gfortran`을 자동 탐색하여 `-g` 디버그 심볼 포함 빌드를 수행합니다. macOS에서는 `/usr/bin/lldb`가 자동으로 네이티브 디버거로 연동됩니다.

### Open Watcom 2.0 (클래식 F77 전용, `$WATCOM` 설정 시)

```bash
export WATCOM=/opt/watcom
export PATH=$WATCOM/binl64:$PATH
```

`$WATCOM`이 설정된 경우 Open Watcom이 최우선으로 사용됩니다. 설정하지 않으면 항상 `gfortran`이 먼저 탐색됩니다.

---

## 예제 코드

**클래식 포트란 77 (고정 형식)**
* `main.for`: 기본 시작 템플릿
* `examples/hello.for`: 기본 Hello World 포트란 77
* `examples/fibonacci.for`: 피보나치 수열 생성기
* `examples/stats.for`: 배열 통계 및 RMS 계산 예제
* `examples/modular/`: 다중 소스 파일 모듈화 예제
  * `main.for`: 메인 실행 진입점 (`PROGRAM MODULAR`)
  * `math_sub.for`: 연산 함수 모듈 (`INTEGER FUNCTION CALCSUM`)
  * `io_sub.for`: 콘솔 출력 서브루틴 모듈 (`SUBROUTINE PRINTSUM`)

**모던 포트란 (자유 형식, F90+)**
* `examples/modern_fibonacci.f90`: `ALLOCATABLE` 배열과 모던 `DO` 루프를 활용한 피보나치 수열
