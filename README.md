# Turbo F77 (tf77)

> **Retro Borland Turbo Vision TUI IDE for FORTRAN 77**

**Turbo F77 (`tf77`)** is a retro Terminal User Interface (TUI) integrated development environment for **FORTRAN 77**, faithfully recreating the legendary look, feel, and ergonomics of Borland's **Turbo Pascal** and **Turbo C** from the golden DOS era of the 1990s.

Built as a self-contained Go binary using `tcell`, Turbo F77 features the iconic classic Turbo Blue canvas (`#0000A8`), double-line box frames (`╔═╗`), top pull-down menus, bottom status bar, `Alt+F5` User Screen, PC speaker sound effects, an interactive step debugger with a real-time Watches window, fixed-format column guides, and smart multi-file build support powered by the **Open Watcom 2.0 FORTRAN 77** toolchain (with automatic `gfortran` fallback).

---

## Key Features

* **Authentic Borland Turbo Vision TUI**:
  * Classic Turbo Blue editor canvas (`#0000A8`) with double-line borders (`╔═╗`, `║ ║`, `╚═╝`)
  * 3D drop shadows and classic window headers (`[■] 1 NONAME00.FOR [▲]`)
  * Pull-down menu bar (`File`, `Edit`, `Search`, `Run`, `Compile`, `Debug`, `Options`, `Window`, `Help`)
  * Bottom hotkey bar with real-time column indicator (`Col: 1 [Label]`, `Col: 6 [Cont]`, `Col: 7 [Stmt]`, `Col: 73+ [Ident]`)

* **Alt + Hotkey Direct Menu Navigation**:
  * Open any top menu instantly with `Alt+F` (File), `Alt+E` (Edit), `Alt+S` (Search), `Alt+R` (Run), `Alt+C` (Compile), `Alt+D` (Debug), `Alt+O` (Options), `Alt+W` (Window), `Alt+H` (Help)
  * macOS Option key translation support (`Option+F = ƒ`, `Option+E = ´`, etc.) and `Esc` prefix navigation
  * Single-letter hotkey jump and item trigger while menus are active

* **FORTRAN 77 Column Guides (Subtle Vertical Lines)**:
  * Subtle vertical guide lines (`│`, `#2E60C8`) at **Column 6** (Continuation boundary) and **Column 72** (Statement limit)
  * Intelligent whitespace rendering: only displays on blank spaces, never obscuring your code
  * Easy toggle via `Options ➔ Column Guides: ON/OFF`

* **Smart Multi-File Compilation & Auto-Dependency Resolution**:
  * Automatically detects and compiles all companion `.for` / `.f` / `.f77` files (`SUBROUTINE`, `FUNCTION`, `BLOCK DATA`) in the same folder
  * Intelligently avoids duplicate main entry points by filtering out separate `PROGRAM ...` files
  * Editing a subroutine and pressing `F9 (Make)` automatically builds the main program together
  * "Compiling..." dialog reports total combined lines and linked files (`main.for (+2)`)
  * Multi-file error navigation: clicking an error in a companion file automatically opens that file and jumps to the exact line

* **Interactive FORTRAN 77 Debugger & Watches Window**:
  * **F4**: Toggle breakpoint (highlighted with a full-width red bar `●`)
  * **F5**: Start debugging / Continue to next breakpoint
  * **F7**: Trace Into (single-step execution into subroutines)
  * **F8**: Step Over (step across loops and routine calls)
  * Current execution pointer highlighted with a full-width yellow bar (`►`)
  * Bottom **Watches Window**: monitors variables, types, and values in real time
  * Audio ping feedback upon hitting breakpoints or stepping

* **Alt+F5 User Screen**:
  * Seamlessly toggle to a fullscreen DOS console viewer to inspect program standard output and exit codes
  * Press any key to return to the Turbo F77 IDE

