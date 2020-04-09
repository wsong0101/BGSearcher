
# BGSearcher
### About
보드게임 통합 검색 서비스를 제공하는 사이트입니다.

### 설치 및 실행
구동을 위해 Google Cloud Platform의 Storage와 Firestore를 사용합니다.
해당 서비스에 접근할 수 있는 권한이 필요합니다.
[https://cloud.google.com/docs/authentication/production](https://cloud.google.com/docs/authentication/production)

```
git clone https://github.com/wsong0101/BGSearcher.git
cd ./BGSearcher
```

다음과 같은 내용의 config.yaml 파일을 생성합니다.
```
# Server Configurations
server:
	port: 3333
	passwd: ""
```

```
go build
export GOOGLE_APPLICATION_CREDENTIALS="[PATH]"
go run server.go
```
```
브라우저에서 127.0.0.1:3333 접속 확인
```
