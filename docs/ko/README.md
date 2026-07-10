# gacha

터미널에서 투자 질문을 바로 물어보세요.

`gacha`는 투자 리서치를 위한 작은 터미널 앱입니다.
사용자가 쉬운 말로 질문합니다.
Gacha는 그 질문을 구조화된 AI 리서치 흐름으로 바꿉니다.

투자 결과를 완벽하게 예측할 수는 없습니다.
그래도 꼼꼼한 조사는 더 나은 확률을 만들 수 있습니다.
Gacha는 최신 데이터 확인, 대안 비교, 리스크 점검을 도와줍니다.
불명확한 질문을 더 깔끔한 판단으로 바꾸는 것이 목표입니다.

English: [../../README.md](../../README.md)

## 요약

Gacha를 설치하고, `gch`를 실행하고, 최초 1회 투자 프로필을 설정한 뒤
투자 질문을 입력하면 됩니다.

```bash
curl -fsSL https://raw.githubusercontent.com/dkstm95/gacha/main/install.sh | sh
"$HOME/.local/bin/gch"
```

Gacha는 자동 매매 도구가 아니라 리서치 도구입니다.
AI가 최신 데이터를 확인하고, 대안을 비교하고, 리스크를 드러내고,
나중에 다시 볼 수 있는 리포트를 만들도록 돕습니다.

## 어떤 판단을 도와주나요?

Gacha는 투자 판단의 선명도가 서로 다른 상황을 지원합니다.

- 탐색: 아직 무엇에 어떻게 투자해야 할지 모르는 경우
- 테마: 투자하고 싶은 분야는 있지만 구체적 대상은 모르는 경우
- 진입: 살 대상은 정했지만 언제 어떤 조건에서 살지 모르는 경우
- 보유: 보유 중이고 유지, 축소, 매도, 손절 기준이 필요한 경우
- 포트폴리오: 비중, 집중도, 중복 노출, 리스크를 보고 싶은 경우

목표는 투자를 확실하게 만드는 것이 아닙니다.
더 빠르고 차분하며 일관된 리서치를 돕는 것입니다.

## 무엇을 해주나요?

- 터미널 안에서 투자 질문을 묻고 답을 받습니다.
- AI가 의견을 내기 전에 최신 데이터를 확인하도록 지시합니다.
- 짧고 쉬운 기본 리포트부터 보여줍니다.
- 원하면 완성된 리포트를 저장합니다.
- AI 설정이 없어도 붙여넣기용 프롬프트를 만들어 줍니다.

Gacha가 직접 매매하지는 않습니다.
시장 데이터를 직접 가져오지도 않습니다.
리서치 흐름을 만들고, 연결된 AI 도구에 전달합니다.

## 빠른 시작

### AI 에이전트를 사용한다면

사용 중인 코딩 어시스턴트나 터미널 에이전트에게 설치를 맡길 수 있습니다.

```text
다음 README를 따라 Gacha를 설치해줘:
https://github.com/dkstm95/gacha
그다음 gch를 실행하고 첫 설정을 완료할 수 있게 도와줘.
```

### macOS와 Linux

설치:

```bash
curl -fsSL https://raw.githubusercontent.com/dkstm95/gacha/main/install.sh | sh
```

실행:

```bash
"$HOME/.local/bin/gch"
```

설치하면 두 명령어를 사용할 수 있습니다.

- `gch`: 평소에 쓰기 좋은 짧은 명령어
- `gacha`: 전체 명령어

Gacha를 쓰기 위해 개발 도구를 따로 설치할 필요는 없습니다.

첫 실행 때 Gacha가 투자 프로필을 설정하기 위한 몇 가지 질문을 합니다.
건너뛸 수 있고, 나중에 `/profile`에서 다시 바꿀 수 있습니다.

AI 설정을 물어볼 수도 있습니다.
안내에 따라 사용할 AI 계정을 연결하면 됩니다.

설치 프로그램이 `$HOME/.local/bin`이 `PATH`에 없다고 안내하면, 같은
터미널에서 `gch`를 쓰기 전에 출력된 `export PATH=...` 명령을 실행하세요.
새 터미널에서도 짧은 명령어를 계속 쓰려면 그 줄을 셸 프로필(zsh는
`~/.zprofile`, bash는 `~/.bash_profile`)에 추가한 뒤 새 터미널을 여세요.

### Windows

최신 Windows zip을 내려받으세요.

```text
https://github.com/dkstm95/gacha/releases/latest
```

압축을 풉니다.
`gacha.exe`를 명령어용 폴더로 옮깁니다.
그 뒤 새 터미널 창을 엽니다.

```powershell
gacha setup
```

짧은 `gch` 명령도 원한다면 파일을 하나 더 만드세요.
같은 폴더에서 `gacha.exe`를 `gch.exe`로 복사하면 됩니다.

```powershell
Copy-Item gacha.exe gch.exe
```

Windows에서는 아직 AI 자동 설정을 지원하지 않습니다.
Gacha가 안내하는 추가 AI 도구를 먼저 설치하세요.
그 뒤 setup을 실행하세요.

```powershell
gacha setup
```

## 첫 질문하기

앱을 시작합니다.

```bash
gch
```

간단한 프롬프트 중심 화면이 열립니다.
첫 실행이라면 먼저 기본 리서치 선호도를 묻습니다.
관심 시장, 투자 기간, 리스크 성향, 리포트 스타일, 자주 하는 판단을 설정합니다.

질문을 입력하세요.

```text
NVDA 지금 사도 될까?
```

앱을 열지 않고 한 번만 물어볼 수도 있습니다.

```bash
gch "NVDA 지금 사도 될까?"
```