* **Open Watcom 2.0 & GNU Fortran Toolchain Integration**:
  * Automatic path detection for `wfl386`, `wfc386`, `wfl`, `owcc` (`$WATCOM`, `PATH`, standard directories)
  * Fallback support for `gfortran` and `f77`
  * Borland modal "Compiling..." statistics box (Main file, Compiler, Total lines, Errors, Warnings, Elapsed time)
  * Open Watcom diagnostic parser with click-to-jump Error List window

* **Retro PC Speaker Sound Effects (Sound FX)**:
  * Two-tone chime on compile success, buzz on errors, and ping on debugger breakpoints
  * Toggle audio anytime via `Options ➔ Sound: ON/OFF`

---

## Keyboard Shortcuts

| Shortcut | Action | Description |
|---|---|---|
| **Alt + F** | **File Menu** | Open File menu (`New`, `Open`, `Save`, `Exit`) |
| **Alt + E** | **Edit Menu** | Open Edit menu (`Undo`, `Cut`, `Copy`, `Paste`) |
| **Alt + S** | **Search Menu** | Open Search menu (`Find`, `Again`, `Go to Line`) |
| **Alt + R** | **Run Menu** | Open Run menu (`Run`, `User Screen`) |
| **Alt + C** | **Compile Menu** | Open Compile menu (`Compile`, `Make`) |
| **Alt + D** | **Debug Menu** | Open Debug menu (`Cont`, `Step`, `Trace`, `BP`) |
| **Alt + O** | **Options Menu** | Open Options menu (`Line Nums`, `Guides`, `Sound`) |
| **Alt + W** | **Window Menu** | Open Window menu (`Tile`, `Cascade`, `Close`) |
| **Alt + H** | **Help Menu** | Open Help menu (`About`, `FORTRAN 77`) |
| **F1** | Help / About | Display Turbo F77 information and help |
| **F2** | Save | Save current buffer (or Save As if new) |
| **F3** | Open | Open file selection dialog |
| **F4** | **Breakpoint** | Toggle breakpoint on current line (`●`) |
| **F5** | **Debug / Continue** | Start debugging or continue execution |
| **F7** | **Trace Into** | Single step into routine |
| **F8** | **Step Over** | Step over statement / loop |
| **Ctrl + F2** | **Reset Debugger** | Reset and terminate active debug session |
| **Ctrl + F9** | **Run** | Build, execute, and view output in User Screen |
| **Alt + F9** | **Compile** | Compile target with "Compiling..." dialog |
| **F9** | Make | Build program with companion modules |
| **Alt + F5** | **User Screen** | Toggle full-screen program output screen |
| **Ctrl + F** | **Find** | Search text in active buffer |
| **Ctrl + L** | **Search Again** | Repeat search for next match |
| **Alt + G** | **Go to Line** | Jump to line number (`Ctrl+G`) |
| **Alt + L** | **Line Numbers** | Toggle line numbers (`F6`) |
| **F10** | Menu Bar | Focus top pull-down menu bar |
| **Alt + X** | Exit | Exit Turbo F77 IDE |
| **Shift + Arrows** | **Select Block** | Select text block |
| **Ctrl + C** / **Ctrl + Ins** | **Copy** | Copy selected text to clipboard |
| **Ctrl + X** / **Shift + Del** | **Cut** | Cut selected text to clipboard |
| **Ctrl + V** / **Shift + Ins** | **Paste** | Paste clipboard text |
| **Esc** | Close / Cancel | Dismiss modal dialog / clear selection |

---

## Build and Installation

### Prerequisites
* Go 1.20 or newer
* (Optional) **Open Watcom 2.0 FORTRAN 77** or **gfortran**

```bash
# 1. Build the binary
go build -o bin/tf77 ./cmd/tf77

# 2. Launch Turbo F77
./bin/tf77

# Or open a specific Fortran file
./bin/tf77 main.for
```

---

## Open Watcom 2.0 Setup

If Open Watcom 2.0 is installed, set the `WATCOM` environment variable so `tf77` can automatically detect `wfl386`:

```bash
export WATCOM=/opt/watcom
export PATH=$WATCOM/binl64:$PATH
```

If Open Watcom is not installed, `tf77` will automatically fall back to `gfortran` or `f77` when available.

---

