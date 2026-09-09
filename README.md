# Turbo F77 (Version 1.0) 🚀

> **Retro Borland Turbo Pascal / Turbo C Look & Feel IDE for FORTRAN 77**

**Turbo F77 (`tf77`)**는 90년대 볼랜드(Borland)의 전설적인 **Turbo Pascal**과 **Turbo C** 특유의 비주얼 인터페이스(Turbo Vision 파란색 에디터 캔버스 `#0000A8`, 이중선 프레임 `╔═╗`, 상단 풀다운 메뉴바, 하단 핫키 바, `Alt+F5` User Screen)에 **Open Watcom 2.0 FORTRAN 77** 컴파일러 툴체인을 결합한 레트로 터미널 개발 환경(TUI IDE)입니다.

단일 Go 바이너리로 빌드되어 macOS, Linux, Windows 어디서나 가볍고 네이티브하게 실행됩니다.

---

## 📸 주요 특징

* **Classic Borland Turbo Vision UI**:
  * 시그니처 터보 블루 에디터 캔버스 (`#0000A8`) 및 이중선 박스 드로잉 (`╔═╗`, `║ ║`, `╚═╝`)
  * 입체 텍스트 그림자(Drop Shadow) 및 레트로 윈도우 헤더 (`[■] 1 NONAME00.FOR [▲]`)
  * 상단 풀다운 메뉴바 (`File`, `Edit`, `Search`, `Run`, `Compile`, `Options`, `Window`, `Help`)
  * 하단 핫키 바 (`F1 Help`, `F2 Save`, `F3 Open`, `Alt+F9 Compile`, `F9 Make`, `Ctrl+F9 Run`, `Alt+F5 User`, `F10 Menu`)

* **FORTRAN 77 고정 포맷(Fixed-Format) 완벽 지원**:
  * **1열 주석**: 1번째 열이 `C`, `c`, `*`, `!`인 경우 라인 전체 주석 처리
  * **1~5열 레이블**: 문장 레이블(Statement Label) 전용 스타일링
  * **6열 연속행 기호**: `Continuation` 문자 강조
  * **하단 상태바 컬럼 가이드**: 현재 커서 위치에 따라 `Col: 1 [Label]`, `Col: 6 [Cont]`, `Col: 7 [Stmt]`, `Col: 73+ [Ident]` 실시간 안내

* **FORTRAN 77 Syntax Highlighting**:
  * 대소문자 무구분 키워드(`PROGRAM`, `SUBROUTINE`, `DO`, `CONTINUE`, `IF`, `THEN`, `ELSE`, `ENDIF`, `READ`, `WRITE`, `PRINT` 등)
  * 내장 수학 함수(`SQRT`, `SIN`, `COS`, `ABS`, `MOD`, `EXP`, `LOG` 등)
  * 타입 선언(`INTEGER`, `REAL`, `DOUBLE PRECISION`, `COMPLEX`, `LOGICAL`, `CHARACTER`)
  * 리터럴 및 논리 연산자(`.TRUE.`, `.FALSE.`, `.AND.`, `.OR.`, `.EQ.`, `.NE.`, `.LT.`, `.GE.` 등)

* **Open Watcom 2.0 툴체인 연동 & Compiling 모달**:
  * `wfl386`, `wfc386`, `wfl`, `wfc` 자동 경로 탐색 (`WATCOM` 환경변수, 시스템 PATH, 표준 설치 디렉토리)
  * 컴파일 시 볼랜드 특유의 **"Compiling..." 팝업 박스**(대상 파일, 총 라인 수, 에러/워닝 수, 경과 시간)
  * Open Watcom 전용 진단 메시지(`file.for(line): Error! E1024: ...`, `*ERR*`, `*WRN*`) 정밀 파싱
  * 컴파일 에러 발생 시 **에러 목록 다이얼로그 표시 및 소스코드 해당 라인으로 즉시 점프**

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

* **볼랜드 레트로 사운드 이펙트 (Sound FX)**:
  * 컴파일 성공 시 경쾌한 2단 비프음, 빌드 실패 시 묵직한 에러 버즈음
  * 디버거 브레이크포인트 적중 시 핑 사운드
  * `F10` ➔ `Options` ➔ `Sound: ON / OFF` 토글 지원

---

## ⌨️ 단축키 안내

| 단축키 | 기능 | 설명 |
|---|---|---|
| **Alt + F** | **File Menu** | File 메뉴 열기 |
| **Alt + E** | **Edit Menu** | Edit 메뉴 열기 |
| **Alt + S** | **Search Menu** | Search 메뉴 열기 |
| **Alt + R** | **Run Menu** | Run 메뉴 열기 |
| **Alt + C** | **Compile Menu** | Compile 메뉴 열기 |
| **Alt + D** | **Debug Menu** | Debug 메뉴 열기 |
| **Alt + O** | **Options Menu** | Options 메뉴 열기 |
| **Alt + W** | **Window Menu** | Window 메뉴 열기 |
| **Alt + H** | **Help Menu** | Help 메뉴 열기 |
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
| **F9** | Make | 빌드 실행 |
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

## 🛠️ 빌드 및 실행

```bash
# 1. 빌드
go build -o bin/tf77 ./cmd/tf77

# 2. 실행
./bin/tf77

# 또는 특정 포트란 파일 열기
./bin/tf77 examples/fibonacci.for
```

---

## ⚙️ Open Watcom 2.0 FORTRAN 77 설정

Open Watcom 2.0이 설치되어 있는 경우 환경 변수를 지정해 두면 `tf77`가 자동으로 탐색하여 빌드를 수행합니다:

```bash
export WATCOM=/opt/watcom
export PATH=$WATCOM/binl64:$PATH
```

시스템에 Open Watcom 컴파일러가 없는 경우 `gfortran` 또는 대체 컴파일러가 감지되면 fallback으로 빌드되며, 컴파일러가 없을 경우 설치 안내 가이드 메시지가 표시됩니다.