모델을 직접 고를 필요는 없습니다. Gacha는 로컬 AI 런타임을 통해 모델
라우팅을 처리하고, 필요하면 런타임 기본값으로 돌아갑니다.

## 질문 예시

```text
6개월에서 12개월 관점에서 무엇에 투자하면 좋을까?
AI 인프라에 투자하고 싶은데 어떤 종목이나 ETF를 비교해야 할까?
반도체에 투자하고 싶은데 무엇을 비교해야 할까?
TSLA를 보유 중인데 언제 줄이거나 팔아야 할까?
내 포트폴리오를 점검해줘: AAPL 35%, NVDA 30%, SGOV 35%
```

좋은 질문에는 다음을 함께 쓰면 좋습니다.

- 목표
- 투자 기간
- 감당 가능한 위험
- 현재 보유 종목

## 설정이 안 되어 있다면

준비 상태를 확인합니다.

```bash
gch doctor
```

이 명령은 다음을 보여줍니다.

- AI 설정 준비 여부
- AI 계정 연결 여부
- 리포트 저장 위치

설정을 고칩니다.

```bash
gch setup
```

macOS와 Linux에서는 `gch setup`이 AI 설정을 도와줍니다.
그 뒤 계정 로그인을 시작합니다.

Windows에서는 필요한 AI 도구를 먼저 별도로 설치해야 합니다.

그래도 AI 설정을 실행할 수 없으면 Gacha가 프롬프트를 출력합니다.
웹 AI에 그대로 붙여넣을 수 있는 프롬프트입니다.

## 앱 안의 명령어

앱 안에서 사용할 수 있습니다.

```text
/profile
/settings
/theme
/help
/quit
```

`/profile`은 투자 프로필을 보거나 수정합니다.
`/settings`는 언어와 테마 설정 화면을 엽니다.
`/theme`은 테마 선택 화면으로 바로 이동합니다.
방향키와 enter로 고르거나 전체 명령을 직접 입력할 수 있습니다.

```text
/language auto
/language en
/language ko
/theme system
/theme dark
/theme light
/theme gacha
```

설정, 진단, 업데이트는 앱 밖에서 실행합니다.

```bash
gch profile
gch setup
gch doctor
gch update
```

## 개인정보와 데이터 흐름

Gacha는 투자 프로필과 저장한 리포트를 사용자의 컴퓨터에 저장합니다.
리포트는 사용자가 저장을 선택한 경우에만 파일로 남습니다.

질문을 입력하면 Gacha는 리서치 프롬프트를 연결된 OpenCode 런타임과
AI provider로 보냅니다.
해당 요청에는 사용자가 연결한 AI provider의 데이터 정책이 적용됩니다.

현재 버전의 Gacha에는 제품 분석이나 telemetry 코드가 없습니다.
`gch update`는 사용자가 업데이트 명령을 실행할 때만 GitHub에 접속합니다.

## 리포트 저장

AI가 리포트를 완성하면 Gacha가 저장 여부를 물어봅니다.
저장하면 파일로 남습니다.

기본 위치:

```text
~/.local/share/gacha/reports
```

사용자가 저장에 동의한 경우에만 저장합니다.
복사/붙여넣기용 프롬프트는 리포트로 저장하지 않습니다.

## 언어

Gacha는 터미널 언어에 맞추려고 합니다.
질문에 한국어가 있으면 AI에게 한국어 답변을 요청합니다.

앱 안에서 언어를 바꾸려면:

```text
/settings
```

그다음 `Language`를 고르거나 `/language ko`, `/language en`,
`/language auto`를 직접 입력하세요.

## 업데이트

macOS와 Linux:

```bash
gch update
```

현재 컴퓨터에 맞는 앱 파일을 내려받습니다.
그 뒤 기존 파일을 교체합니다.

릴리스 페이지에서 최신 Windows zip을 내려받으세요.
기존 `gacha.exe`를 교체하세요.
그 뒤 새 터미널을 여세요.

## 제거

macOS와 Linux:

```bash
rm -f ~/.local/bin/gacha ~/.local/bin/gch
```

설치할 때 `GACHA_INSTALL_DIR`를 따로 지정했다면, 그 폴더에서 `gacha`와
`gch`를 제거하세요.

Windows:

```powershell
Remove-Item .\gacha.exe
Remove-Item .\gch.exe
```

실행 파일을 옮겨 둔 폴더에서 실행하세요.

Gacha의 로컬 투자 프로필과 저장 리포트까지 지우려면:

```bash
rm -rf ~/.config/gacha ~/.local/share/gacha
```

이 명령은 OpenCode나 AI provider 인증 정보는 제거하지 않습니다.
다른 곳에서도 쓰지 않는 경우에만 별도로 제거하세요.

## 최신 데이터

투자 정보는 빠르게 바뀝니다.
Gacha는 AI에게 현재 웹/시장 데이터를 확인하라고 지시합니다.
사용자가 "최신"이라고 쓰지 않아도 그렇게 합니다.

현재 데이터를 확인할 수 없으면 추천을 내리지 않아야 합니다.

좋은 Gacha 리포트는 다음을 분명히 보여줘야 합니다.

- 어떤 데이터를 언제, 어디서 확인했는지
- 쉬운 결론
- 다음 행동과 재검토 시점
- 가장 큰 리스크와 반대 의견
- 매수, 보유, 매도, 관망 조건

## 한계

Gacha는 다음을 하지 않습니다.

- 자동 매매
- 수익 보장
- 전문 금융, 세무, 법률 자문 대체
- 현재 버전에서 직접 시장 데이터 조회

Gacha는 엄격한 리서치 흐름을 만듭니다.
최신 웹/시장 데이터 조사는 연결된 AI 도구가 수행합니다.

## 개발자 문서

개발 문서: [../development.md](../development.md)