## Example Projects

* `main.for`: Default starter template calculating cumulative sums
* `examples/hello.for`: Classic Hello World in FORTRAN 77
* `examples/fibonacci.for`: Fibonacci sequence generator
* `examples/stats.for`: Array statistics and RMS calculation
* `examples/modular/`: **Multi-file modular program**
  * `main.for`: Main program entry point (`PROGRAM MODULAR`)
  * `math_sub.for`: Arithmetic function (`INTEGER FUNCTION CALCSUM`)
  * `io_sub.for`: Formatted console output (`SUBROUTINE PRINTSUM`)

---
---

# Turbo F77 (tf77) - 한국어 안내

> **FORTRAN 77을 위한 볼랜드 터보 비전(Turbo Vision) 레트로 TUI IDE**

**Turbo F77 (`tf77`)**는 90년대 볼랜드(Borland)의 전설적인 **Turbo Pascal**과 **Turbo C** 특유의 비주얼 인터페이스(Turbo Vision 파란색 에디터 캔버스 `#0000A8`, 이중선 프레임 `╔═╗`, 상단 풀다운 메뉴바, 하단 핫키 바, `Alt+F5` User Screen, 사운드 FX)에 **Open Watcom 2.0 FORTRAN 77** 컴파일러 툴체인 및 인터랙티브 디버거를 결합한 레트로 터미널 개발 환경(TUI IDE)입니다.

단일 Go 바이너리로 빌드되어 macOS, Linux, Windows 어디서나 가볍고 네이티브하게 실행됩니다.

---

## 주요 특징

* **Classic Borland Turbo Vision UI**:
  * 시그니처 터보 블루 에디터 캔버스 (`#0000A8`) 및 이중선 박스 드로잉 (`╔═╗`, `║ ║`, `╚═╝`)
  * 입체 텍스트 그림자(Drop Shadow) 및 레트로 윈도우 헤더 (`[■] 1 NONAME00.FOR [▲]`)
  * 상단 풀다운 메뉴바 (`File`, `Edit`, `Search`, `Run`, `Compile`, `Debug`, `Options`, `Window`, `Help`)
  * 하단 핫키 바 (`F1 Help`, `F2 Save`, `F3 Open`, `Alt+F9 Compile`, `F9 Make`, `Ctrl+F9 Run`, `Alt+F5 User`, `F10 Menu`)

* **Alt + 첫 글자 메뉴 즉시 호출**:
  * `Alt+F`(File), `Alt+E`(Edit), `Alt+S`(Search), `Alt+R`(Run), `Alt+C`(Compile), `Alt+D`(Debug), `Alt+O`(Options), `Alt+W`(Window), `Alt+H`(Help) 단축키 지원
  * macOS Option 키 유니코드 매핑(`Option+F = ƒ`, `Option+E = ´` 등) 및 `Esc` 프리픽스 완벽 지원
  * 메뉴 활성화 상태에서 첫 글자 단축키로 즉시 메뉴 이동 및 항목 실행

* **FORTRAN 77 전용 세로 구분선 (Column Guides)**:
  * **6열 (연속행 경계)** 및 **72열 (문장 끝 한계)**에 은은한 소프트 블루(`│`, `#2E60C8`) 세로 가이드선 표시
  * 지능형 공백 렌더링: 코드가 없는 빈 공백 셀에만 표시되며 사용자의 소스코드를 절대 가리지 않음
  * 상단 메뉴 `Options ➔ Column Guides: ON/OFF` 토글 지원

* **스마트 다중 소스 파일 자동 탐색 및 빌드 (Multi-file Build)**:
  * 메인 파일이 있는 폴더에서 별도의 `PROGRAM` 정의가 없는 보조 모듈(`SUBROUTINE`, `FUNCTION`, `BLOCK DATA`) 파일들을 자동 감지하여 일괄 컴파일 & 링크
  * 같은 폴더에 독립 실행 프로그램이 여러 개 있어도 중복 main 충돌 없이 오직 보조 모듈만 선별하여 취합
  * 서브루틴 파일을 편집하다가 빌드해도 폴더 내 메인 프로그램과 함께 빌드되어 정상 실행 파일 생성
  * 컴파일 다이얼로그에 총 컴파일 라인 수 및 빌드된 보조 파일 개수(`main.for (+2)`) 표시
  * 보조 파일 에러 발생 시 에러 목록에서 클릭 시 해당 보조 파일로 자동 전환 및 해당 줄 즉시 이동

* **FORTRAN 77 고정 포맷(Fixed-Format) 완벽 지원 & 구문 강조**:
  * **1열 주석**: 1번째 열이 `C`, `c`, `*`, `!`인 경우 라인 전체 주석 처리
  * **1~5열 레이블**: 문장 레이블(Statement Label) 전용 스타일링
  * **6열 연속행 기호**: `Continuation` 문자 강조
  * **하단 상태바 실시간 가이드**: `Col: 1 [Label]`, `Col: 6 [Cont]`, `Col: 7 [Stmt]`, `Col: 73+ [Ident]` 안내
  * 대소문자 무구분 키워드, 내장 수학 함수(`SQRT`, `SIN`, `COS`, `ABS`, `MOD` 등), 타입 선언 및 논리 연산자 완벽 강조

* **인터랙티브 FORTRAN 77 디버거 & Watches Window**:
  * **F4** 브레이크포인트 설정/해제 ➔ **한 줄 전체 빨간색 바(`●`)** 강조
  * **F5** Start Debugging / Continue (다음 브레이크포인트까지 계속 실행)
  * **F7** Trace Into (한 줄씩 실행), **F8** Step Over (스텝 실행)
  * 실행 중 현재 멈춘 라인은 **한 줄 전체 노란색 바(`►`)**로 시선 집중
  * 하단 **Watches 윈도우**(`Debug` 메뉴): `I`, `SUM`, `F1`, `F2` 등 포트란 로컬 변수명, 타입, 값을 실시간 감시
  * 디버깅 스텝/브레이크포인트 적중 시 경쾌한 핑 오디오 사운드 피드백

* **Alt+F5 User Screen**:
  * 컴파일된 포트란 프로그램의 실행 결과를 전체 화면 DOS 콘솔 뷰어로 전환하여 확인
  * 아무 키나 누르면 즉시 Turbo F77 IDE로 복귀

* **Open Watcom 2.0 툴체인 연동 & Compiling 모달**:
  * `wfl386`, `wfc386`, `wfl`, `owcc` 자동 경로 탐색 (`WATCOM` 환경변수, 시스템 PATH, 표준 설치 디렉토리)
  * 컴파일 시 볼랜드 특유의 **"Compiling..." 팝업 박스**(대상 파일, 총 라인 수, 에러/워닝 수, 경과 시간)
  * Open Watcom 전용 진단 메시지 정밀 파싱 및 원클릭 라인 점프 에러 목록 다이얼로그 지원

* **볼랜드 레트로 사운드 이펙트 (Sound FX)**:
  * 컴파일 성공 시 2단 비프음, 빌드 실패 시 에러 버즈음, 디버거 핑음
  * `F10` ➔ `Options` ➔ `Sound: ON / OFF` 토글 지원

---

## 단축키 안내

| 단축키 | 기능 | 설명 |
|---|---|---|
| **Alt + F** | **File Menu** | File 메뉴 열기 (`New`, `Open`, `Save`, `Exit`) |
| **Alt + E** | **Edit Menu** | Edit 메뉴 열기 (`Undo`, `Cut`, `Copy`, `Paste`) |
| **Alt + S** | **Search Menu** | Search 메뉴 열기 (`Find`, `Again`, `Go to Line`) |
| **Alt + R** | **Run Menu** | Run 메뉴 열기 (`Run`, `User Screen`) |
| **Alt + C** | **Compile Menu** | Compile 메뉴 열기 (`Compile`, `Make`) |
| **Alt + D** | **Debug Menu** | Debug 메뉴 열기 (`Cont`, `Step`, `Trace`, `BP`) |
| **Alt + O** | **Options Menu** | Options 메뉴 열기 (`Line Nums`, `Guides`, `Sound`) |
| **Alt + W** | **Window Menu** | Window 메뉴 열기 (`Tile`, `Cascade`, `Close`) |
| **Alt + H** | **Help Menu** | Help 메뉴 열기 (`About`, `FORTRAN 77`) |
| **F1** | Help / About | Turbo F77 정보 및 도움말 |
| **F2** | Save | 현재 버퍼 저장 / 다른 이름으로 저장 |
| **F3** | Open | 파일 브라우저 열기 |
| **F4** | **Breakpoint** | 현재 라인에 브레이크포인트(`●`) 설정/해제 |
| **F5** | **Debug / Continue** | 디버깅 시작 / 다음 브레이크포인트까지 실행 |
| **F7** | **Trace Into** | 한 줄씩 실행 (서브루틴 내부 진입) |
| **F8** | **Step Over** | 한 줄씩 실행 (함수/루프 건너뛰기) |
| **Ctrl + F2** | **Reset Debugger** | 디버깅 세션 리셋 및 종료 |
| **Ctrl + F9** | **Run** | 빌드, 실행 후 **User Screen**에 결과 표시 |
| **Alt + F9** | **Compile** | "Compiling..." 통계 모달과 함께 빌드 |
| **F9** | Make | 보조 모듈 파일들과 함께 전체 빌드 |
| **Alt + F5** | **User Screen** | 프로그램 실행 결과 전체화면 토글 |
| **Ctrl + F** | **Find** | 문자열 검색 다이얼로그 |
| **Ctrl + L** | **Search Again** | 다음 일치 항목 찾기 |
| **Alt + G** | **Go to Line** | 특정 라인 번호로 이동 (`Ctrl+G`) |
| **Alt + L** | **Line Numbers** | 라인 번호 표시 on/off (`F6`) |
| **F10** | Menu Bar | 상단 풀다운 메뉴바 포커스 |
| **Alt + X** | Exit | Turbo F77 종료 |
| **Shift + 방향키** | **Select Block** | 텍스트 블록 선택 |
| **Ctrl + C** / **Ctrl + Ins** | **Copy** | 클립보드로 복사 |
| **Ctrl + X** / **Shift + Del** | **Cut** | 클립보드로 잘라내기 |
| **Ctrl + V** / **Shift + Ins** | **Paste** | 커서 위치에 붙여넣기 |
| **Esc** | Close | 활성 모달 닫기 / 선택 영역 해제 |

---

## 빌드 및 실행

```bash
# 1. 빌드
go build -o bin/tf77 ./cmd/tf77

# 2. 실행
./bin/tf77

# 또는 특정 포트란 파일 열기
./bin/tf77 examples/fibonacci.for
```

---

## Open Watcom 2.0 FORTRAN 77 설정

Open Watcom 2.0이 설치되어 있는 경우 환경 변수를 지정해 두면 `tf77`가 자동으로 탐색하여 빌드를 수행합니다:

```bash
export WATCOM=/opt/watcom
export PATH=$WATCOM/binl64:$PATH
```

시스템에 Open Watcom 컴파일러가 없는 경우 `gfortran` 또는 대체 컴파일러가 감지되면 fallback으로 빌드되며, 컴파일러가 없을 경우 설치 안내 가이드 메시지가 표시됩니다.

---

## 예제 코드

* `main.for`: 누적 합 계산 및 시작 템플릿
* `examples/hello.for`: 기본 Hello World 포트란 77
* `examples/fibonacci.for`: 피보나치 수열 생성기
* `examples/stats.for`: 배열 통계 및 RMS 계산 예제
* `examples/modular/`: **다중 소스 파일 모듈화 예제**
  * `main.for`: 메인 실행 진입점 (`PROGRAM MODULAR`)
  * `math_sub.for`: 연산 함수 모듈 (`INTEGER FUNCTION CALCSUM`)
  * `io_sub.for`: 콘솔 출력 서브루틴 모듈 (`SUBROUTINE PRINTSUM`)
